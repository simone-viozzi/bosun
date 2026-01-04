# WIP: Design smell scan — milestone 3

## Scope
- **Scope label:** milestone3 design smells
- **Included paths:**
  - `internal/app/executor/` (execution model)
  - `internal/app/planner/` (planning model)
  - `internal/domain/jobs/` (Job and ExecutionPlan types)
  - `internal/ports/` (interface contracts)
  - `internal/config/schema/` (configuration model)
- **Diff-only vs broader scan:** broader scan within these modules
- **Files inspected (representative):**
  - `internal/app/executor/executor.go`
  - `internal/app/planner/planner.go`
  - `internal/domain/jobs/types.go`
  - `internal/ports/executor.go`, `internal/ports/worker.go`, `internal/ports/compose.go`
  - `internal/config/schema/config_v1.go`
  - Tests: `internal/app/executor/executor_test.go`, `internal/config/schema/config_v1_test.go`

---

## Verification of prior findings

### Prior findings verification
| ID | Title | Status | Notes |
|----|-------|--------|-------|
| 4 | Executor API mismatch / unused `discoverer` param | **Confirmed** | `Executor.New` accepts a `JobDiscoverer` but does not use it; `Execute` returns an error — implementation uses `ExecuteJob` (by-job) instead. |
| 23 | Planner vs Executor mismatch (plan not authoritative) | **Confirmed** | `Planner.Plan` returns an `ExecutionPlan`, but `Executor.ExecuteJob` performs a fixed sequence (stop → worker → restart in defer) rather than interpreting the plan. |
| 24 | Execution plan incompleteness (no start step, no per-step policies) | **Confirmed** | `Planner` explicitly omits a start step (TODO); `PlanStep` lacks per-step policy fields (retry, timeout, restart policy). |
| 25 | Error handling & retry policy under-specified | **Confirmed** | `validateImage` fails fast (no pull), `ComposeController` and `WorkerRunner` have no retry/backoff semantics. |
| 26 | `ports` imports domain types (coupling tradeoff) | **Confirmed (intentional tradeoff)** | `ports.ExecutionResult` and several port types embed `jobs.*` domain types (e.g., `JobRun`, `ExecutionPlan`) — acceptable but increases coupling; consider DTOs if independent evolution desired.

## Findings

### 1) Planner vs Executor mismatch (plan not authoritative) ✅
- **Location(s):** `internal/app/planner/planner.go` (Plan builds steps: stop → run_worker), `internal/app/executor/executor.go` (ExecuteJob directly calls `compose.StopStack`, runs worker, and restarts stack in a deferred cleanup).
- **Evidence:** Planner creates `PlanStep` list but explicitly omits a start/restart step (comment: "Note: Step 3 (start containers) would be added in a future milestone"). Executor ignores `plan.Steps` when performing actions and instead uses a fixed sequence including a deferred restart.
- **Why it’s a smell:** The planner produces a human-readable/deterministic plan but the executor does not use it to drive execution. This duplicates stop/start semantics and risks divergence between preview (DryRun) and actual execution. It also makes it harder to evolve semantics (e.g., adding retries, per-step timeouts, conditional starts).
- **Remediation direction:** Choose one of:
  - Make `ExecutionPlan` authoritative: planner produces full lifecycle steps (including start/restart) and executor implements a step interpreter that runs steps in order (and records per-step results).
  - Or document explicitly that plans are only previews and keep executor’s fixed flow; update tests & docs to make contract explicit.
- **Dependencies:** Changing this affects DryRun semantics, CLI UX, and possibly monitoring/recording formats.

---

