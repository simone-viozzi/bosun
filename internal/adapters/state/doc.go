// Package state provides JobStateStore adapter implementations.
//
// The default adapter is InMemoryStateStore, which uses sync.Map for zero-durability
// state storage (state is lost on daemon restart). This is the M4 default.
//
// Issue #177 adds a durable adapter (BoltDB or SQLite) as a drop-in replacement
// to enable catch-up runs and persistent circuit-breaker state across restarts.
package state
