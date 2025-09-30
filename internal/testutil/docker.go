//go:build integration
// +build integration

package testutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func mustDocker(t *testing.T) *client.Client {
	t.Helper()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("docker client: %v", err)
	}
	return cli
}

// HostPort returns the published host port mapped to the given container port (e.g., 80/tcp) for a service in a compose project.
func HostPort(t *testing.T, ctx context.Context, project, service string, containerPort uint16) int {
	t.Helper()
	cli := mustDocker(t)

	args := filters.NewArgs(
		filters.KeyValuePair{Key: "label", Value: "com.docker.compose.project=" + project},
		filters.KeyValuePair{Key: "label", Value: "com.docker.compose.service=" + service},
	)
	ctrs, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil || len(ctrs) == 0 {
		t.Logf("found %d containers for project %s service %s", len(ctrs), project, service)
		for _, c := range ctrs {
			t.Logf("container: %s, labels: %v", c.Names, c.Labels)
		}
		t.Fatalf("list containers for %s/%s: %v", project, service, err)
	}

	ins, err := cli.ContainerInspect(ctx, ctrs[0].ID)
	if err != nil {
		t.Fatalf("inspect container: %v", err)
	}
	bindings := ins.NetworkSettings.Ports[nat.Port(fmt.Sprintf("%d/tcp", containerPort))]
	if len(bindings) == 0 {
		t.Fatalf("no port bindings for %d/tcp on %s", containerPort, service)
	}
	return atoiOrFail(t, bindings[0].HostPort)
}

func DumpLogs(t *testing.T, ctx context.Context, project, outDir string) {
	t.Helper()
	cli := mustDocker(t)

	args := filters.NewArgs(
		filters.KeyValuePair{Key: "label", Value: "com.docker.compose.project=" + project},
	)
	ctrs, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		t.Logf("dump logs: list error: %v", err)
		return
	}

	err = os.MkdirAll(outDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create output directory %s: %v", outDir, err)
		return
	}

	for _, c := range ctrs {
		rc, err := cli.ContainerLogs(ctx, c.ID, container.LogsOptions{
			ShowStdout: true, ShowStderr: true, Timestamps: true, Tail: "500",
		})
		if err != nil {
			t.Logf("logs for %s: %v", c.Names, err)
			continue
		}
		name := "container_" + strings.TrimLeft(c.Names[0], "/") + ".log"
		fp := filepath.Join(outDir, name)
		f, err := os.Create(fp)
		if err != nil {
			t.Logf("create log file %s: %v", fp, err)
			continue
		}
		_, err = io.Copy(f, rc)
		if err != nil {
			t.Logf("copy logs to %s: %v", fp, err)
		}
		_ = f.Close()
		_ = rc.Close()
	}
}

// atoiOrFail converts the string s to an integer using strconv.Atoi.
// If the conversion fails, it calls t.Fatalf to fail the test.
// It marks itself as a test helper using t.Helper().
func atoiOrFail(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("atoi: %v", err)
	}
	return n
}
