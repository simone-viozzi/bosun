# WIP: M3 Code Cleanup Items

**Created**: 2025-12-29
**Priority**: Low — cleanup after MVP functional

## 1. Type Duplication: StepResult / ExecutionResult

**Issue**: Both domain and ports layers define similar types:

| Type | Domain (`jobs/run.go`) | Ports (`ports/executor.go`) |
|------|------------------------|----------------------------|
| `StepResult` | ✓ (lacks StartedAt, Details) | ✓ (used by executor) |
| `ExecutionResult` | ✓ (uses domain StepResult) | ✓ (uses ports StepResult) |

**Current State**:
- `ports.StepResult` and `ports.ExecutionResult` are **actually used** by executor
- `jobs.StepResult` and `jobs.ExecutionResult` in domain layer are **unused**

**Options**:
1. **Remove domain versions** — they're not used, ports versions are sufficient
2. **Consolidate to domain** — move canonical types to domain, have ports re-export or embed
3. **Keep both** — if domain needs pure types without port dependencies (unlikely)

**Recommendation**: Option 1 (remove unused domain types) is simplest. The executor returns `ports.ExecutionResult` which is the contract consumers see.

**Files to modify**:
- `internal/domain/jobs/run.go` — remove `ExecutionResult` and `StepResult` structs
- Keep `JobRun` and `RunStatus` (these ARE used)

## 2. String Literal vs Constant for Status Check

**File**: `internal/cmd/job_run.go` line ~224

```go
// Current:
if step.Status != "success" {

// Should be:
if step.Status != jobs.RunStatusSuccess {
```

**Impact**: Low — works because `RunStatus` is a string type, but less type-safe.

## 3. Timeout Flag Parsing (TODO)

**File**: `internal/cmd/job_run.go` line ~187

```go
// TODO: Parse timeout strings and set overrides
// For M3 MVP, we'll use defaults
```

**Action**: Parse `--timeout`, `--stop-timeout`, `--start-timeout` string flags to `time.Duration` and set on `ExecuteOptions`.

---

*Address these after MVP is functional and tested.*
