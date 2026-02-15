# Research: Milestone 3.75 — Critical Bug Fixes

**Date**: 2026-02-15
**Feature**: [spec.md](spec.md)

## Research Tasks

### R1: Docker Log Multiplexing & stdcopy.StdCopy

**Question**: How does Docker multiplexing work, and how should `stdcopy.StdCopy` be used?

**Decision**: Use `stdcopy.StdCopy(writer, writer, reader)` to write both stdout and stderr to the same writer while stripping the 8-byte multiplex headers.

**Rationale**:
- Docker multiplexes stdout/stderr with 8-byte headers when `TTY=false` (the default for worker containers).
- Each frame header: `[stream_type:1][padding:3][frame_size:4]` followed by payload.
- `stdcopy.StdCopy(stdout, stderr, reader)` demultiplexes the stream, stripping headers and routing each frame to the correct writer.
- Passing the same writer for both stdout and stderr preserves the current "combined output" behavior while fixing the corruption.
- The package `github.com/docker/docker/pkg/stdcopy` is already available — Docker SDK v28.5.2 is in `go.mod`.

**Alternatives considered**:
1. **Separate stdout/stderr capture**: Would require changing `WorkerConfig.LogWriter` (single writer) and `WorkerResult.Logs` (single string) in the ports package. This is out of scope — a port interface change would require updating all consumers.
2. **Enable TTY=true**: Would disable multiplexing entirely, but TTY enables terminal processing which is undesirable for scripted worker containers (line buffering, terminal escapes, etc.).
3. **Manual header stripping**: Re-implement what stdcopy already does. Unnecessary complexity.

### R2: stopContainer Error Handling & Exit Codes

**Question**: How should `stopContainer` report errors and what constants should be used for exit codes?

**Decision**: Change `stopContainer` to return `(int, error)`. Define named constants for signal-based exit codes. Log `ContainerStop` errors.

**Rationale**:
- Current code returns magic `137` when `ContainerInspect` fails, claiming SIGKILL when the actual cause is unknown.
- Unix convention: exit code = 128 + signal number (137 = SIGKILL, 143 = SIGTERM).
- Named constants improve readability and enable consumer code to match on exit codes.
- The caller (`Runner.Run`) already handles the exit code — it just needs to also handle the error.

**Constants to define** (in `runner.go` or a shared location):
```go
const (
    ExitCodeSIGTERM = 128 + 9   // 137 - Killed by SIGKILL (timeout)
    ExitCodeSIGKILL = 128 + 15  // 143 - Terminated by SIGTERM
)
```
Wait — that's reversed. Correcting:
```go
const (
    ExitCodeSIGKILL = 128 + 9   // 137 - Killed by SIGKILL
    ExitCodeSIGTERM = 128 + 15  // 143 - Terminated by SIGTERM
)
```

**Alternatives considered**:
1. **Keep returning int only, change to -1 for unknown**: Loses error context, callers can't distinguish "inspect failed" from "container exited with -1".
2. **Add error field to WorkerResult**: Would change the port interface. Out of scope.

### R3: streamLogs Error Propagation

**Question**: How should `streamLogs` report errors?

**Decision**: Change `streamLogs` to return `error`. The caller (`Runner.Run`) should log the error but not fail the job — log streaming is best-effort.

**Rationale**:
- Currently errors are silently dropped in two places: `ContainerLogs` failure and `io.Copy`/`stdcopy.StdCopy` failure.
- The goroutine calling `streamLogs` communicates via `logDone` channel. We can add an error channel or return the error and check after `<-logDone`.
- Log streaming failure shouldn't fail the job — the worker may have completed successfully even if log capture failed.
- Error should be logged (not returned as job failure) to aid debugging.

**Implementation approach**:
- Change `streamLogs` signature to return `error`.
- In the goroutine, capture the error and make it available after `<-logDone`.
- Log the error at warning level in `Runner.Run`.

**Alternatives considered**:
1. **Use slog/log package for internal logging**: The runner doesn't currently have a logger. Adding one would require injecting it via constructor. Acceptable but adds to the change surface.
2. **Propagate as job error**: Would cause false job failures when only log capture failed. Wrong semantics.

### R4: Coverage File Organization

**Question**: Where should coverage files be placed and how should the Makefile be updated?

**Decision**: Create a top-level `coverage/` directory. Update all Makefile targets to use `coverage/` prefix. Update `.gitignore` accordingly.

**Rationale**:
- Currently 7+ coverage files land in the repo root: `coverage.out`, `coverage.html`, `coverage.txt`, `coverage.all.out`, `coverage.all.html`, `coverage.all.txt`, `coverage.integration.out`.
- Plus 3 directories: `covdata-unit/`, `covdata-integration/`, `covdata-merged/`.
- Moving everything under `coverage/` keeps the root clean.
- The `.gitignore` already has individual entries — replace them with a single `coverage/` entry.

**Implementation approach**:
- Add `COVERAGE_DIR := coverage` variable at the top of Makefile.
- Prefix all coverage file paths with `$(COVERAGE_DIR)/`.
- Move `covdata-*` directories under `$(COVERAGE_DIR)/` too.
- Update `.gitignore` to just ignore `coverage/`.
- Delete leftover root-level coverage files (not tracked by git, but clean up the working tree).

**Alternatives considered**:
1. **Use `.coverage/` (hidden directory)**: Would hide from `ls` but some tools struggle with hidden directories. Standard Go projects use visible directories.
2. **Use `build/` or `out/`**: Overloaded names. `coverage/` is self-documenting.

### R5: DumpLogs Test Helper Fix

**Question**: Should the test helper fix be identical to the production fix?

**Decision**: Yes — use the same `stdcopy.StdCopy(f, f, rc)` pattern in `DumpLogs`.

**Rationale**:
- The test helper writes container logs to files for debugging. Corrupted logs in test output make failures harder to diagnose.
- Same root cause, same fix. The container log stream has the same 8-byte header format.
- Test infrastructure should match production behavior.

**Alternatives considered**:
1. **Separate stdout/stderr files per container**: Nice-to-have but out of scope for this bug fix. Can be done later if needed.
