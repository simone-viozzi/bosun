package executor

import (
	"context"
	"time"
)

// noopEventEmitter is a no-op implementation of ports.EventEmitter used when
// no emitter is provided to New(). It silently discards all events.
// This avoids nil pointer dereferences in CLI commands and tests that do not
// require event tracking (T076 will inject a real emitter for daemon mode).
type noopEventEmitter struct{}

func (noopEventEmitter) EmitJobScheduled(_ context.Context, _ string, _ string, _ time.Time)     {}
func (noopEventEmitter) EmitJobStarted(_ context.Context, _ string, _ string)                    {}
func (noopEventEmitter) EmitJobCompleted(_ context.Context, _ string, _ string, _ time.Duration) {}
func (noopEventEmitter) EmitJobFailed(_ context.Context, _ string, _ string, _ error, _ time.Duration) {
}
func (noopEventEmitter) EmitJobSkipped(_ context.Context, _ string, _ string)       {}
func (noopEventEmitter) EmitJobAdded(_ context.Context, _ string)                   {}
func (noopEventEmitter) EmitJobRemoved(_ context.Context, _ string)                 {}
func (noopEventEmitter) EmitJobChanged(_ context.Context, _ string, _, _ string)    {}
func (noopEventEmitter) EmitJobCircuitBroken(_ context.Context, _ string, _ int)    {}
func (noopEventEmitter) EmitJobDuplicateName(_ context.Context, _ string, _ string) {}