### 2) API mismatch: `JobExecutor` vs `Executor` implementation (unused discoverer, Execute-by-name not implemented) ⚠️
- **Location(s):** `internal/ports/executor.go` (interface expects Execute(ctx, jobName string, ...)), `internal/app/executor/executor.go` (constructor receives `discoverer ports.JobDiscoverer` but does not use it; `Execute` method returns an error; public method `ExecuteJob(ctx, job jobs.Job, ...)` exists instead).
- **Evidence:** `Executor.New` signature: `New(discoverer ports.JobDiscoverer, planner ports.JobPlanner, ...)` with comment "Not used directly, kept for compatibility"; `Execute` returns "Execute by name not implemented in M3 - use ExecuteJob".
- **Why it’s a smell:** Interface and implementation diverge, leading to confusion about how clients should invoke execution and where job lookup belongs. The unused `discoverer` parameter is dead-weight and may indicate incomplete design choices.
- **Remediation direction:** Decide one direction and apply consistently:
  - Implement `Execute(ctx, jobName string)` using `JobDiscoverer` inside `Executor` (lookup job, then call the same internal flow), or
  - Change the `JobExecutor` interface to accept `jobs.Job` (or a typed DTO), and remove the unused discoverer from `Executor.New`.
- **Blocking question:** Which contract do maintainers prefer (Execute-by-name with discoverer inside executor vs Execute-with-Job and discoverer usage externalized)?

---

### 3) `BackupEnabled` naming is confusing (config-level API/compatibility risk) ⚠️
- **Location(s):** `internal/config/schema/config_v1.go` (`VolumeConfig.BackupEnabled`), docs (`docs/config.md`) and tests use `bosun.volume.backupEnabled`.
- **Evidence:** Existing repo memories note this confusion; the schema, tests and docs already use `backupEnabled` label/key.
- **Why it’s a smell:** Ambiguous or inconsistent naming can cause UX confusion; renaming is a breaking change across labels, tests, and docs. Conversely keeping a slightly awkward name perpetuates the UX debt.
- **Remediation direction:** Options:
  - Keep `BackupEnabled` but improve docs and add an alias handling (accept `backupsEnabled`/`volumeBackupsEnabled`) in the loader for a deprecation path, or
  - Rename to `BackupsEnabled` / `VolumeBackupsEnabled` and add a migration/deprecation path (breaking change, requires spec/docs update).
- **Question:** Is a breaking rename acceptable now, or prefer non-breaking approach (alias + docs)?

---

### 4) Execution plan incompleteness and step model limitations (start/retry/policies) 💡
- **Location(s):** `internal/app/planner/planner.go`, `internal/domain/jobs/types.go` (StepTypeStartContainers exists but not used by planner), `internal/app/executor/executor.go` (restart in defer, not in plan execution).
- **Evidence:** Planner's TODO mentions adding start step in future milestone. PlanStep currently contains descriptive fields but lacks per-step policies (timeouts, retries, conditional semantics).
- **Why it’s a smell:** The step model is under-specified for production concerns (retries, per-step timeout, conditional restart policies). As features grow, adding per-step metadata later may be painful if plan and executor semantics have diverged.
- **Remediation direction:** Extend PlanStep with optional policy fields (RetryPolicy, Timeout, RestartPolicy) and ensure planner and executor have a shared interpretation.

---

### 5) Executor complexity and single large method (`ExecuteJob`) → maintainability risk ✅
- **Location(s):** `internal/app/executor/executor.go` (ExecuteJob ~160 lines)
- **Evidence:** `ExecuteJob` builds run/result objects, handles plan generation, image validation, stop, run worker, and deferred restart — many responsibilities in one function.
- **Why it’s a smell:** Long methods that orchestrate many steps are harder to test, reason about, and extend. Many tests mock components but internal logic still bundles policy and error-handling.
- **Remediation direction:** Refactor to smaller functions (e.g., `prepareRun`, `executeSteps(plan)`, `stopStack`, `runWorker`, `restartIfNeeded`, `recordStepResult`) or implement a step interpreter that executes `jobs.ExecutionPlan`.

---

### 6) Error handling & retry policy is under-specified (no retries, limited context handling) ⚠️
- **Location(s):** `internal/app/executor/executor.go` (`validateImage` fails fast, stop/start have no retries), `internal/ports/compose.go` (Start/Stop options limited to timeouts)
- **Evidence:** `validateImage` returns a wrapped ErrImageNotFound and comment "pull not implemented in M3"; start/stop behavior has conservative defaults and no retry/backoff.
- **Why it’s a smell:** Real-world operations (docker pulls, transient stop/start failures) often need retry/backoff policies and clearer semantics around what is fatal vs recoverable.
- **Remediation direction:** Define acceptable retry/backoff policies at the port/plan level (e.g., allow planner to attach per-step retry metadata; add optional pull behavior for images, or document as explicit limitation).

