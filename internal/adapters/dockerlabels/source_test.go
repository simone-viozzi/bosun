package dockerlabels

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
	"sort"
)

// mockDockerClient is a minimal mock that doesn't actually connect to Docker
// These tests validate the Meta enrichment logic without requiring Docker
type mockDockerClient struct{}

func (m *mockDockerClient) ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error) {
	return []container.Summary{
		{
			ID:    "container1",
			Names: []string{"/test-container"},
			Image: "test:latest",
			Labels: map[string]string{
				"bosun.test":                 "true",
				dlabels.LabelInstance:        "prod-01",
				"com.docker.compose.project": "myproject",
				"com.docker.compose.service": "web",
			},
		},
		{
			ID:    "container2",
			Names: []string{"/test-container-no-instance"},
			Image: "test2:latest",
			Labels: map[string]string{
				"bosun.test":                 "true",
				"com.docker.compose.project": "myproject",
				"com.docker.compose.service": "db",
			},
		},
	}, nil
}

func (m *mockDockerClient) VolumeList(ctx context.Context, opts volume.ListOptions) (volume.ListResponse, error) {
	return volume.ListResponse{
		Volumes: []*volume.Volume{
			{
				Name:   "test-volume",
				Driver: "local",
				Labels: map[string]string{
					"bosun.test":          "true",
					dlabels.LabelInstance: "prod-01",
				},
			},
			{
				Name:   "test-volume-no-instance",
				Driver: "nfs",
				Labels: map[string]string{
					"bosun.test": "true",
				},
			},
		},
	}, nil
}

func (m *mockDockerClient) NetworkList(ctx context.Context, opts network.ListOptions) ([]network.Summary, error) {
	return []network.Summary{
		{
			ID:     "net1",
			Name:   "test-network",
			Driver: "bridge",
			Scope:  "local",
			Labels: map[string]string{
				"bosun.test":          "true",
				dlabels.LabelInstance: "prod-01",
			},
		},
		{
			ID:     "net2",
			Name:   "test-network-no-instance",
			Driver: "overlay",
			Scope:  "swarm",
			Labels: map[string]string{
				"bosun.test": "true",
			},
		},
	}, nil
}

func TestSnapshotContainers_MetaEnrichment(t *testing.T) {
	source := &DockerLabelSource{CLI: &mockDockerClient{}}
	sel := ports.Selector{
		Prefixes:       []string{"bosun."},
		IncludeStopped: false,
	}

	entities, err := source.snapshotContainers(context.Background(), sel)
	if err != nil {
		t.Fatalf("snapshotContainers failed: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 containers, got %d", len(entities))
	}

	// Validate first container with instance
	c1 := entities[0]
	if c1.Meta["compose.project"] != "myproject" {
		t.Errorf("Expected compose.project=myproject, got %s", c1.Meta["compose.project"])
	}
	if c1.Meta["compose.service"] != "web" {
		t.Errorf("Expected compose.service=web, got %s", c1.Meta["compose.service"])
	}
	if c1.Meta["image"] != "test:latest" {
		t.Errorf("Expected image=test:latest, got %s", c1.Meta["image"])
	}
	if c1.Meta["instance"] != "prod-01" {
		t.Errorf("Expected instance=prod-01, got %s", c1.Meta["instance"])
	}

	// Validate second container without instance
	c2 := entities[1]
	if _, hasInstance := c2.Meta["instance"]; hasInstance {
		t.Errorf("Expected no instance field for container2, but got %s", c2.Meta["instance"])
	}
}

func TestSnapshotVolumes_MetaEnrichment(t *testing.T) {
	source := &DockerLabelSource{CLI: &mockDockerClient{}}
	sel := ports.Selector{
		Prefixes:       []string{"bosun."},
		IncludeStopped: false,
	}

	entities, err := source.snapshotVolumes(context.Background(), sel)
	if err != nil {
		t.Fatalf("snapshotVolumes failed: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 volumes, got %d", len(entities))
	}

	// Validate first volume with instance
	v1 := entities[0]
	if v1.Kind != dlabels.KindVolume {
		t.Errorf("Expected KindVolume, got %s", v1.Kind)
	}
	if v1.Meta["driver"] != "local" {
		t.Errorf("Expected driver=local, got %s", v1.Meta["driver"])
	}
	if v1.Meta["instance"] != "prod-01" {
		t.Errorf("Expected instance=prod-01, got %s", v1.Meta["instance"])
	}

	// Validate second volume without instance
	v2 := entities[1]
	if v2.Meta["driver"] != "nfs" {
		t.Errorf("Expected driver=nfs, got %s", v2.Meta["driver"])
	}
	if _, hasInstance := v2.Meta["instance"]; hasInstance {
		t.Errorf("Expected no instance field for volume2, but got %s", v2.Meta["instance"])
	}
}

func TestSnapshotNetworks_MetaEnrichment(t *testing.T) {
	source := &DockerLabelSource{CLI: &mockDockerClient{}}
	sel := ports.Selector{
		Prefixes:       []string{"bosun."},
		IncludeStopped: false,
	}

	entities, err := source.snapshotNetworks(context.Background(), sel)
	if err != nil {
		t.Fatalf("snapshotNetworks failed: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 networks, got %d", len(entities))
	}

	// Validate first network with instance
	n1 := entities[0]
	if n1.Kind != dlabels.KindNetwork {
		t.Errorf("Expected KindNetwork, got %s", n1.Kind)
	}
	if n1.Meta["driver"] != "bridge" {
		t.Errorf("Expected driver=bridge, got %s", n1.Meta["driver"])
	}
	if n1.Meta["scope"] != "local" {
		t.Errorf("Expected scope=local, got %s", n1.Meta["scope"])
	}
	if n1.Meta["instance"] != "prod-01" {
		t.Errorf("Expected instance=prod-01, got %s", n1.Meta["instance"])
	}

	// Validate second network without instance
	n2 := entities[1]
	if n2.Meta["driver"] != "overlay" {
		t.Errorf("Expected driver=overlay, got %s", n2.Meta["driver"])
	}
	if n2.Meta["scope"] != "swarm" {
		t.Errorf("Expected scope=swarm, got %s", n2.Meta["scope"])
	}
	if _, hasInstance := n2.Meta["instance"]; hasInstance {
		t.Errorf("Expected no instance field for network2, but got %s", n2.Meta["instance"])
	}
}

func TestEntitySorting(t *testing.T) {
	entities := []dlabels.LabeledEntity{
		{Kind: dlabels.KindNetwork, Name: "net-b"},
		{Kind: dlabels.KindContainer, Name: "ctr-b"},
		{Kind: dlabels.KindVolume, Name: "vol-a"},
		{Kind: dlabels.KindContainer, Name: "ctr-a"},
	}

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

	want := []string{"ctr-a", "ctr-b", "vol-a", "net-b"}
	for i, name := range want {
		if entities[i].Name != name {
			t.Errorf("entities[%d].Name = %s, want %s", i, entities[i].Name, name)
		}
	}
}
