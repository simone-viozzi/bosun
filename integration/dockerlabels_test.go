//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
	"github.com/simone-viozzi/bosun/internal/testutil"
)

// Test_Integration_DockerLabels_VolumeAndNetworkDiscovery validates that volumes and networks
// with matching label prefixes are discovered and included in snapshots.
func Test_Integration_DockerLabels_VolumeAndNetworkDiscovery(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start a compose stack with labeled volumes and networks
	stack := testutil.StartCompose(t, ctx, "dockerlabels-compose.yaml")

	// Create a network with bosun. labels manually (Docker Compose doesn't propagate network labels)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("failed to create Docker client: %v", err)
	}

	networkName := stack.Project + "-bosun-test-net"
	netResp, err := cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Labels: map[string]string{
			"bosun.test": "true",
			"bosun.type": "integration-test",
		},
	})
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	t.Cleanup(func() {
		_ = cli.NetworkRemove(context.Background(), netResp.ID)
	})

	// Create a DockerLabelSource
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		t.Fatalf("failed to create DockerLabelSource: %v", err)
	}

	// Take a snapshot with bosun. prefix filter
	sel := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: false,
	}

	snapshot, err := source.Snapshot(ctx, sel)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Validate snapshot contains entities
	if len(snapshot.Entities) == 0 {
		t.Fatal("Expected at least one entity in snapshot")
	}

	// Count entities by kind
	var containers, volumes, networks int
	for _, entity := range snapshot.Entities {
		switch entity.Kind {
		case dlabels.KindContainer:
			containers++
		case dlabels.KindVolume:
			volumes++
		case dlabels.KindNetwork:
			networks++
		}
	}

	t.Logf("Snapshot contains: %d containers, %d volumes, %d networks", containers, volumes, networks)

	// Verify we have at least one container (from the compose stack)
	if containers == 0 {
		t.Error("Expected at least one container with bosun. labels")
	}

	// Verify we have at least one volume
	if volumes == 0 {
		t.Error("Expected at least one volume with bosun. labels")
	}

	// Verify we have at least one network
	if networks == 0 {
		t.Error("Expected at least one network with bosun. labels")
	}

	// Verify entities have non-empty labels
	for _, entity := range snapshot.Entities {
		if len(entity.Labels) == 0 {
			t.Errorf("Entity %s (%s) has no labels but should have been filtered", entity.Name, entity.Kind)
		}
		// Verify all labels match the prefix
		for key := range entity.Labels {
			if len(key) < 6 || key[:6] != "bosun." {
				t.Errorf("Entity %s (%s) has label %s that doesn't start with bosun.", entity.Name, entity.Kind, key)
			}
		}
	}

	// Verify volumes have ID == Name
	for _, entity := range snapshot.Entities {
		if entity.Kind == dlabels.KindVolume && entity.ID != entity.Name {
			t.Errorf("Volume %s has ID=%s but expected ID to equal Name", entity.Name, entity.ID)
		}
	}

	// Verify Meta enrichment
	for _, entity := range snapshot.Entities {
		switch entity.Kind {
		case dlabels.KindContainer:
			// Containers should have image in Meta
			if _, hasImage := entity.Meta["image"]; !hasImage {
				t.Errorf("Container %s missing 'image' in Meta", entity.Name)
			}
		case dlabels.KindVolume:
			// Volumes should have driver in Meta
			if _, hasDriver := entity.Meta["driver"]; !hasDriver {
				t.Errorf("Volume %s missing 'driver' in Meta", entity.Name)
			}
		case dlabels.KindNetwork:
			// Networks should have driver and scope in Meta
			if _, hasDriver := entity.Meta["driver"]; !hasDriver {
				t.Errorf("Network %s missing 'driver' in Meta", entity.Name)
			}
			if _, hasScope := entity.Meta["scope"]; !hasScope {
				t.Errorf("Network %s missing 'scope' in Meta", entity.Name)
			}
		}
	}

	t.Logf("Integration test completed successfully with project: %s", stack.Project)
}
