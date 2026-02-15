package concurrency_test

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/app/concurrency"
)

// --- T035: Unit test StackLockManager.Lock/Unlock — mutual exclusion on same stack ---

func TestStackLockManager_Lock_MutualExclusion(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	// Lock the stack.
	if err := mgr.Lock(ctx, "web-stack"); err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	// Verify it's locked.
	if !mgr.IsLocked("web-stack") {
		t.Error("IsLocked() = false after Lock()")
	}

	// Attempt to lock from another goroutine — should block.
	blocked := make(chan struct{})
	acquired := make(chan struct{})
	go func() {
		close(blocked)
		_ = mgr.Lock(ctx, "web-stack")
		close(acquired)
	}()

	<-blocked
	// Give goroutine time to block on the mutex.
	time.Sleep(50 * time.Millisecond)

	// Should NOT have acquired yet.
	select {
	case <-acquired:
		t.Fatal("second Lock() should have blocked, but acquired immediately")
	default:
		// good — blocked as expected
	}

	// Unlock — the waiting goroutine should acquire.
	mgr.Unlock("web-stack")

	select {
	case <-acquired:
		// good
	case <-time.After(2 * time.Second):
		t.Fatal("second Lock() should have acquired after Unlock(), but timed out")
	}

	// Clean up.
	mgr.Unlock("web-stack")
}

func TestStackLockManager_Lock_IndependentStacks(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	// Lock two different stacks — should not interfere.
	if err := mgr.Lock(ctx, "stack-a"); err != nil {
		t.Fatalf("Lock(stack-a) error = %v", err)
	}
	if err := mgr.Lock(ctx, "stack-b"); err != nil {
		t.Fatalf("Lock(stack-b) error = %v", err)
	}

	if !mgr.IsLocked("stack-a") || !mgr.IsLocked("stack-b") {
		t.Error("both stacks should be locked")
	}

	mgr.Unlock("stack-a")
	mgr.Unlock("stack-b")
}

func TestStackLockManager_IsLocked_UnknownStack(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	if mgr.IsLocked("nonexistent") {
		t.Error("IsLocked() = true for never-locked stack")
	}
}

// --- T036: Unit test StackLockManager.LockAll — sorted alphabetical acquisition ---

func TestStackLockManager_LockAll_SortedAcquisition(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	// LockAll with unsorted stacks.
	stacks := []string{"db-stack", "web-stack", "cache-stack"}
	if err := mgr.LockAll(ctx, stacks); err != nil {
		t.Fatalf("LockAll() error = %v", err)
	}

	// All should be locked.
	for _, s := range stacks {
		if !mgr.IsLocked(s) {
			t.Errorf("IsLocked(%q) = false after LockAll", s)
		}
	}

	mgr.UnlockAll(stacks)

	// All should be unlocked.
	for _, s := range stacks {
		if mgr.IsLocked(s) {
			t.Errorf("IsLocked(%q) = true after UnlockAll", s)
		}
	}
}

func TestStackLockManager_LockAll_NoDeadlock(t *testing.T) {
	mgr := concurrency.NewStackLockManager()

	// Two goroutines trying to lock overlapping stack sets.
	// Without sorted ordering, this could deadlock.
	stacksA := []string{"web-stack", "db-stack"}
	stacksB := []string{"db-stack", "web-stack"}

	var wg sync.WaitGroup
	wg.Add(2)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		_ = mgr.LockAll(ctx, stacksA)
		time.Sleep(10 * time.Millisecond)
		mgr.UnlockAll(stacksA)
	}()

	go func() {
		defer wg.Done()
		ctx := context.Background()
		_ = mgr.LockAll(ctx, stacksB)
		time.Sleep(10 * time.Millisecond)
		mgr.UnlockAll(stacksB)
	}()

	select {
	case <-done:
		// No deadlock — both goroutines completed.
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected: goroutines did not complete within 5 seconds")
	}
}

