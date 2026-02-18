//go:build integration

package integration

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/app/scheduler"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
)

// --- T080: Daemon lifecycle — start, signal handling, graceful shutdown ---

func TestIntegration_DaemonLifecycle_GracefulShutdown(t *testing.T) {
	t.Parallel()

	exec := newIntegrationMockExecutor()
	ev := &integrationMockEvents{}
	store := newIntegrationMockStateStore()

	s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 2}, slog.Default())
	ctx := context.Background()

	// Register a job.
	job := jobs.Job{
		Name:     "lifecycle-job",
		Schedule: "@every 500ms",
		Enabled:  true,
	}
	if err := s.AddJob(ctx, job); err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Start scheduler in background, simulating daemon.
	ctx, cancel := context.WithCancel(context.Background())
	startErr := make(chan error, 1)
	go func() {
		startErr <- s.Start(ctx)
	}()

	// Wait for at least one execution.
	select {
	case <-exec.called:
		// Job fired successfully.
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("timed out waiting for job execution")
	}

	// Signal shutdown (like SIGTERM).
	cancel()

	// Wait for Start to return.
	select {
	case err := <-startErr:
		if err != nil {
			t.Errorf("Start returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for graceful shutdown")
	}

	// Verify events.
	if !ev.hasEvent("JobScheduled") {
		t.Error("expected JobScheduled event")
	}
	if !ev.hasEvent("JobStarted") {
		t.Error("expected JobStarted event")
	}
}