---

### 7) Coupling: `ports` references domain types (tradeoff) ⚠️
- **Location(s):** `internal/ports/executor.go` (ExecutionResult uses `jobs.JobRun`, `jobs.ExecutionPlan`), `internal/ports/worker.go` references domain-like structs.
- **Why it’s a smell:** Ports importing domain types increases coupling and can complicate independent evolution of domain and ports. This is not necessarily wrong (hexagonal architecture often uses domain types), but ask whether the team prefers to keep ports free of domain imports (use DTOs) or accept the coupling.
- **Remediation direction:** If decoupling desired, introduce lightweight DTOs in `ports` to avoid importing `internal/domain`.

---

## New findings (continuing numbering from existing findings)

### #8 — Plan CreatedAt responsibility mismatch (planner vs caller) 🔧
- **Location(s):** `internal/app/planner/planner.go` (`Plan` sets `CreatedAt`), `internal/domain/jobs/types.go` (comment: "CreatedAt records when the plan was generated. Set by the caller, not the planner").
- **Evidence:** Planner sets `CreatedAt: time.Now().UTC()` during `Plan()`; domain docs state caller should set `CreatedAt` for determinism.
- **Why it’s a smell:** Inconsistent responsibility can lead to surprising test behavior and makes it unclear which component owns plan metadata (affects caching, deterministic testing).
- **Remediation direction:** Choose a single owner: (A) Planner sets `CreatedAt` (remove note in domain docs), or (B) Planner returns plan with zero `CreatedAt` and caller sets it; update tests and DryRun path accordingly.
- **Questions:** [non-blocking] Which approach is preferred for determinism and testing: planner-owned timestamp or caller-owned timestamp? (Options: "Planner sets CreatedAt", "Caller sets CreatedAt (preferred for determinism)")
- **Confidence:** High

---

### #9 — Duplicate/overlapping ExecutionResult & StepResult types across `ports` and `domain` 🔗
- **Location(s):** `internal/ports/executor.go` (`ExecutionResult`, `StepResult`), `internal/domain/jobs/run.go` (`ExecutionResult`, `StepResult`).
- **Evidence:** Both packages define similarly named structs representing execution results; `ports.ExecutionResult` embeds `jobs.JobRun` and `jobs.ExecutionPlan`, while `domain.ExecutionResult` has its own `JobRun` and `ExecutionResult` shapes.
- **Why it’s a smell:** Duplicate types can cause confusion about ownership, conversion code, and can lead to accidental drift (two places to change a field). It also complicates API contracts for adapters and consumers.
- **Remediation direction:** Consolidate on a single representation:
  - Option A: Keep domain `ExecutionResult` as source of truth and have `ports` reference domain types (accept coupling), or
  - Option B: Introduce DTOs in `ports` and keep domain types internal; add conversion helpers.
- **Questions:** [non-blocking] Which model does the team prefer: single domain type (accept coupling) or ports-DTOs (decouple)?
- **Confidence:** High

---

### #10 — CLI layering & wiring (business logic in CLI + adapter imports) ⚠️
- **Location(s):** `internal/cmd/job_run.go` — imports `adapters/docker/compose`, `adapters/docker/worker`, `adapters/dockerlabels`, `adapters/joblabels`, creates docker client and performs discovery; constructs `executor` and calls `ExecuteJob`.
- **Evidence:** CLI creates adapter instances and performs job discovery + selection logic; it also builds execute options and parses flags into execution behavior.
- **Why it’s a smell:** CLI includes orchestration & discovery logic (business concerns) and directly depends on adapter implementations, increasing coupling and making the CLI harder to unit test in isolation.
- **Remediation direction:** Move wiring and discovery into an app-level factory or bootstrap package (`internal/bootstrap` or `internal/app/factory`), expose a thin CLI that only converts args and calls app services.
- **Questions:** [blocking] Should CLI be strictly thin (only argument parsing + presentation) with wiring moved out? (Yes / No)
- **Best-practice claim:** Prefer thin CLI (UNVERIFIED)
- **Confidence:** High

