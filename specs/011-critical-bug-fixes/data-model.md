# Data Model: Milestone 3.75 — Critical Bug Fixes

**Date**: 2026-02-15

## No New Entities

This milestone is purely bug-fix. No new data entities, types, or relationships are introduced.

## Existing Entities Modified (Internal Only)

### Exit Code Constants (new, `runner.go`)

```go
const (
    // ExitCodeSIGKILL is the exit code when a container is killed by SIGKILL (128 + 9).
    ExitCodeSIGKILL = 137

    // ExitCodeSIGTERM is the exit code when a container is terminated by SIGTERM (128 + 15).
    ExitCodeSIGTERM = 143
)
```

These constants replace the current magic number `137` in `stopContainer`. They are internal to the worker adapter package — not exported via ports.

### Function Signature Changes (internal, no port changes)

| Function | Current | New |
|----------|---------|-----|
| `streamLogs` | `streamLogs(ctx, containerID, writer)` (no return) | `streamLogs(ctx, containerID, writer) error` |
| `stopContainer` | `stopContainer(ctx, containerID) int` | `stopContainer(ctx, containerID) (int, error)` |

Both are unexported methods on `*Runner`. No port interface changes required.
