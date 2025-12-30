// Package executor implements job execution orchestration.
//
// This package provides the Executor service that coordinates the complete
// job execution lifecycle: job discovery, plan generation, image validation,
// stack stop, worker execution, and stack restart.
//
// # Execution Flow
//
//  1. Discover job by name (via JobDiscoverer port)
//  2. Generate execution plan (via JobPlanner port)
//  3. Pre-validate worker image exists (fail fast)
//  4. Stop target Compose stack (via ComposeController port)
//  5. Run worker container (via WorkerRunner port)
//  6. Restart stack (ALWAYS, even on worker failure)
//
// # Guaranteed Restart
//
// The executor uses defer to ensure that the Compose stack is restarted
// even if:
//   - Worker fails with non-zero exit code
//   - Worker times out
//   - Execution is interrupted (Ctrl+C)
//   - Context is cancelled
//
// This design prevents leaving production stacks in a stopped state.
//
// # Error Handling
//
//   - Pre-validation failure: Return error, stack NOT stopped
//   - Stop failure: Return error, stack may be partially stopped
//   - Worker failure: Restart stack, return WorkerError
//   - Start failure: Return error, stack may be partially started
//
// # Signal Handling
//
// The executor respects context cancellation (e.g., Ctrl+C). When cancelled:
//  1. In-flight operations are cancelled
//  2. Worker is stopped (SIGTERM → SIGKILL)
//  3. Stack is restarted (best effort)
//  4. ExecutionResult reflects cancellation
//
// # References
//
//   - GitHub Issue: #121
//   - Port Interface: internal/ports/executor.go (#114)
//   - Research: .serena/memories/m3_failure_handling.md
package executor
