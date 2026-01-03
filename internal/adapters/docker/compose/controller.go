package compose

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// Controller implements ports.ComposeController using Docker API.
type Controller struct {
	docker *client.Client
}

// NewController creates a new Compose controller.
func NewController(docker *client.Client) *Controller {
	return &Controller{
		docker: docker,
	}
}

// ListStackContainers returns all containers in a Compose stack.
func (c *Controller) ListStackContainers(ctx context.Context, projectName string) ([]ports.StackContainer, error) {
	// Filter by com.docker.compose.project label
	filterArgs := filters.NewArgs(
		filters.Arg("label", "com.docker.compose.project="+projectName),
	)

	containers, err := c.docker.ContainerList(ctx, container.ListOptions{
		All:     true, // Include stopped containers
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for stack %s: %w", projectName, err)
	}

	result := make([]ports.StackContainer, 0, len(containers))
	for _, cnt := range containers {
		// Extract service name
		serviceName := cnt.Labels["com.docker.compose.service"]
		if serviceName == "" {
			// Skip containers without service label
			continue
		}

		// Parse depends_on label (comma-separated service names)
		dependsOn := parseDependsOn(cnt.Labels["com.docker.compose.depends_on"])

		// Container name without leading /
		name := strings.TrimPrefix(cnt.Names[0], "/")

		result = append(result, ports.StackContainer{
			ID:          cnt.ID,
			Name:        name,
			ServiceName: serviceName,
			State:       cnt.State,
			DependsOn:   dependsOn,
			Labels:      cnt.Labels,
		})
	}

	return result, nil
}

// IsStackRunning returns true if all containers in the stack are running.
func (c *Controller) IsStackRunning(ctx context.Context, projectName string) (bool, error) {
	containers, err := c.ListStackContainers(ctx, projectName)
	if err != nil {
		return false, err
	}

	if len(containers) == 0 {
		return false, nil // Stack not found or empty
	}

	for _, cnt := range containers {
		if cnt.State != "running" {
			return false, nil
		}
	}

	return true, nil
}

// StopStack stops all containers in reverse dependency order.
func (c *Controller) StopStack(ctx context.Context, projectName string, opts ports.StopOptions) error {
	containers, err := c.ListStackContainers(ctx, projectName)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return &jobs.StopError{
			StackName:     projectName,
			ContainerName: "",
			ContainerID:   "",
			Cause:         jobs.ErrStackNotFound,
		}
	}

	// Sort topologically, then reverse for stop order (dependents first)
	sorted, err := topologicalSort(containers)
	if err != nil {
		return fmt.Errorf("failed to determine container order for stack %s: %w", projectName, err)
	}
	reversed := reverse(sorted)

	// Stop each container
	timeout := int(opts.Timeout.Seconds())
	for _, cnt := range reversed {
		if cnt.State != "running" {
			// Already stopped, skip
			continue
		}

		err := c.docker.ContainerStop(ctx, cnt.ID, container.StopOptions{
			Timeout: &timeout,
		})
		if err != nil {
			return &jobs.StopError{
				StackName:     projectName,
				ContainerName: cnt.Name,
				ContainerID:   cnt.ID,
				Cause:         err,
			}
		}
	}

	return nil
}

// StartStack starts all containers in dependency order.
func (c *Controller) StartStack(ctx context.Context, projectName string, opts ports.StartOptions) error {
	containers, err := c.ListStackContainers(ctx, projectName)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return &jobs.StartError{
			StackName:     projectName,
			ContainerName: "",
			ContainerID:   "",
			Cause:         jobs.ErrStackNotFound,
		}
	}

	// Sort topologically for start order (dependencies first)
	sorted, err := topologicalSort(containers)
	if err != nil {
		return fmt.Errorf("failed to determine container order for stack %s: %w", projectName, err)
	}

	// Start each container
	for _, cnt := range sorted {
		if cnt.State == "running" {
			// Already running, skip
			continue
		}

		err := c.docker.ContainerStart(ctx, cnt.ID, container.StartOptions{})
		if err != nil {
			return &jobs.StartError{
				StackName:     projectName,
				ContainerName: cnt.Name,
				ContainerID:   cnt.ID,
				Cause:         err,
			}
		}
	}

	return nil
}

// parseDependsOn parses the com.docker.compose.depends_on label.
// Format: "service1,service2,service3" or empty string.
func parseDependsOn(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
