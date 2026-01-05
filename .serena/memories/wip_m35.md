# Milestone 3.5: Post-M3 Cleanup & Bug Fixes

**Generated:** 2026-01-05
**Source:** GitHub Issues #138 (parent) + 15 sub-issues
**Repository:** simone-viozzi/bosun

---

## Issue #138

title:	🚀 Milestone 3.5: Post-M3 Cleanup & Bug Fixes
state:	OPEN
author:	simone-viozzi
labels:	prio:P0, type:milestone
comments:	0
assignees:
projects:
milestone:
number:	138
--
# 🚀 Milestone 3.5: Post-M3 Cleanup & Bug Fixes

## Summary

Address critical bugs and incomplete tasks discovered after M3 (Job Execution MVP) completion. This cleanup milestone ensures M3 functionality is production-ready before starting M4 (Scheduling Engine).

## Goal

- **Primary outcome**: All M3 bugs fixed, documentation complete
- **Impact**: Users can reliably use `bosun job run` and understand how to use it

## Scope

### In-scope

- [ ] Fix config validation issues with test compose files (#136)
- [ ] Add missing "restart stack" step to execution plan display (#135)
- [ ] Rename confusing "BackupEnabled" config property (#134)
- [ ] Fix duplicate error messages in CLI output (#133)
- [ ] Complete basic M3 documentation in README (#120)

### Out-of-scope

- M4 features (scheduling, daemon mode)
- M5 features (observability, robustness)
- M6 features (comprehensive docs)

## Success criteria

- [ ] `bosun config validate` works correctly with all test compose files
- [ ] `bosun plan show` displays all execution steps including restart
- [ ] Config validation output uses clear, accurate property names
- [ ] CLI error messages appear only once (no duplicates)
- [ ] README documents basic `bosun job run` usage with examples

## Dependencies (Linked Issues)

- Blocked by:
  - #85 (M3: Job Execution MVP) ✅ completed
- Blocks:
  - #86 (M4: Scheduling Engine & Runtime) — cannot start until cleanup complete

## Deliverables

- [ ] Bug fixes for #133, #134, #135, #136
- [ ] Updated README with job execution documentation (#120)

## Work breakdown (child issues)

**Critical Bugs (P0)**
- [ ] #136 — Config validation fails with test compose
- [ ] #135 — Missing restart step in execution plan display

**High Priority (P1)**
- [ ] #134 — Rename "BackupEnabled" to clearer property name

**Documentation (P3)**
- [ ] #120 — M3 Basic docs update

**Low Priority (P3)**
- [ ] #133 — Fix duplicate error messages

## Risks / unknowns

- **Risk**: Additional bugs may be discovered during fixes
  - Mitigation: Add to this milestone as discovered
- **Unknown**: Documentation scope (basic vs comprehensive)
  - Resolution: Keep basic for M3.5, defer comprehensive to M6

## Definition of done

- [ ] All P0 and P1 issues closed
- [ ] P3 issues closed or explicitly deferred
- [ ] README has "Running Jobs" section with working examples
- [ ] All tests pass (unit + integration)
- [ ] No known critical bugs in M3 functionality

## Notes

- This milestone bridges M3 completion and M4 start
- Focus on making existing M3 features solid, not adding new features
- Once complete, M4 (Scheduling Engine) can begin

---

## Issue #136

title:	validation fails with testutil/compose/joblabels-compose
state:	OPEN
author:	simone-viozzi
labels:	area:testing, bug, prio:P0, type:task
comments:	0
assignees:
projects:
milestone:
number:	136
--
```
❯ docker compose -p joblabels -f internal/testutil/compose/joblabels-compose.yaml up -d
[+] Running 8/8
 ✔ Network joblabels_default               Created                                                                                               0.1s
 ✔ Network joblabels_backend               Created                                                                                               0.1s
 ✔ Volume joblabels_pgdata                 Created                                                                                               0.0s
 ✔ Volume joblabels_redis-data             Created                                                                                               0.0s
 ✔ Container joblabels-postgres-1          Started                                                                                               0.5s
 ✔ Container joblabels-no-job-labels-1     Started                                                                                               0.5s
 ✔ Container joblabels-redis-1             Started                                                                                               0.5s
 ✔ Container joblabels-disabled-service-1  Started                                                                                               0.5s
❯ bin/bosun plan list --project joblabels
NAME          SCHEDULE   CONTAINERS  VOLUMES  STACKS
----          --------   ----------  -------  ------
daily-backup  0 2 * * *  2           2        joblabels
❯ bin/bosun config validate --print

Validation errors:

container "joblabels-no-job-labels-1" (ee026b2b48dc):
  - unknown key: bosun.other

network "joblabels_backend" (8770b1118289):
  - unknown key: bosun.network

Found 2 error(s)
```

---

## Issue #135

title:	Are container restarted after a job that stops them? If yes, why it is not included in the plan?
state:	OPEN
author:	simone-viozzi
labels:	bug, prio:P0, type:bug
comments:	0
assignees:
projects:
milestone:
number:	135
--
```
❯ bin/bosun plan show daily-backup --project joblabels --format json
{
  "jobName": "daily-backup",
  "steps": [
    {
      "type": "stop_containers",
      "description": "Stop stack \"joblabels\" using docker compose stop (2 container(s))",
      "containerIds": [
        "9aa1605cc33dae798ae4e1fcee08abce3c40946695efc9dc8dfee269dbb0e5a1",
        "f7fd17c6d2edd258a3dcd672bf67bd4e1c2f382764023a1b4f848816d555c9dc"
      ],
      "containerNames": [
        "9aa1605cc33d",
        "f7fd17c6d2ed"
      ],
      "useComposeStop": true,
      "composeProject": "joblabels"
    },
    {
      "type": "run_worker",
      "description": "Run worker \"backup-worker:test\" with 2 volumes attached",
      "workerImage": "backup-worker:test",
      "volumeMounts": [
        {
          "name": "joblabels_pgdata",
          "mountPath": "/backup/postgres",
          "mode": "ro"
        },
        {
          "name": "joblabels_redis-data",
          "mountPath": "/backup/redis",
          "mode": "ro"
        }
      ]
    }
  ],
  "createdAt": "2026-01-03T18:27:18.398318345Z"
}
```

Here, the plan stops at run worker, but then the container stopped at the first step are not restarted.

If they are restarted, it should be included in the plan

---

## Issue #134

title:	what is the property "BackupEnabled" in config validation
state:	OPEN
author:	simone-viozzi
labels:	bug, prio:P1, type:bug
comments:	0
assignees:
projects:
milestone:
number:	134
--
```
❯ ./../../../bin/bosun config validate --print
{
  "Instance": "",
  "StopGracePeriod": 30000000000,
  "HealthCheckInterval": 30000000000,
  "AutoRestart": true,
  "LogLevel": "info",
  "BackupEnabled": false,
  "MaxSize": 10737418240,
  "Priority": 100
}
```

Here what is backup enabled? Why there is a reference to backup? is this "job enabled"?

---

## Issue #133

title:	repeated error message in case of unvalid arg
state:	OPEN
author:	simone-viozzi
labels:	bug, prio:P3, type:bug
comments:	0
assignees:
projects:
milestone:
number:	133
--
```
❯ ./bin/bosun completition
Error: unknown command "completition" for "bosun"

Did you mean this?
        completion

Run 'bosun --help' for usage.
2026/01/03 19:18:02 unknown command "completition" for "bosun"

Did you mean this?
        completion
```

---

## Issue #120

title:	Task: M3 Basic docs update
state:	OPEN
author:	simone-viozzi
labels:	area:docs, prio:P3, type:task
comments:	0
assignees:
projects:
milestone:
number:	120
--
# Task: M3 Basic docs update

## Summary

Update README with basic job execution documentation.

## Parent / grouping

- Milestone / Parent issue: #85 (M3: Job Execution MVP)

## Requirements

- [ ] Add "Running Jobs" section to README
- [ ] Document `bosun job run` usage
- [ ] Document `--dry-run` flag
- [ ] Add example compose file with job labels
- [ ] Verify `make docs` still works

## Non-goals

- Comprehensive user guide (M6)
- Example worker images (M6)

## Dependencies (Linked Issues)

- Blocked by:
  - `bosun job run` CLI command
- Blocks:
  - (none)

## Acceptance criteria

- [ ] README has job execution section
- [ ] Examples are accurate and tested
- [ ] No broken links

## Validation / test plan

- How to verify: Follow README instructions on fresh setup
- Edge cases: N/A

## Notes

- Keep brief; comprehensive docs are M6

---

## Issue #139

title:	Fix invalid labels in test compose file (bosun.other, bosun.network)
state:	OPEN
author:	simone-viozzi
labels:	bug, prio:P0, type:task
comments:	1
assignees:
projects:
milestone:
number:	139
--
# Task: Fix invalid labels in test compose file

## Summary

The test compose file `internal/testutil/compose/joblabels-compose.yaml` contains invalid labels (`bosun.other`, `bosun.network`) that cause validation failures. These are test artifacts, not design features.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)
- Related: #136 (validation fails with test compose)

## Requirements

- [ ] Remove `bosun.other: "value"` from `joblabels-compose.yaml:46`
- [ ] Change `bosun.network: "true"` to valid `bosun.network.priority: "50"` at line 68
- [ ] Verify `bosun config validate` passes on all test compose files
- [ ] Update any tests that expect these invalid labels

## Non-goals

- Adding new label namespaces (e.g., `bosun.other.*`)
- Changing validation to allow arbitrary labels

## Evidence

```yaml
# joblabels-compose.yaml
no-job-labels:
  labels:
    bosun.other: "value"    # ← Invalid - not in schema

networks:
  backend:
    labels:
      bosun.network: "true"   # ← Invalid - should be bosun.network.priority
```

## Acceptance criteria

- [ ] `bosun config validate` passes on `joblabels-compose.yaml`
- [ ] All integration tests pass
- [ ] No unknown key errors for test compose files

## Validation / test plan

- How to verify: `bin/bosun config validate --project joblabels` returns no errors
- Edge cases: N/A

## Notes

- This directly fixes #136
- The valid label is `bosun.network.priority` (defined in schema)
author:	simone-viozzi
association:	owner
edited:	false
status:	none
--
## Decision Update (2026-01-04)

**Decision made:** Change validation strategy to **lenient (warning)** mode for unknown `bosun.*` keys.

### Updated Requirements

Instead of removing the invalid labels, we should:

1. **Change loader behavior** (`internal/config/loader/loader.go:125`):
   - Log warnings for unknown `bosun.*` keys instead of errors
   - Optionally: add `--strict` flag for users who want strict validation

2. **Fix test files** (still required):
   - Remove `bosun.other: "value"` (not useful)
   - Change `bosun.network: "true"` → `bosun.network.priority: "50"` (valid field)

### Rationale

- **Graceful degradation**: Users can add custom labels without breaking
- **Rolling deployments**: New labels won't break older bosun versions
- **Extension-friendly**: Other tools may add `bosun.*` prefixed labels
- **Precedent**: Docker, Kubernetes are permissive with unknown labels

See Decision #5 in `wip_smell_milestone3` memory for full context.
--

---

## Issue #140

title:	Rename "backup" terminology to generic "job" terminology
state:	OPEN
author:	simone-viozzi
labels:	bug, prio:P1, type:task
comments:	0
assignees:
projects:
milestone:
number:	140
--
# Task: Rename "backup" terminology to generic "job" terminology

## Summary

Bosun is a **generic job orchestrator**, NOT a backup tool. However, the codebase is littered with "backup" terminology that implies a specific use case. This creates confusion and limits perceived applicability.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Requirements

Production code changes (tests are OK to keep "backup" as example use case):

- [ ] `internal/config/schema/config_v1.go`: Rename `BackupEnabled` → `Enabled` or `JobEnabled`
- [ ] `internal/config/schema/config_v1.go`: Update label `bosun.volume.backupEnabled` → `bosun.volume.enabled`
- [ ] `internal/config/schema/job_labels.go`: Update comments (8 occurrences of "backup")
- [ ] `internal/domain/jobs/run.go`: Fix package doc "backup job execution" → "job execution"
- [ ] `internal/adapters/docker/worker/doc.go`: Fix "backup worker" → "worker" references
- [ ] `internal/cmd/plan_list.go`: Fix "List discovered backup jobs" → "List discovered jobs"
- [ ] `internal/cmd/plan_show.go`: Fix "backup job" → "job"
- [ ] `internal/cmd/job_run.go`: Fix "backup job" references
- [ ] `internal/ports/doc.go`: Fix "backup job definitions" → "job definitions"
- [ ] `docs/config.md`: Update documentation

## Non-goals

- Renaming test strings like `"daily-backup"`, `"backup-job"` — these are valid use case examples

## Acceptance criteria

- [ ] No production code contains "backup" except in example/doc contexts
- [ ] All tests pass
- [ ] `bosun.volume.backupEnabled` label still works (backward compat alias) OR migration documented

## Validation / test plan

- How to verify: `grep -r "backup" internal/ --include="*.go" | grep -v "_test.go"` returns only acceptable occurrences
- Edge cases: Existing compose files with `bosun.volume.backupEnabled` — need migration path

## Notes

- This is a **breaking change** if we remove `backupEnabled` label without alias
- Consider: add alias in loader to accept both old and new label for one release cycle

---

## Issue #141

title:	CLI imports adapters directly (hexagonal architecture violation)
state:	OPEN
author:	simone-viozzi
labels:	prio:P2, type:task
comments:	0
assignees:
projects:
milestone:
number:	141
--
# Task: Fix CLI → Adapter import violation

## Summary

The CLI layer (`internal/cmd/`) directly imports adapter packages, violating hexagonal architecture principles. This increases coupling, duplicates wiring logic across commands, and makes testing harder.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Requirements

- [ ] Move adapter wiring out of CLI into app layer or bootstrap
- [ ] Create app-level factory (e.g., `app.NewExecutorWithDefaults()`)
- [ ] Update CLI commands to use app-level services only
- [ ] Remove direct adapter imports from `internal/cmd/*.go`

## Current violations

```go
// internal/cmd/job_run.go imports:
import (
    "github.com/simone-viozzi/bosun/internal/adapters/docker/compose"
    "github.com/simone-viozzi/bosun/internal/adapters/docker/worker"
    "github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
    "github.com/simone-viozzi/bosun/internal/adapters/joblabels"
)
```

Similar violations in: `plan_list.go`, `plan_show.go`, `validate.go`, `snapshot.go`

## Expected architecture

```
CLI (cmd/) → App (app/) → Ports (ports/) ← Adapters (adapters/)
                              ↑
                           Domain (domain/)
```

CLI should call app-level services, NOT construct adapters directly.

## Non-goals

- Complete dependency injection framework
- Changing adapter interfaces

## Acceptance criteria

- [ ] `internal/cmd/*.go` has no imports from `internal/adapters/*`
- [ ] App layer provides factory/service methods for CLI to use
- [ ] All tests pass
- [ ] CLI behavior unchanged

## Validation / test plan

- How to verify: `grep -r "internal/adapters" internal/cmd/` returns no matches
- Edge cases: Test commands still need adapters via DI

## Notes

- Consider adding import linter rule to prevent regression
- Wiring should live in `internal/app/` or `cmd/bosun/main.go`

---

## Issue #142

title:	ExecutionPlan should be authoritative (executor ignores plan steps)
state:	OPEN
author:	simone-viozzi
labels:	prio:P1, type:task
comments:	0
assignees:
projects:
milestone:
number:	142
--
# Task: Make ExecutionPlan authoritative

## Summary

The current design creates an `ExecutionPlan` but the executor ignores it and hardcodes the execution flow. The plan should drive execution, not just be a preview.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)
- Related: #135 (missing restart step in plan display)

## Current behavior (broken)

```go
// executor.go - ExecuteJob()
plan, _ := e.planner.Plan(ctx, job)  // Plan is created...

// ...but then executor hardcodes the flow:
e.compose.StopStack(...)   // Hardcoded step 1
e.worker.Run(...)          // Hardcoded step 2
defer e.compose.StartStack(...) // Hardcoded step 3 (in defer)
```

The plan is only used for logging, NOT for driving execution.

## Expected behavior

```go
// Executor should interpret plan steps
plan, _ := e.planner.Plan(ctx, job)
for _, step := range plan.Steps {
    result := e.executeStep(ctx, step)
    // record result
}
```

## Requirements

- [ ] Planner adds `StartContainers` step to plan (currently omitted)
- [ ] Executor iterates plan steps instead of hardcoding flow
- [ ] Each step type has a handler in executor
- [ ] Step results are recorded per-step
- [ ] DryRun and Execute use the same plan (consistency)

## Non-goals

- Per-step retry policies (defer to M4)
- Parallel step execution

## Acceptance criteria

- [ ] `bosun plan show <job>` displays start/restart step
- [ ] Executor executes exactly what the plan shows
- [ ] DryRun plan matches actual execution flow
- [ ] All tests pass

## Validation / test plan

- How to verify: `bosun plan show job --format json` includes `start_containers` step
- Edge cases: Plan with zero steps, plan with only worker step

## Notes

- This is a design fix, not just #135 (which is just the display issue)
- Consider: `PlanStep` may need policy metadata (timeouts, retries) in future

---

## Issue #143

title:	Fix JobExecutor interface mismatch (Execute vs ExecuteJob)
state:	OPEN
author:	simone-viozzi
labels:	prio:P1, type:task
comments:	0
assignees:
projects:
milestone:
number:	143
--
# Task: Fix JobExecutor interface mismatch

## Summary

The `ports.JobExecutor` interface defines `Execute(ctx, jobName string, ...)` but the implementation provides `ExecuteJob(ctx, job jobs.Job, ...)`. The interface and implementation don't match.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Current state

```go
// ports/executor.go - Interface expects job lookup by name
type JobExecutor interface {
    Execute(ctx context.Context, jobName string, opts ExecuteOptions) (ExecutionResult, error)
    DryRun(ctx context.Context, jobName string) (ExecutionPlan, error)
}

// app/executor/executor.go - Implementation accepts Job directly
func (e *Executor) ExecuteJob(ctx context.Context, job jobs.Job, ...) { ... }
func (e *Executor) Execute(...) { return error("not implemented") }  // Stub!
```

The `Execute` method returns "not implemented" and constructor accepts unused `discoverer` parameter.

## Recommended fix

Change interface to accept `jobs.Job` (Option B - more modular):

```go
type JobExecutor interface {
    Execute(ctx context.Context, job jobs.Job, opts ExecuteOptions) (ExecutionResult, error)
    DryRun(ctx context.Context, job jobs.Job) (ExecutionPlan, error)
}
```

**Rationale:** Single responsibility - executor executes, discovery is separate.

## Requirements

- [ ] Update `ports.JobExecutor` interface to accept `jobs.Job`
- [ ] Rename `ExecuteJob` → `Execute` in executor
- [ ] Rename `DryRunJob` → `DryRun` in executor
- [ ] Remove unused `discoverer` parameter from `executor.New()`
- [ ] Update all callers

## Non-goals

- Adding job discovery to executor (keep separate)

## Acceptance criteria

- [ ] Interface matches implementation
- [ ] No unused constructor parameters
- [ ] No "not implemented" stub methods
- [ ] All tests pass

## Validation / test plan

- How to verify: `go build ./...` succeeds, interface is satisfied
- Edge cases: N/A

## Notes

- Alternative (Option A): Implement `Execute(jobName)` using injected discoverer
- Team chose Option B for modularity

---

## Issue #144

title:	Worker runner swallows errors and uses magic exit codes
state:	OPEN
author:	simone-viozzi
labels:	prio:P1, type:task
comments:	0
assignees:
projects:
milestone:
number:	144
--
# Task: Fix worker runner error handling

## Summary

The worker runner (`internal/adapters/docker/worker/runner.go`) silently swallows important errors and uses magic numeric exit codes, reducing reliability and debuggability.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Problems

### 1. Silent error swallowing
```go
// streamLogs returns silently on error
func (r *Runner) streamLogs(...) {
    // ...
    if err := scanner.Err(); err != nil {
        return  // ← Error is dropped!
    }
}
```

### 2. Magic exit codes
```go
// stopContainer returns hardcoded 137 on inspect error
func (r *Runner) stopContainer(...) int {
    inspect, err := r.client.ContainerInspect(...)
    if err != nil {
        return 137  // ← Magic number, unclear why 137
    }
}
```

**Note:** 137 = 128 + 9 (SIGKILL), but this should be documented.

## Requirements

- [ ] `streamLogs`: Return or log errors instead of silent return
- [ ] Replace magic `137` with named constant: `const ExitCodeKilled = 137`
- [ ] Document exit code meanings (137 = SIGKILL, 143 = SIGTERM)
- [ ] Add error context (which container, which operation)

## Non-goals

- Adding retry/backoff (defer to separate issue)
- Changing worker lifecycle behavior

## Acceptance criteria

- [ ] No silently dropped errors in worker runner
- [ ] Exit codes use named constants with documentation
- [ ] All tests pass

## Validation / test plan

- How to verify: `grep -r "return$" internal/adapters/docker/worker/` finds no silent returns
- Edge cases: Container stops during log streaming, inspect fails

## Notes

- Related to smell #32 in `wip_smell_milestone3`
- Consider: Should worker errors be retryable? (separate issue)

---

## Issue #145

title:	Improve integration test harness (adopt testcontainers)
state:	OPEN
author:	simone-viozzi
labels:	prio:P2, type:task
comments:	0
assignees:
projects:
milestone:
number:	145
--
# Task: Improve integration test harness reliability

## Summary

Integration tests (`integration/*.go`) are slow, brittle, and have limited isolation because they:
- Build and run the full binary via `exec.Command`
- Call external `docker compose` CLI directly
- Rely on local environment setup
- Have 3-minute timeouts that may still flake in CI

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Current issues

### 1. Heavy E2E approach
```go
// integration/job_execution_test.go
cmd := exec.Command("./bosun", "job", "run", ...)  // Full binary
// Slow, opaque failures
```

### 2. External docker compose dependency
```go
// Tests call docker compose directly
exec.Command("docker", "compose", "-f", ..., "up", "-d")
```

### 3. TODO acknowledges the problem
```go
// internal/testutil/harness.go
// TODO: use testcontainers-go for isolation
```

## Requirements

- [ ] Add unit tests for executor/planner with mocked ports
- [ ] Adopt testcontainers-go for reliable Docker setup
- [ ] Create fast harness that doesn't require binary build
- [ ] Move slow E2E tests to separate suite (run less frequently)
- [ ] Document test environment requirements

## Non-goals

- Removing all E2E tests (some are valuable)
- Mocking Docker API completely

## Acceptance criteria

- [ ] Fast unit tests (<1s) for core services (executor, planner)
- [ ] Integration tests use testcontainers (deterministic)
- [ ] CI runs reliably without manual Docker setup
- [ ] All tests pass

## Validation / test plan

- How to verify: `go test -short ./internal/app/...` completes under 1s
- Edge cases: Tests still work without local docker compose

## Notes

- Related to smells #8, #30 in `wip_smell_milestone3`
- Consider: Separate `make test-unit` and `make test-integration` targets

---

## Issue #146

title:	Refactor large executor and runner methods (complexity reduction)
state:	OPEN
author:	simone-viozzi
labels:	prio:P2, type:task
comments:	0
assignees:
projects:
milestone:
number:	146
--
# Task: Refactor large executor and runner methods

## Summary

Several orchestration methods exceed acceptable complexity thresholds, making them hard to test, reason about, and extend.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Problem methods

### 1. `Executor.ExecuteJob` (~167 lines)
- Responsibilities: plan generation, image validation, stop, worker run, restart
- Many paths through one function
- Difficult to test individual steps

### 2. `Runner.Run` (~103 lines)
- Responsibilities: create, start, wait, logs, cleanup
- Concurrency + lifecycle in one function
- Hard to isolate failure modes

### 3. `ValidateJobLabels` (~159 lines)
- Many validation rules in sequence
- Could be table-driven

## Requirements

**For Executor:**
- [ ] Extract: `prepareRun(job) -> JobRun`
- [ ] Extract: `validateImage(image) -> error`
- [ ] Extract: `executeSteps(plan) -> []StepResult` (prepares for #142)
- [ ] Keep `ExecuteJob` as thin orchestrator

**For Runner:**
- [ ] Extract: `createWorkerContainer(...) -> containerID`
- [ ] Extract: `waitForCompletion(...) -> (exitCode, error)`
- [ ] Extract: `collectLogs(...) -> string`

**For Validator:**
- [ ] Consider: table-driven validation rules

## Non-goals

- Changing execution semantics
- Adding new validation rules

## Acceptance criteria

- [ ] No method exceeds 80 lines (guideline)
- [ ] Each method has single clear responsibility
- [ ] All tests pass
- [ ] Coverage unchanged or improved

## Validation / test plan

- How to verify: `gocyclo -over 15 internal/app/executor internal/adapters/docker/worker` returns clean
- Edge cases: Error handling remains correct

## Notes

- Related to smells #11, #12, #14 in `wip_smell_milestone3`
- This prepares for #142 (plan-driven execution)

---

## Issue #147

title:	Extract CLI output formatting helpers (reduce duplication)
state:	OPEN
author:	simone-viozzi
labels:	prio:P3, type:task
comments:	0
assignees:
projects:
milestone:
number:	147
--
# Task: Extract CLI output formatting helpers

## Summary

CLI commands duplicate JSON/YAML serialization logic and format validation across multiple files, increasing maintenance burden.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Duplication evidence

### Flag setup (repeated 5+ times)
```go
// plan_list.go, plan_show.go, job_run.go, snapshot.go, ...
var format string
cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text|json|yaml)")
// Same validation repeated each time
```

### Output rendering (repeated pattern)
```go
// render*JSONOutput in 4+ files
func renderSomethingJSONOutput(data SomeType) error {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(data)
}

// render*YAMLOutput in 4+ files (nearly identical)
```

## Requirements

- [ ] Create `internal/cmd/output/` package with helpers
- [ ] Extract: `output.RegisterFormatFlag(cmd *cobra.Command)`
- [ ] Extract: `output.Print(format string, data interface{}) error`
- [ ] Update all CLI commands to use shared helpers
- [ ] Add tests for output formatting

## Non-goals

- Changing output format or structure
- Adding new format types (e.g., CSV)

## Acceptance criteria

- [ ] No duplicated `json.NewEncoder` / `yaml.NewEncoder` calls in cmd/
- [ ] Format validation is centralized
- [ ] All commands use shared output helpers
- [ ] All tests pass

## Validation / test plan

- How to verify: `grep -r "json.NewEncoder" internal/cmd/*.go` returns 0 matches (should be in output package)
- Edge cases: Nil slices render as `[]` not `null`

## Notes

- Related to smells #7, #19 in `wip_smell_milestone3`
- Consider: Structured logging/output for machine parsing

---

## Issue #148

title:	Replace magic strings with typed constants (Mode, ContainerState)
state:	OPEN
author:	simone-viozzi
labels:	prio:P3, type:task
comments:	0
assignees:
projects:
milestone:
number:	148
--
# Task: Replace magic strings with typed constants

## Summary

Several places use magic string values for mode, container state, and other enums, increasing risk of typos and reducing type safety.

## Parent / grouping

- Milestone / Parent issue: #138 (M3.5: Post-M3 Cleanup)

## Examples

### 1. Volume mount mode (in domain)
```go
// internal/domain/jobs/types.go
type VolumeAttachment struct {
    Mode string `json:"mode"` // "ro" or "rw" (magic strings)
}

// Used as:
mount.Mode = "ro"  // Could typo as "r0", "or", etc.
```

### 2. Container state (in adapters)
```go
// internal/adapters/docker/compose/controller.go
if cnt.State != "running" {  // Magic string comparison
    // ...
}
```

## Requirements

**For VolumeMode:**
- [ ] Create `type VolumeMode string` in `internal/domain/jobs/`
- [ ] Add constants: `const (ModeReadOnly VolumeMode = "ro"; ModeReadWrite VolumeMode = "rw")`
- [ ] Update `VolumeAttachment.Mode` to use typed field
- [ ] Add validation method if needed

**For ContainerState:**
- [ ] Consider: `type ContainerState string` with constants
- [ ] Or use Docker SDK types if available

## Non-goals

- Changing JSON serialization format
- Adding new modes or states

## Acceptance criteria

- [ ] No raw "ro"/"rw" strings in comparisons
- [ ] No raw "running"/"exited" strings in comparisons
- [ ] Type system prevents invalid values
- [ ] All tests pass

## Validation / test plan

- How to verify: `grep -r '"ro"' internal/ | grep -v _test` returns only const definitions
- Edge cases: JSON marshaling still produces correct output

## Notes

- Related to smell #10 in `wip_smell_milestone3`
- Consider: Enum validation on config load

---