---

### #11 — Config merge semantics treat false as zero (surprising semantics) ⚠️
- **Location(s):** `internal/config/merge/merge.go` (function `isZeroValue` special-cases bool to treat false as zero).
- **Evidence:** `isZeroValue` returns `!v.Bool()` for bools; comment notes limitation and suggests using defaults to indicate "not set".
- **Why it’s a smell:** Users cannot distinguish between "explicitly set to false" and "not set"; this makes disabling features via labels non-intuitive and can cause subtle bugs.
- **Remediation direction:** Use pointer-typed fields for optional booleans or maintain explicit "set" metadata (e.g., a parallel map of set keys), or at minimum document the limitation prominently and provide an opt-in strict merge mode.
- **Questions:** [non-blocking] Is changing the schema to use pointer bools acceptable now? (Yes / Defer)
- **Confidence:** High

---

### #12 — Loader rejects unknown keys strictly (validation policy decision) ⚠️
- **Location(s):** `internal/config/loader/loader.go` (`FromLabels`) — `errs.AddUnknownKey(key, scope)` when a label is not present in spec.
- **Evidence:** Unknown label keys are treated as validation errors and returned to callers as `ValidationErrors`.
- **Why it’s a smell:** Strict rejection can make rolling deployments and extensions brittle if labels are added by other tools; sometimes a warning would be preferable.
- **Remediation direction:** Add a configurable mode: `strict` (error on unknown keys) vs `lenient` (collect warnings), or provide a discovery-only mode that collects unknown keys for a telemetry/diagnostic view.
- **Questions:** [non-blocking] Preferred default behavior: strict (current) or lenient (warn)?
- **Confidence:** Medium

---

### #13 — Integration tests are heavy and brittle; limited isolation 🧪
- **Location(s):** `integration/*.go` (e.g., `integration/job_execution_test.go`) — tests build the binary, run compiled CLI, and depend on `docker compose` externally.
- **Evidence:** Tests use `exec.Command` to run the built `bosun` binary, and call `docker compose` directly in assertions (system-level dependencies). Many tests are 3-minute timeouts.
- **Why it’s a smell:** Slow, brittle E2E tests slow developer feedback and are fragile in CI environments. They also provide limited unit-level coverage for internal services (planner, executor) where mocking ports would be faster and more deterministic.
- **Remediation direction:** Add unit tests that mock ports interfaces (compose/worker/discovery), adopt testcontainers or a harness that can be run in CI reliably, and move slow tests to a separate integration suite that runs less frequently.
- **Questions:** [non-blocking] Agree to add more unit-level tests and a fast harness? (Yes / Defer)
- **Confidence:** High

---

### #14 — Adapter error typing and domain coupling (compose returns `jobs.StopError`) ⚠️
- **Location(s):** `internal/adapters/docker/compose/controller.go` — returns `&jobs.StopError{...}` and `&jobs.StartError{...}`; `internal/adapters/docker/worker/runner.go` returns generic errors on create/start.
- **Evidence:** Adapter wraps low-level errors into domain-specific error structs from `internal/domain/jobs` which couple adapters to domain error types.
- **Why it’s a smell:** Adapters referencing domain error types increases coupling and can make adapter reuse harder; it also leaks domain semantics into lower layers.
- **Remediation direction:** Return port-level errors (e.g., `ports.StopError` / error constants) and let app/domain layer map them into domain error types if needed.
- **Questions:** [non-blocking] Prefer adapter -> ports errors, or is domain-level error returning acceptable here? (Options: "ports errors", "domain errors OK")
- **Confidence:** Medium

---

