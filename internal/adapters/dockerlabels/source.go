package dockerlabels

import (
	"context"
	"time"

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

// Snapshot implements the LabelSource interface
func (d *DockerLabelSource) Snapshot(ctx context.Context, sel ports.Selector) (dlabels.Snapshot, error) {
	// TODO: Implement full snapshot logic
	//    For now, return empty snapshot
	return dlabels.Snapshot{
		Entities: []dlabels.LabeledEntity{},
		TakenAt:  time.Now(),
	}, nil
}
