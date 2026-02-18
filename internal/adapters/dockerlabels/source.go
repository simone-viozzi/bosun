package dockerlabels

import (
	"context"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"

	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// dockerClient defines the subset of Docker client methods we use
type dockerClient interface {
	ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error)
	VolumeList(ctx context.Context, opts volume.ListOptions) (volume.ListResponse, error)
	NetworkList(ctx context.Context, opts network.ListOptions) ([]network.Summary, error)
}

type DockerLabelSource struct {
	CLI dockerClient
}

func NewFromEnv() (*DockerLabelSource, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerLabelSource{CLI: cli}, nil
}

// NewFromClient creates a DockerLabelSource using an existing Docker client.
// This allows sharing a single Docker client instance across the application.
func NewFromClient(cli *client.Client) *DockerLabelSource {
	return &DockerLabelSource{CLI: cli}
}

// buildLabelFilters constructs Docker API filters for ProjectFilter and StackFilter.
// Multiple values in each filter are OR'd (match any), but ProjectFilter and StackFilter
// together are AND'd (must match at least one from each if both specified).
func buildLabelFilters(sel ports.Selector) filters.Args {
	args := filters.NewArgs()

	// Filter by Docker Compose project label
	for _, project := range sel.ProjectFilter {
		args.Add("label", "com.docker.compose.project="+project)
	}

	// Filter by bosun.stack label
	for _, stack := range sel.StackFilter {
		args.Add("label", dlabels.LabelStack+"="+stack)
	}

	return args
}

// snapshotContainers collects containers from Docker, filters by label prefixes,
// and returns labeled entities for containers with matching labels.
// Uses Docker's native label filtering for ProjectFilter and StackFilter for efficiency.
func (s *DockerLabelSource) snapshotContainers(ctx context.Context, sel ports.Selector) ([]dlabels.LabeledEntity, error) {
	opts := container.ListOptions{
		All:     sel.IncludeStopped,
		Filters: buildLabelFilters(sel),
	}
	ctrs, err := s.CLI.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	out := make([]dlabels.LabeledEntity, 0, len(ctrs))
	for _, c := range ctrs {
		fl := FilterByPrefixes(c.Labels, sel.Prefixes)
		if len(fl) == 0 {
			continue
		}
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		ent := dlabels.LabeledEntity{
			Kind:   dlabels.KindContainer,
			ID:     c.ID,
			Name:   name,
			Labels: fl,
			Meta: map[string]string{
				"compose.project": c.Labels["com.docker.compose.project"],
				"compose.service": c.Labels["com.docker.compose.service"],
				"image":           c.Image,
			},
		}
		if instance := c.Labels[dlabels.LabelInstance]; instance != "" {
			ent.Meta["instance"] = instance
		}
		out = append(out, ent)
	}
	return out, nil
}

// snapshotVolumes collects volumes from Docker, filters by label prefixes,
// and returns labeled entities for volumes with matching labels.
// Applies ProjectFilter and StackFilter if specified.
func (s *DockerLabelSource) snapshotVolumes(ctx context.Context, sel ports.Selector) ([]dlabels.LabeledEntity, error) {
	opts := volume.ListOptions{
		Filters: buildLabelFilters(sel),
	}
	vl, err := s.CLI.VolumeList(ctx, opts)
	if err != nil {
		return nil, err
	}

	out := make([]dlabels.LabeledEntity, 0, len(vl.Volumes))
	for _, v := range vl.Volumes {
		fl := FilterByPrefixes(v.Labels, sel.Prefixes)
		if len(fl) == 0 {
			continue
		}
		ent := dlabels.LabeledEntity{
			Kind:   dlabels.KindVolume,
			ID:     v.Name,
			Name:   v.Name,
			Labels: fl,
			Meta: map[string]string{
				"driver": v.Driver,
			},
		}
		if instance := v.Labels[dlabels.LabelInstance]; instance != "" {
			ent.Meta["instance"] = instance
		}
		out = append(out, ent)
	}
	return out, nil
}

// snapshotNetworks collects networks from Docker, filters by label prefixes,
// and returns labeled entities for networks with matching labels.
// Applies ProjectFilter and StackFilter if specified.
func (s *DockerLabelSource) snapshotNetworks(ctx context.Context, sel ports.Selector) ([]dlabels.LabeledEntity, error) {
	opts := network.ListOptions{
		Filters: buildLabelFilters(sel),
	}
	nets, err := s.CLI.NetworkList(ctx, opts)
	if err != nil {
		return nil, err
	}

	out := make([]dlabels.LabeledEntity, 0, len(nets))
	for _, n := range nets {
		fl := FilterByPrefixes(n.Labels, sel.Prefixes)
		if len(fl) == 0 {
			continue
		}
		ent := dlabels.LabeledEntity{
			Kind:   dlabels.KindNetwork,
			ID:     n.ID,
			Name:   n.Name,
			Labels: fl,
			Meta: map[string]string{
				"driver": n.Driver,
				"scope":  n.Scope,
			},
		}
		if instance := n.Labels[dlabels.LabelInstance]; instance != "" {
			ent.Meta["instance"] = instance
		}
		out = append(out, ent)
	}
	return out, nil
}

// Snapshot implements the LabelSource interface
func (s *DockerLabelSource) Snapshot(ctx context.Context, sel ports.Selector) (dlabels.Snapshot, error) {
	g, ctx := errgroup.WithContext(ctx)

	var containers, volumes, networks []dlabels.LabeledEntity

	g.Go(func() error {
		var err error
		containers, err = s.snapshotContainers(ctx, sel)
		return err
	})

	g.Go(func() error {
		var err error
		volumes, err = s.snapshotVolumes(ctx, sel)
		return err
	})

	g.Go(func() error {
		var err error
		networks, err = s.snapshotNetworks(ctx, sel)
		return err
	})

	if err := g.Wait(); err != nil {
		return dlabels.Snapshot{}, err
	}

	entities := slices.Concat(containers, volumes, networks)

	// Sort entities by Kind (container < volume < network), then by Name
	kindOrder := map[dlabels.Kind]int{
		dlabels.KindContainer: 0,
		dlabels.KindVolume:    1,
		dlabels.KindNetwork:   2,
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].Kind != entities[j].Kind {
			return kindOrder[entities[i].Kind] < kindOrder[entities[j].Kind]
		}
		return entities[i].Name < entities[j].Name
	})

	return dlabels.Snapshot{
		Entities: entities,
		TakenAt:  time.Now(),
	}, nil
}
