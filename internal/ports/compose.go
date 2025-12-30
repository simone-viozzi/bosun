// Package ports defines the ComposeController interface.
// This file is a contract definition for implementation in #118.
//
// GitHub Issue: #115
// Spec: specs/009-job-execution-mvp/spec.md

package ports

import (
	"context"
	"time"
)

// ComposeController manages Docker Compose stacks via Docker API.
// Uses container labels for stack discovery and dependency ordering.
//
// Implementation Notes:
//   - Stack identification: com.docker.compose.project label
//   - Dependency order: com.docker.compose.depends_on label + topological sort
//   - Stop order: Reverse dependency (dependents first)
//   - Start order: Forward dependency (dependencies first)
//
// M3 Limitations:
//   - Does NOT wait for health checks after start
//   - Does NOT support depends_on: condition: service_healthy
type ComposeController interface {
	// StopStack stops all containers in a Compose stack.
	// Containers are stopped in reverse dependency order (dependents first).
	//
	// Behavior:
	//   - Sends SIGTERM to each container
	//   - Waits up to opts.Timeout per container
	//   - SIGKILL after grace period if still running
	//
	// Returns:
	//   - StopError if any container fails to stop
	//   - context.Canceled if context is cancelled
	StopStack(ctx context.Context, projectName string, opts StopOptions) error

	// StartStack starts all containers in a Compose stack.
	// Containers are started in dependency order (dependencies first).
	//
	// Behavior:
	//   - Starts containers sequentially in topological order
	//   - Does NOT wait for health checks (M3 limitation)
	//
	// Returns:
	//   - StartError if any container fails to start
	//   - context.Canceled if context is cancelled
	StartStack(ctx context.Context, projectName string, opts StartOptions) error

	// ListStackContainers returns all containers belonging to a stack.
	// Uses com.docker.compose.project label for identification.
	//
	// Returns:
	//   - All containers with matching project label
	//   - Empty slice if stack not found (not an error)
	ListStackContainers(ctx context.Context, projectName string) ([]StackContainer, error)

	// IsStackRunning returns true if all containers in the stack are running.
	// Returns false if any container is not in "running" state.
	IsStackRunning(ctx context.Context, projectName string) (bool, error)
}

// StopOptions configures stack stop behavior.
type StopOptions struct {
	// Timeout for stopping each container.
	// Default: 30s (from jobs.DefaultStopTimeout)
	// After timeout, Docker sends SIGKILL.
	Timeout time.Duration
}

// DefaultStopOptions returns options with default values.
func DefaultStopOptions() StopOptions {
	return StopOptions{
		Timeout: 30 * time.Second,
	}
}

// StartOptions configures stack start behavior.
type StartOptions struct {
	// Timeout for starting each container.
	// Default: 30s (from jobs.DefaultStartTimeout)
	Timeout time.Duration
}

// DefaultStartOptions returns options with default values.
func DefaultStartOptions() StartOptions {
	return StartOptions{
		Timeout: 30 * time.Second,
	}
}

// StackContainer represents a container within a Compose stack.
type StackContainer struct {
	// ID is the Docker container ID.
	ID string

	// Name is the container name (without leading /).
	Name string

	// ServiceName from com.docker.compose.service label.
	ServiceName string

	// State is the container state (running, exited, etc.).
	State string

	// DependsOn lists service names this container depends on.
	// Parsed from com.docker.compose.depends_on label.
	DependsOn []string

	// Labels contains all container labels.
	Labels map[string]string
}
