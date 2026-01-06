# WIP: Runner.Run Deep Investigation

## Summary

Investigation of `internal/adapters/docker/worker/runner.go` reveals multiple design issues:
1. **Missing stdcopy demultiplexing** — logs are corrupted
2. **Magic exit code 137** — hardcoded assumption on error
3. **Silent error swallowing** — log streaming errors lost
4. **Code duplication** — patterns repeated in compose, testutil
5. **Missing external library usage** — testcontainers-go could simplify testing

---

## Finding 1: Missing stdcopy.StdCopy Demultiplexing (BUG)

**Location:** `runner.go:153-165` — `streamLogs` method

**Current code:**
```go
func (r *Runner) streamLogs(ctx context.Context, containerID string, writer io.Writer) {
    reader, err := r.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
        ShowStdout: true,
        ShowStderr: true,
        Follow:     true,
    })
    if err != nil {
        return // Best effort - ignore errors ← PROBLEM #1
    }
    defer func() { _ = reader.Close() }()

    // Docker multiplexes stdout/stderr with 8-byte header
    // Use stdcopy.StdCopy to demux, or just copy raw for combined output
    _, _ = io.Copy(writer, reader)  // ← PROBLEM #2: Raw copy corrupts output
}
```

**Problem:** Docker container logs are multiplexed with 8-byte headers when TTY=false.
The current `io.Copy` produces **corrupted output** with binary header bytes.

**Evidence from Docker SDK docs:**
> "When the TTY setting is disabled, the stream is multiplexed to separate stdout and stderr.
> Each header is eight bytes containing stream type information and frame size."

**Correct implementation (from Docker SDK examples):**
```go
import "github.com/moby/moby/pkg/stdcopy"

// Demultiplex stdout/stderr properly
stdcopy.StdCopy(os.Stdout, os.Stderr, reader)
```

**Impact:** Users see garbled output with `\x01\x00\x00\x00\x00\x00\x00\x0c` prefixes in logs.

---

## Finding 2: Magic Exit Code 137 on Error (WRONG)

**Location:** `runner.go:130-145` — `stopContainer` method

**Current code:**
```go
func (r *Runner) stopContainer(ctx context.Context, containerID string) int {
    timeout := int(jobs.GracePeriod.Seconds())
    _ = r.docker.ContainerStop(ctx, containerID, container.StopOptions{
        Timeout: &timeout,
    })

    inspect, err := r.docker.ContainerInspect(ctx, containerID)
    if err != nil {
        return 137 // Assume SIGKILL (128 + 9) ← WRONG ASSUMPTION
    }

    return inspect.State.ExitCode
}
```

**Problem:** If `ContainerInspect` fails (network error, container removed, API issue),
returning `137` is misleading — we're claiming the container was killed when we don't know.

**Better approach:**
- Return an explicit error type indicating "exit code unknown"
- Or return -1 with an error to indicate inspection failed
- Log the actual error for debugging

**Fix direction:**
```go
func (r *Runner) stopContainer(ctx context.Context, containerID string) (int, error) {
    // ... stop logic ...

    inspect, err := r.docker.ContainerInspect(ctx, containerID)
    if err != nil {
        return -1, fmt.Errorf("failed to inspect container after stop: %w", err)
    }
    return inspect.State.ExitCode, nil
}
```

---

## Finding 3: Silent Error Swallowing (SMELL)

**Locations:**
- `runner.go:155` — `streamLogs` returns silently on error
- `runner.go:137` — `ContainerStop` error ignored
- `runner.go:165` — `io.Copy` error ignored

**Impact:**
- No way to know if log streaming failed
- No way to know if container stop failed
- Debugging production issues is harder

**Fix direction:** At minimum, log errors even if not returning them.

---

## Finding 4: Code Duplication Across Modules

### Docker client creation
| Location | Pattern |
|----------|---------|
| `internal/adapters/docker/worker/runner.go` | Takes `*client.Client` |
| `internal/adapters/docker/compose/controller.go` | Takes `*client.Client` |
| `internal/adapters/dockerlabels/source.go` | Creates `NewClientWithOpts` |
| `internal/testutil/docker.go` | Creates `NewClientWithOpts` |
| `internal/cmd/job_run.go` | Creates `NewClientWithOpts` twice |

