// Package contracts defines the port interfaces for M4 Scheduling Engine.
package contracts

import (
	"context"
)

// StackLockManager provides per-stack mutex for preventing concurrent
// execution on the same Compose stack.
//
// This is an app-level service, not a port interface (internal implementation detail).
// Implementation location: internal/app/concurrency/stack_lock.go
type StackLockManager interface {
	// Lock acquires exclusive lock for a stack.
	// Blocks until lock is available or context is cancelled.
	// Returns nil on success, ctx.Err() if cancelled.
	//
	// Lock acquisition is re-entrant safe (same goroutine can acquire multiple times).
	// However, locks should be released via defer to ensure cleanup.
	Lock(ctx context.Context, stackName string) error

	// LockAll acquires locks for multiple stacks in sorted (alphabetical) order.
	// Sorted acquisition prevents deadlocks when two jobs target overlapping stack sets.
	// On context cancellation or error, releases any already-acquired locks.
	LockAll(ctx context.Context, stacks []string) error

	// Unlock releases lock for a stack.
	// Safe to call multiple times (idempotent).
	// If stack was never locked, this is a no-op.
	Unlock(stackName string)

	// UnlockAll releases locks for multiple stacks.
	// Safe to call multiple times (idempotent).
	UnlockAll(stacks []string)

	// IsLocked returns whether stack is currently locked.
	// Useful for observability/debugging, not for synchronization.
	IsLocked(stackName string) bool
}

// Implementation Notes:
//
// The StackLockManager uses sync.Map to store per-stack mutexes:
//
//	type StackLockManager struct {
//	    locks sync.Map // map[string]*sync.Mutex
//	}
//
// When Lock() is called:
// 1. Get or create mutex for stack (LoadOrStore)
// 2. Acquire mutex with context cancellation support
// 3. Return error if context cancelled
//
// When Unlock() is called:
// 1. Load mutex for stack
// 2. Release mutex (idempotent)
//
// Thread safety: All methods are safe for concurrent use.

// Example Usage:
//
// Basic lock/unlock:
//
//	mgr := concurrency.NewStackLockManager()
//
//	ctx := context.Background()
//	if err := mgr.Lock(ctx, "web-stack"); err != nil {
//	    log.Fatalf("Failed to acquire lock: %v", err)
//	}
//	defer mgr.Unlock("web-stack")
//
//	// Critical section: execute job on web-stack
//	executeJob(ctx)
//
// With timeout:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if err := mgr.Lock(ctx, "db-stack"); err != nil {
//	    log.Printf("Lock acquisition timed out: %v", err)
//	    return
//	}
//	defer mgr.Unlock("db-stack")
//
//	// Execute job
//
// Checking lock state (observability):
//
//	if mgr.IsLocked("cache-stack") {
//	    log.Printf("cache-stack is currently busy")
//	}
//
// Integration with Executor:
//
// The Scheduler wraps job execution with stack locking:
//
//	func (s *Scheduler) executeJob(ctx context.Context, entry *scheduledJobEntry) error {
//	    stacks := entry.job.TargetStacks
//
//	    // Acquire all stack locks (sorted order, deadlock-free)
//	    if err := s.stackLocks.LockAll(ctx, stacks); err != nil {
//	        return fmt.Errorf("failed to acquire stack locks: %w", err)
//	    }
//	    defer s.stackLocks.UnlockAll(stacks)
//
//	    // Execute job (stop stack, run worker, start stack)
//	    return s.executor.Execute(ctx, entry.job.Name, opts)
//	}

// Concurrency Guarantees:
//
// 1. Mutual Exclusion: Only one job can hold a lock for a given stack at a time.
// 2. Fairness: Locks are granted in FIFO order (via sync.Mutex semantics).
// 3. No Deadlocks: LockAll acquires in sorted alphabetical order; context cancellation as fallback.
// 4. Independent Stacks: Jobs on different stacks can run concurrently.
// 5. Multi-Stack Safety: Jobs targeting overlapping stack sets are serialized without deadlock.
//
// Example scenario:
// - Job A (targets ["db-stack", "web-stack"]) locks "db-stack" then "web-stack" (sorted)
// - Job B (targets ["web-stack", "db-stack"]) also locks "db-stack" then "web-stack" (same sorted order)
// - No deadlock possible regardless of concurrent scheduling
// - Job C (targets "cache-stack") acquires lock concurrently (non-overlapping stacks)