func TestStackLockManager_LockAll_Dedup(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	// Duplicate stacks should be deduplicated.
	stacks := []string{"web-stack", "web-stack", "db-stack"}
	if err := mgr.LockAll(ctx, stacks); err != nil {
		t.Fatalf("LockAll() error = %v", err)
	}

	// Should work without deadlocking (locking same mutex twice would deadlock).
	mgr.UnlockAll([]string{"web-stack", "db-stack"})
}

func TestStackLockManager_LockAll_Empty(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	if err := mgr.LockAll(ctx, nil); err != nil {
		t.Fatalf("LockAll(nil) error = %v", err)
	}
	if err := mgr.LockAll(ctx, []string{}); err != nil {
		t.Fatalf("LockAll([]) error = %v", err)
	}
}

// --- T037: Unit test StackLockManager.LockAll — context cancellation releases already-acquired locks ---

func TestStackLockManager_LockAll_ContextCancellation(t *testing.T) {
	mgr := concurrency.NewStackLockManager()

	// Pre-lock "db-stack" so LockAll will block on it.
	bgCtx := context.Background()
	if err := mgr.Lock(bgCtx, "db-stack"); err != nil {
		t.Fatalf("Lock(db-stack) error = %v", err)
	}

	// Create a cancellable context with short timeout.
	ctx, cancel := context.WithTimeout(bgCtx, 200*time.Millisecond)
	defer cancel()

	// LockAll should acquire "cache-stack" (alphabetically first) then block on "db-stack".
	stacks := []string{"db-stack", "cache-stack"}
	err := mgr.LockAll(ctx, stacks)
	if err == nil {
		t.Fatal("LockAll() should have returned error on context cancellation")
	}

	// The already-acquired lock ("cache-stack") should have been released.
	time.Sleep(50 * time.Millisecond)
	if mgr.IsLocked("cache-stack") {
		t.Error("cache-stack should be unlocked after rollback on context cancellation")
	}

	// Clean up the pre-locked "db-stack".
	mgr.Unlock("db-stack")
}

func TestStackLockManager_Lock_ContextCancellation(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	// Lock the stack.
	if err := mgr.Lock(ctx, "contended-stack"); err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	// Try to lock with a cancelled context.
	cancelCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	err := mgr.Lock(cancelCtx, "contended-stack")
	if err == nil {
		t.Fatal("Lock() should have returned error on context cancellation")
	}

	// Original lock should still be held.
	if !mgr.IsLocked("contended-stack") {
		t.Error("original lock should still be held")
	}

	mgr.Unlock("contended-stack")
}

// --- T038: Concurrent mutual exclusion stress test ---

func TestStackLockManager_ConcurrentMutualExclusion(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	var (
		counter atomic.Int32
		maxSeen atomic.Int32
		wg      sync.WaitGroup
		workers = 10
	)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_ = mgr.Lock(ctx, "shared-stack")
			defer mgr.Unlock("shared-stack")

			cur := counter.Add(1)
			for {
				old := maxSeen.Load()
				if cur <= old || maxSeen.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(1 * time.Millisecond)
			counter.Add(-1)
		}()
	}

	wg.Wait()

	if maxSeen.Load() > 1 {
		t.Errorf("max concurrent = %d, want 1 (mutual exclusion violated)", maxSeen.Load())
	}
}

func TestStackLockManager_LockAll_Order(t *testing.T) {
	mgr := concurrency.NewStackLockManager()
	ctx := context.Background()

	stacks := []string{"z-stack", "a-stack", "m-stack"}
	expected := []string{"a-stack", "m-stack", "z-stack"}

	if err := mgr.LockAll(ctx, stacks); err != nil {
		t.Fatalf("LockAll() error = %v", err)
	}
	mgr.UnlockAll(stacks)

	if !sort.StringsAreSorted(expected) {
		t.Error("expected list should be sorted")
	}
}