### #15 — Worker runner's timeouts/stop behavior and magic exit codes (clarity & robustness) ⚠️
- **Location(s):** `internal/adapters/docker/worker/runner.go` (`stopContainer` returns 137 on inspect error; timedOut semantics set based on errCh path)
- **Evidence:** `stopContainer` returns `137` when ContainerInspect fails, and `timedOut` is set based on an `errCh` read (which may represent timeout or other errors). No retry or backoff is attempted.
- **Why it’s a smell:** Magic numeric exit codes and ambiguous timeout handling reduce clarity and make it harder to map errors to actionable remediation steps; missing retry behavior reduces robustness to transient Docker API issues.
- **Remediation direction:** Return explicit error types (or wrap codes), document 137/143 semantics clearly, and consider retry/backoff for transient Docker errors.
- **Questions:** [non-blocking] Should we introduce a small retry/backoff for container operations (start/stop) in adapters? (Yes / No)
- **Confidence:** Medium


## Questions for user
- [blocking] Should `ExecutionPlan` be authoritative and drive execution (i.e., make the Executor execute plan steps) or remain a preview (DryRun only) while Executor uses a fixed internal flow? (options: "Plan-is-source-of-truth", "Plan-is-preview and document it")
- [blocking] Which API contract should be canonical for job execution?
  - Option A: `JobExecutor.Execute(ctx, jobName string, ...)` implemented by `Executor` using an injected `JobDiscoverer` (executor performs discovery + execution).
  - Option B: `JobExecutor` accepts `jobs.Job` directly (keep `ExecuteJob`) and discovery remains external to callers.
- [blocking] `BackupEnabled` rename: do we accept a breaking rename now (`BackupsEnabled` or `VolumeBackupsEnabled`) or prefer to keep the current key and add documentation/aliasing for clarity? (options: "Rename now (breaking)", "Keep + alias/docs")
- [non-blocking] Do we want per-step policy metadata in `PlanStep` (timeouts, retry counts, restart behavior) to be part of M4 planning, or postpone until usage patterns demand? (options: "Add now", "Defer")
- [non-blocking] Should `Executor` attempt to pull worker images on demand (instead of failing fast) or should image pulls remain out-of-scope for now?
- [non-blocking] Is the team comfortable with `ports` importing `internal/domain` types, or should we introduce DTOs in `ports` to reduce coupling?

---

## Confidence / Notes
- Findings are based on direct inspection of planner, executor, ports, domain types, schema and tests in the M3-specified packages.
- Some decisions may be intentionally "MVP by design" (e.g., missing start step annotated as future work). Where behavior is deliberate, I marked it as a question rather than a hard issue.
- I did not exhaustively scan every adapter; I focused on the orchestration surface and contracts as requested.

---

## Suggested next scout / follow-ups

- Implement a small follow-up scan to trace how `ExecutionPlan` is used by callers and whether any external consumers assume the plan contains start steps (impact analysis for adding start step).
- If the team prefers plan-driven execution, a focused scout to identify all assumptions in tests and CLI that would need updating.

---

## Progress log
- **Start:** Opened prior smell memories and confirmed scope (read `wip_smell_milestone3` and `wip_smell-design-m3`). Next step: scan code for concrete evidence.
- **Mid:** Scanned planner, executor, ports, adapters, CLI and config; confirmed prior smells (#4, #23, #24, #25, #26) and collected evidence for new findings (#8–#15).
- **End:** Wrote verification table, appended new findings, and recorded questions and remediation directions; awaiting answers to blocking questions to prioritize follow-ups.

---

- Implement a small follow-up scan to trace how `ExecutionPlan` is used by callers and whether any external consumers assume the plan contains start steps (impact analysis for adding start step).
- If the team prefers plan-driven execution, a focused scout to identify all assumptions in tests and CLI that would need updating.

---

## META feedback (delete once stable)
- What happened: The planner and executor grew two different "contracts" for execution behavior (preview vs runtime), causing duplication.
- Why it’s a problem: Increases maintenance surface and causes subtle divergence in user expectations (dry-run vs actual run).
- Proposed change: Add explicit statement in the spec describing whether plans are authoritative, and add tests exercising plan-driven execution if chosen.
