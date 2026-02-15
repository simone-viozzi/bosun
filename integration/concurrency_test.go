//go:build integration

package integration

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/app/scheduler"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// --- T078: Per-stack mutex serializes jobs targeting same stack ---

func TestIntegration_PerStackMutex_Serialization(t *testing.T) {
	t.Parallel()

	var maxConcurrent int32
	var currentConcurrent int32
	var mu sync.Mutex
	executionOrder := make([]string, 0)

	exec := newIntegrationMockExecutor()
	exec.execFunc = func(_ context.Context, job jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		cur := atomic.AddInt32(&currentConcurrent, 1)
		if cur > atomic.LoadInt32(&maxConcurrent) {
			atomic.StoreInt32(&maxConcurrent, cur)
		}
		mu.Lock()
		executionOrder = append(executionOrder, job.Name)
		mu.Unlock()

		time.Sleep(200 * time.Millisecond) // Simulate work
		atomic.AddInt32(&currentConcurrent, -1)
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}

	ev := &integrationMockEvents{}
	store := newIntegrationMockStateStore()

	// Parallelism=5 but stack lock should serialize.
	s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 5}, slog.Default())
	ctx := context.Background()

	// Two jobs targeting the same stack.
	for _, name := range []string{"stack-job-a", "stack-job-b"} {
		job := jobs.Job{
			Name:         name,
			Schedule:     "* * * * * *",
			Enabled:      true,
			TargetStacks: []string{"shared-stack"},
		}
		if err := s.AddJob(ctx, job); err != nil {
			t.Fatalf("AddJob(%q): %v", name, err)
		}
	}

	s.StartCron()
	defer s.StopCron()

	// Wait for several executions.
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 4 {
		select {
		case <-exec.called:
			received++
		case <-timeout:
			t.Fatalf("timed out after %d executions", received)
		}
	}

	// With same stack, max concurrent should be 1 (serialized by stack lock).
	if maxConcurrent > 1 {
		t.Errorf("max concurrent = %d, want 1 (same stack should serialize)", maxConcurrent)
	}
}

// --- T079: Global semaphore limits parallelism ---

func TestIntegration_GlobalSemaphore_LimitsParallelism(t *testing.T) {
	t.Parallel()

	var maxConcurrent int32
	var currentConcurrent int32

	exec := newIntegrationMockExecutor()
	exec.execFunc = func(_ context.Context, job jobs.Job, _ ports.ExecuteOptions) (ports.ExecutionResult, error) {
		cur := atomic.AddInt32(&currentConcurrent, 1)
		for {
			old := atomic.LoadInt32(&maxConcurrent)
			if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
				break
			}
		}
		time.Sleep(300 * time.Millisecond) // Simulate work
		atomic.AddInt32(&currentConcurrent, -1)
		return ports.ExecutionResult{
			Run: jobs.JobRun{Status: jobs.RunStatusSuccess},
		}, nil
	}

	ev := &integrationMockEvents{}
	store := newIntegrationMockStateStore()

	// Serial mode: parallelism = 1.
	t.Run("Serial_Parallelism1", func(t *testing.T) {
		atomic.StoreInt32(&maxConcurrent, 0)
		atomic.StoreInt32(&currentConcurrent, 0)

		s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 1}, slog.Default())
		ctx := context.Background()

		for _, name := range []string{"serial-a", "serial-b"} {
			job := jobs.Job{
				Name:     name,
				Schedule: "* * * * * *",
				Enabled:  true,
			}
			if err := s.AddJob(ctx, job); err != nil {
				t.Fatalf("AddJob(%q): %v", name, err)
			}
		}

		s.StartCron()
		received := 0
		timeout := time.After(5 * time.Second)
		for received < 3 {
			select {
			case <-exec.called:
				received++
			case <-timeout:
				t.Fatalf("timed out after %d executions", received)
			}
		}
		s.StopCron()

		if atomic.LoadInt32(&maxConcurrent) > 1 {
			t.Errorf("max concurrent = %d, want <= 1 (Parallelism=1)", maxConcurrent)
		}
	})

	// Concurrent mode: parallelism = 3.
	t.Run("Concurrent_Parallelism3", func(t *testing.T) {
		atomic.StoreInt32(&maxConcurrent, 0)
		atomic.StoreInt32(&currentConcurrent, 0)

		s := scheduler.New(exec, ev, store, scheduler.Options{Parallelism: 3}, slog.Default())
		ctx := context.Background()

		for _, name := range []string{"par-a", "par-b", "par-c"} {
			job := jobs.Job{
				Name:     name,
				Schedule: "* * * * * *",
				Enabled:  true,
			}
			if err := s.AddJob(ctx, job); err != nil {
				t.Fatalf("AddJob(%q): %v", name, err)
			}
		}

		s.StartCron()
		received := 0
		timeout := time.After(5 * time.Second)
		for received < 6 {
			select {
			case <-exec.called:
				received++
			case <-timeout:
				t.Fatalf("timed out after %d executions", received)
			}
		}
		s.StopCron()

		// Hard limit is 3.
		if atomic.LoadInt32(&maxConcurrent) > 3 {
			t.Errorf("max concurrent = %d, want <= 3 (Parallelism=3)", maxConcurrent)
		}
	})
}