**Fix:** Create `internal/dockerclient/client.go` factory

### ContainerStop pattern
| Location | Pattern |
|----------|---------|
| `worker/runner.go:stopContainer` | `docker.ContainerStop` + inspect |
| `compose/controller.go:StopStack` | `docker.ContainerStop` in loop |

**Observation:** Similar but not identical — compose stops multiple containers with topological order.

### ContainerLogs pattern
| Location | Pattern |
|----------|---------|
| `worker/runner.go:streamLogs` | `ContainerLogs` → `io.Copy` (buggy) |
| `testutil/docker.go:DumpLogs` | `ContainerLogs` → `io.Copy` (also buggy!) |

**Both have the same bug** — missing stdcopy demultiplexing!

---

## Finding 5: External Libraries Could Help

### testcontainers-go (for testing)

**Current state:** Tests use raw Docker Compose CLI + manual container management.

**testcontainers-go benefits:**
- Built-in wait strategies (log, exec, HTTP)
- Automatic cleanup
- Proper log demultiplexing
- Container lifecycle management
- Works reliably in CI

**Example (from docs):**
```go
func TestContainerExec(t *testing.T) {
    ctx := context.Background()
    ubuntuC, err := testcontainers.Run(ctx, "ubuntu:22.04",
        testcontainers.WithCmd([]string{"sleep", "300"}),
        testcontainers.WithWaitStrategy(wait.ForExec([]string{"test", "-f", "/etc/os-release"})),
    )
    testcontainers.CleanupContainer(t, ubuntuC)
    // ... use container ...
}
```

**Recommendation:** Adopt for integration tests (noted TODO in `internal/testutil/harness.go`)

### stdcopy from Docker SDK

**Required fix:** Import and use `github.com/moby/moby/pkg/stdcopy`

```go
import "github.com/docker/docker/pkg/stdcopy"

func (r *Runner) streamLogs(ctx context.Context, containerID string, stdoutWriter, stderrWriter io.Writer) error {
    reader, err := r.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
        ShowStdout: true,
        ShowStderr: true,
        Follow:     true,
    })
    if err != nil {
        return fmt.Errorf("failed to get container logs: %w", err)
    }
    defer reader.Close()

    _, err = stdcopy.StdCopy(stdoutWriter, stderrWriter, reader)
    return err
}
```

---

## Severity Assessment

| Issue | Severity | Type |
|-------|----------|------|
| Missing stdcopy | **HIGH** | Bug — corrupted output |
| Magic 137 | **MEDIUM** | Wrong — misleading exit codes |
| Silent errors | **MEDIUM** | Smell — debugging harder |
| Duplication | **LOW** | Tech debt |
| Missing testcontainers | **LOW** | Tech debt |

---

## Recommended Actions

### Immediate (Bug fixes)
1. Add `stdcopy.StdCopy` to `runner.go:streamLogs`
2. Fix same bug in `testutil/docker.go:DumpLogs`
3. Add named constants for exit codes: `const ExitCodeSIGKILL = 137`

### Short-term (Smell fixes)
4. Change `stopContainer` to return error when inspect fails
5. Log errors in `streamLogs` instead of silently ignoring
6. Add error logging for `ContainerStop` failure

### Medium-term (Tech debt)
7. Create `internal/dockerclient` factory for client creation
8. Evaluate testcontainers-go for integration tests (existing TODO)

---

## Questions for User

- [non-blocking] Should `streamLogs` separate stdout/stderr or combine them?
- [non-blocking] Is returning error from `stopContainer` a breaking change that needs careful handling?
- [non-blocking] Priority for testcontainers-go adoption vs current approach?

---

## Related Smells

- Smell #6: Worker runner silent error swallowing / magic 137
- Smell #32: Worker runner magic exit codes (137) & timeout handling
- Smell #20: Docker client wiring duplicated across CLI/tests
