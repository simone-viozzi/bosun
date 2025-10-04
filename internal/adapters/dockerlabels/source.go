package dockerlabels

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
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

// Snapshot implements the LabelSource interface
func (d *DockerLabelSource) Snapshot(ctx context.Context, sel ports.Selector) (dlabels.Snapshot, error) {
	entities, err := d.snapshotContainers(ctx, sel)
	if err != nil {
		return dlabels.Snapshot{}, err
	}
	return dlabels.Snapshot{
		Entities: entities,
		TakenAt:  time.Now(),
	}, nil
}
