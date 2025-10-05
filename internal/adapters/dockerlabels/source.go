package dockerlabels

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
	"golang.org/x/sync/errgroup"
)

type DockerLabelSource struct {
	CLI *client.Client
}

func NewFromEnv() (*DockerLabelSource, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerLabelSource{CLI: cli}, nil
}

// snapshotContainers collects containers from Docker, filters by label prefixes,
// and returns labeled entities for containers with matching labels.
func (s *DockerLabelSource) snapshotContainers(ctx context.Context, sel ports.Selector) ([]dlabels.LabeledEntity, error) {
	opts := container.ListOptions{All: sel.IncludeStopped}
	ctrs, err := s.CLI.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	var out []dlabels.LabeledEntity
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
		out = append(out, ent)
	}
	return out, nil
}

// snapshotVolumes collects volumes from Docker, filters by label prefixes,
// and returns labeled entities for volumes with matching labels.
func (s *DockerLabelSource) snapshotVolumes(ctx context.Context, sel ports.Selector) ([]dlabels.LabeledEntity, error) {
	vl, err := s.CLI.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	var out []dlabels.LabeledEntity
	for _, v := range vl.Volumes {
		fl := FilterByPrefixes(v.Labels, sel.Prefixes)
		if len(fl) == 0 {
			continue
		}
		out = append(out, dlabels.LabeledEntity{
			Kind:   dlabels.KindVolume,
			ID:     v.Name,
			Name:   v.Name,
			Labels: fl,
			Meta:   map[string]string{},
		})
	}
	return out, nil
}

// snapshotNetworks collects networks from Docker, filters by label prefixes,
// and returns labeled entities for networks with matching labels.
func (s *DockerLabelSource) snapshotNetworks(ctx context.Context, sel ports.Selector) ([]dlabels.LabeledEntity, error) {
	nets, err := s.CLI.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	var out []dlabels.LabeledEntity
	for _, n := range nets {
		fl := FilterByPrefixes(n.Labels, sel.Prefixes)
		if len(fl) == 0 {
			continue
		}
		out = append(out, dlabels.LabeledEntity{
			Kind:   dlabels.KindNetwork,
			ID:     n.ID,
			Name:   n.Name,
			Labels: fl,
			Meta:   map[string]string{},
		})
	}
	return out, nil
}

// Snapshot implements the LabelSource interface
func (d *DockerLabelSource) Snapshot(ctx context.Context, sel ports.Selector) (dlabels.Snapshot, error) {
	g, ctx := errgroup.WithContext(ctx)

	var containers, volumes, networks []dlabels.LabeledEntity

	g.Go(func() error {
		var err error
		containers, err = d.snapshotContainers(ctx, sel)
		return err
	})

	g.Go(func() error {
		var err error
		volumes, err = d.snapshotVolumes(ctx, sel)
		return err
	})

	g.Go(func() error {
		var err error
		networks, err = d.snapshotNetworks(ctx, sel)
		return err
	})

	if err := g.Wait(); err != nil {
		return dlabels.Snapshot{}, err
	}

	entities := slices.Concat(containers, volumes, networks)

	return dlabels.Snapshot{
		Entities: entities,
		TakenAt:  time.Now(),
	}, nil
}
