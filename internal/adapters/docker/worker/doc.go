// Package worker implements backup worker container execution.
//
// This package provides an adapter that implements the ports.WorkerRunner
// interface using the Docker SDK for Go. It manages worker container lifecycle:
// creation, execution, log capture, timeout enforcement, and cleanup.
//
// # Container Lifecycle
//
//  1. Create container with image, environment, and volume mounts
//  2. Start container
//  3. Stream logs from stdout/stderr
//  4. Wait for container to exit or timeout
//  5. On timeout: Send SIGTERM, wait grace period, send SIGKILL
//  6. Capture exit code and logs
//  7. Remove container (unless KeepOnFailure and failed)
//
// # Worker Contract
//
// Worker containers must:
//   - Have a default ENTRYPOINT or CMD that performs backup
//   - Exit with code 0 on success, non-zero on failure
//   - Handle SIGTERM gracefully (optional but recommended)
//   - Write logs to stdout/stderr
//
// Worker containers receive these environment variables:
//   - BOSUN_JOB_NAME: Job identifier
//   - BOSUN_RUN_ID: Unique execution ID (UUID v4)
//   - BOSUN_STACK: Target Compose stack name
//   - BOSUN_DRY_RUN: "true" or "false"
//   - Additional user-defined vars from bosun.job.worker.env.* labels
//
// # Timeout Behavior
//
// When the timeout is reached:
//  1. Send SIGTERM to container PID 1
//  2. Wait up to 10 seconds (grace period)
//  3. If still running, send SIGKILL
//  4. Exit code will be 137 (128 + SIGKILL=9) if killed
//
// # References
//
//   - GitHub Issue: #119
//   - Port Interface: internal/ports/worker.go (#116)
//   - Research: .serena/memories/m3_worker_contract.md
package worker
