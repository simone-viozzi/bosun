// Package concurrency provides concurrency control primitives for job scheduling.
package concurrency

import (
	"context"
	"sort"
	"sync"
)

// StackLockManager provides per-stack mutual exclusion for preventing concurrent
// execution on the same Compose stack. Lock acquisition order is alphabetical
// to prevent deadlocks when jobs target overlapping stack sets.
//
// All methods are safe for concurrent use.
type StackLockManager struct {
	locks sync.Map // map[string]*sync.Mutex
}

// NewStackLockManager creates a new StackLockManager.
func NewStackLockManager() *StackLockManager {
	return &StackLockManager{}
}

// getOrCreateMutex returns the mutex for the given stack name, creating one if necessary.
func (m *StackLockManager) getOrCreateMutex(stackName string) *sync.Mutex {
	val, _ := m.locks.LoadOrStore(stackName, &sync.Mutex{})
	return val.(*sync.Mutex)
}

// Lock acquires an exclusive lock for a single stack.
// It blocks until the lock is available or the context is cancelled.
// Returns nil on success, ctx.Err() if context is cancelled.
func (m *StackLockManager) Lock(ctx context.Context, stackName string) error {
	mu := m.getOrCreateMutex(stackName)

	// Try to acquire with context cancellation support.
	acquired := make(chan struct{})
	go func() {
		mu.Lock()
		close(acquired)
	}()

	select {
	case <-acquired:
		return nil
	case <-ctx.Done():
		// Context cancelled before lock acquired. The background goroutine above
		// will eventually acquire the lock and immediately release it.
		// Limitation: if the current lock holder itself is cancelled and never
		// unlocks (e.g., a cascading cancellation deadlock), this cleanup goroutine
		// will block indefinitely. In practice this is bounded: executeJob always
		// releases stack locks via defer, so holders always unlock eventually.
		go func() {
			<-acquired
			mu.Unlock()
		}()
		return ctx.Err()
	}
}

// LockAll acquires locks for multiple stacks in sorted (alphabetical) order.
// Sorted acquisition prevents deadlocks when two jobs target overlapping stack sets.
// On context cancellation, releases any already-acquired locks before returning.
func (m *StackLockManager) LockAll(ctx context.Context, stacks []string) error {
	if len(stacks) == 0 {
		return nil
	}

	// Deduplicate and sort for consistent ordering.
	sorted := dedupSort(stacks)

	// Track which locks we've acquired for rollback on error.
	acquired := make([]string, 0, len(sorted))

	for _, stack := range sorted {
		if err := m.Lock(ctx, stack); err != nil {
			// Rollback: release all acquired locks.
			m.UnlockAll(acquired)
			return err
		}
		acquired = append(acquired, stack)
	}

	return nil
}

// Unlock releases the lock for the named stack.
// It must be called exactly once per successful Lock acquisition for a given stack.
// Calling Unlock on a stack that is not currently locked will panic (sync.Mutex semantics).
// Callers should rely on proper Lock/Unlock pairing; use UnlockAll with deduplicated
// input to avoid accidental double-unlocks.
func (m *StackLockManager) Unlock(stackName string) {
	val, ok := m.locks.Load(stackName)
	if !ok {
		return
	}
	mu := val.(*sync.Mutex)
	mu.Unlock()
}

// UnlockAll releases locks for multiple stacks.
// Deduplicates the input to prevent calling Unlock multiple times on the same stack.
func (m *StackLockManager) UnlockAll(stacks []string) {
	deduped := dedupSort(stacks)
	for _, stack := range deduped {
		m.Unlock(stack)
	}
}

// IsLocked returns whether a stack is currently locked.
// Useful for observability/debugging, not for synchronization.
func (m *StackLockManager) IsLocked(stackName string) bool {
	val, ok := m.locks.Load(stackName)
	if !ok {
		return false
	}
	mu := val.(*sync.Mutex)
	// Try to acquire — if we can, it wasn't locked.
	if mu.TryLock() {
		mu.Unlock()
		return false
	}
	return true
}

// dedupSort returns a sorted, deduplicated copy of the input slice.
func dedupSort(stacks []string) []string {
	if len(stacks) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(stacks))
	result := make([]string, 0, len(stacks))
	for _, s := range stacks {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	sort.Strings(result)
	return result
}
