package state_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/adapters/state"
	"github.com/simone-viozzi/bosun/internal/ports"
)

func TestInMemoryStateStore_SaveAndLoad(t *testing.T) {
	store := state.NewInMemoryStateStore()
	ctx := context.Background()
	now := time.Now()

	input := ports.JobState{
		JobName:             "daily-backup",
		LastRunAt:           &now,
		LastResult:          "success",
		ConsecutiveFailures: 0,
	}

	if err := store.SaveJobState(ctx, input); err != nil {
		t.Fatalf("SaveJobState() error = %v", err)
	}

	got, err := store.LoadJobState(ctx, "daily-backup")
	if err != nil {
		t.Fatalf("LoadJobState() error = %v", err)
	}
	if got.JobName != input.JobName {
		t.Errorf("JobName = %q, want %q", got.JobName, input.JobName)
	}
	if got.LastResult != input.LastResult {
		t.Errorf("LastResult = %q, want %q", got.LastResult, input.LastResult)
	}
	if got.ConsecutiveFailures != input.ConsecutiveFailures {
		t.Errorf("ConsecutiveFailures = %d, want %d", got.ConsecutiveFailures, input.ConsecutiveFailures)
	}
}

func TestInMemoryStateStore_LoadNotFound(t *testing.T) {
	store := state.NewInMemoryStateStore()
	ctx := context.Background()

	_, err := store.LoadJobState(ctx, "nonexistent")
	if !errors.Is(err, ports.ErrJobStateNotFound) {
		t.Errorf("LoadJobState() error = %v, want ErrJobStateNotFound", err)
	}
}

func TestInMemoryStateStore_ListJobStates(t *testing.T) {
	store := state.NewInMemoryStateStore()
	ctx := context.Background()

	// Empty store
	states, err := store.ListJobStates(ctx)
	if err != nil {
		t.Fatalf("ListJobStates() error = %v", err)
	}
	if len(states) != 0 {
		t.Errorf("ListJobStates() returned %d items, want 0", len(states))
	}

	// Add two jobs
	_ = store.SaveJobState(ctx, ports.JobState{JobName: "job-a", LastResult: "success"})
	_ = store.SaveJobState(ctx, ports.JobState{JobName: "job-b", LastResult: "error: timeout"})

	states, err = store.ListJobStates(ctx)
	if err != nil {
		t.Fatalf("ListJobStates() error = %v", err)
	}
	if len(states) != 2 {
		t.Errorf("ListJobStates() returned %d items, want 2", len(states))
	}

	// Verify both jobs are present
	found := make(map[string]bool)
	for _, s := range states {
		found[s.JobName] = true
	}
	if !found["job-a"] || !found["job-b"] {
		t.Errorf("ListJobStates() missing expected jobs: %v", found)
	}
}

func TestInMemoryStateStore_DeleteJobState(t *testing.T) {
	store := state.NewInMemoryStateStore()
	ctx := context.Background()

	_ = store.SaveJobState(ctx, ports.JobState{JobName: "to-delete", LastResult: "success"})

	// Verify it exists
	_, err := store.LoadJobState(ctx, "to-delete")
	if err != nil {
		t.Fatalf("LoadJobState() before delete error = %v", err)
	}

	// Delete
	if err := store.DeleteJobState(ctx, "to-delete"); err != nil {
		t.Fatalf("DeleteJobState() error = %v", err)
	}

	// Verify it's gone
	_, err = store.LoadJobState(ctx, "to-delete")
	if !errors.Is(err, ports.ErrJobStateNotFound) {
		t.Errorf("LoadJobState() after delete error = %v, want ErrJobStateNotFound", err)
	}
}

func TestInMemoryStateStore_DeleteNonexistent(t *testing.T) {
	store := state.NewInMemoryStateStore()
	ctx := context.Background()

	// Delete nonexistent should not error
	if err := store.DeleteJobState(ctx, "nonexistent"); err != nil {
		t.Errorf("DeleteJobState(nonexistent) error = %v, want nil", err)
	}
}

func TestInMemoryStateStore_SaveOverwrites(t *testing.T) {
	store := state.NewInMemoryStateStore()
	ctx := context.Background()

	// Save initial state
	_ = store.SaveJobState(ctx, ports.JobState{
		JobName:             "job-x",
		LastResult:          "success",
		ConsecutiveFailures: 0,
	})

	// Overwrite
	_ = store.SaveJobState(ctx, ports.JobState{
		JobName:             "job-x",
		LastResult:          "error: timeout",
		ConsecutiveFailures: 1,
	})

	got, _ := store.LoadJobState(ctx, "job-x")
	if got.LastResult != "error: timeout" {
		t.Errorf("LastResult = %q, want %q", got.LastResult, "error: timeout")
	}
	if got.ConsecutiveFailures != 1 {
		t.Errorf("ConsecutiveFailures = %d, want 1", got.ConsecutiveFailures)
	}
}
