package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// Exit code constants for signal-based container termination.
// Unix convention: exit code = 128 + signal number.
const (
	// ExitCodeSIGKILL is returned when a container is killed by SIGKILL (signal 9).
	ExitCodeSIGKILL = 128 + 9 // 137

	// ExitCodeSIGTERM is returned when a container is terminated by SIGTERM (signal 15).
	ExitCodeSIGTERM = 128 + 15 // 143
)

// Runner implements ports.WorkerRunner using Docker API.
type Runner struct {
	docker *client.Client
}

// NewRunner creates a new worker runner.
func NewRunner(docker *client.Client) *Runner {
	return &Runner{
		docker: docker,
	}
}

// Run executes a worker container and waits for completion.
func (r *Runner) Run(ctx context.Context, config ports.WorkerConfig) (ports.WorkerResult, error) {
	startTime := time.Now()

	// Generate container name
	containerName := fmt.Sprintf(jobs.WorkerContainerNameFormat,
		config.JobName,
		config.RunID[:8], // First 8 chars of UUID
	)

	// Convert mounts to Docker format
	mounts := convertMounts(config.Mounts)

	// Create container
	resp, err := r.docker.ContainerCreate(ctx,
		&container.Config{
			Image: config.Image,
			Env:   config.BuildEnv(),
		},
		&container.HostConfig{
			Mounts: mounts,
		},
		nil, nil,
		containerName,
	)
	if err != nil {
		return ports.WorkerResult{}, fmt.Errorf("failed to create worker container: %w", err)
	}

	containerID := resp.ID

	// Ensure cleanup
	shouldRemove := true
	defer func() {
		if shouldRemove {
			// Best effort removal - ignore errors
			_ = r.docker.ContainerRemove(context.Background(), containerID, container.RemoveOptions{
				Force: true,
			})
		}
	}()

	// Start container
	if err := r.docker.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return ports.WorkerResult{}, fmt.Errorf("failed to start worker container: %w", err)
	}

	// Set up log buffer for capturing
	var logBuf bytes.Buffer
	var logWriter io.Writer = &logBuf

	// If real-time streaming requested, use MultiWriter
	if config.LogWriter != nil {
		logWriter = io.MultiWriter(&logBuf, config.LogWriter)
	}

	// Start streaming logs in background
	logDone := make(chan error, 1)
	go func() {
		logDone <- r.streamLogs(ctx, containerID, logWriter)
	}()

	// Wait for container with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	statusCh, errCh := r.docker.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

	var exitCode int
	timedOut := false

	select {
	case err := <-errCh:
		if err != nil {
			// Timeout or other error - stop container
			timedOut = true
			var stopErr error
			exitCode, stopErr = r.stopContainer(context.Background(), containerID)
			if stopErr != nil {
				slog.Warn("failed to stop container",
					"container", containerID,
					"error", stopErr,
				)
			}
		}
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	}

	// Wait for log streaming to complete
	if logErr := <-logDone; logErr != nil {
		slog.Warn("log streaming failed",
			"container", containerID,
			"error", logErr,
		)
	}

	duration := time.Since(startTime)

	// Determine if we should keep the container
	if config.KeepOnFailure && exitCode != 0 {
		shouldRemove = false
	}

	return ports.WorkerResult{
		ExitCode:    exitCode,
		Logs:        logBuf.String(),
		ContainerID: containerID,
		Duration:    duration,
		TimedOut:    timedOut,
	}, nil
}

// stopContainer stops a container with SIGTERM→SIGKILL grace period.
// Returns the exit code after stop and any error encountered.
func (r *Runner) stopContainer(ctx context.Context, containerID string) (int, error) {
	timeout := int(jobs.GracePeriod.Seconds())
	if err := r.docker.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		slog.Warn("ContainerStop failed",
			"container", containerID,
			"error", err,
		)
	}

	// Inspect to get exit code
	inspect, err := r.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return ExitCodeSIGKILL, fmt.Errorf("failed to inspect container %s after stop: %w", containerID, err)
	}

	return inspect.State.ExitCode, nil
}

// streamLogs streams stdout/stderr from a container to the writer.
// This function blocks until the container exits or context is cancelled.
// Docker multiplexes stdout/stderr with 8-byte headers when TTY=false;
// stdcopy.StdCopy strips these headers and routes each frame correctly.
func (r *Runner) streamLogs(ctx context.Context, containerID string, writer io.Writer) error {
	reader, err := r.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true, // Stream until container exits
	})
	if err != nil {
		return fmt.Errorf("failed to get container logs for %s: %w", containerID, err)
	}
	defer func() { _ = reader.Close() }()

	// Demultiplex Docker log stream — both stdout and stderr written to same writer.
	if _, err := stdcopy.StdCopy(writer, writer, reader); err != nil {
		return fmt.Errorf("failed to stream logs for container %s: %w", containerID, err)
	}

	return nil
}

// convertMounts converts ports.VolumeMount to Docker mount format.
func convertMounts(mounts []ports.VolumeMount) []mount.Mount {
	if len(mounts) == 0 {
		return nil
	}

	result := make([]mount.Mount, 0, len(mounts))
	for _, m := range mounts {
		result = append(result, mount.Mount{
			Type:     mount.TypeVolume,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	return result
}
