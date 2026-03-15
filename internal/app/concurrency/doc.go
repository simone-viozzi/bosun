// Package concurrency provides concurrency control primitives for the scheduler.
//
// The primary type is StackLockManager, which provides per-stack mutual exclusion
// to prevent concurrent job execution on the same Compose stack. Locks are acquired
// in sorted alphabetical order to prevent deadlocks when jobs target overlapping
// stack sets.
package concurrency
