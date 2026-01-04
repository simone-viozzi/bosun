# Scope

- Scope label: milestone3 complexity scan
- Included paths:
  - `internal/app/executor/`
  - `internal/app/planner/`
  - `internal/adapters/docker/compose/`
  - `internal/adapters/docker/worker/`
  - `internal/config/loader/`
  - `internal/config/merge/`
- Diff-only: no (full scan within paths)
- What I inspected: top-level types and functions in files under the listed directories, focusing on long functions (>50 lines), multi-responsibility functions, deep nesting, and heavy use of reflection or stateful orchestration. Files reviewed include:
  - `internal/app/executor/executor.go`
  - `internal/app/planner/planner.go`
  - `internal/adapters/docker/compose/controller.go`, `topology.go`
  - `internal/adapters/docker/worker/runner.go`
  - `internal/config/loader/loader.go`, `parse.go`, `job_validation.go`
  - `internal/config/merge/merge.go`

---

# Findings

## 1) Executor: `ExecuteJob` (internal/app/executor/executor.go)
- Location: `internal/app/executor/executor.go` — `Executor.ExecuteJob`
- Lines: ~167 (func start line 40, next func at 207)
- Evidence:
  - Contains planning (`e.planner.Plan`), image validation, stack stop/start, worker run, step result assembly, deferred cleanup logic, status updates, error handling, and finalization in one function.
  - Multiple early returns interleaved with state updates (run status, result.StepResults), and a non-trivial deferred cleanup block that itself contains branching.
  - Heavy use of domain-specific error wrapping/translation and timeline tracking (StartedAt, CompletedAt).
- Why it's a smell:
  - Single function implements many responsibilities (plan + validate + stop stack + run worker + restart stack + result aggregation) — hard to reason about and test unit-by-unit.
  - Deep nesting and multiple early returns make control flow hard to follow and easy to regress when changing one step.
- Remediation direction:
  - Extract step helpers: e.g., `planJob`, `stopStack`, `runWorker`, `restartStack` (the defer body), and `buildResult` helpers.
  - Introduce a small `step` abstraction or orchestrator that composes steps and returns step results, improving testability and reducing local state coupling.
- Dependencies / assumptions:
  - Current implementation uses `ports` structs and job-run metadata; refactor should preserve state transitions and error semantics.
- Relationship to other findings:
  - Closely related to `Runner.Run` (worker execution) and `compose.Controller` methods (stop/start); extracting steps may allow clearer boundaries with those components.


## 2) Runner: `Run` (internal/adapters/docker/worker/runner.go)
- Location: `internal/adapters/docker/worker/runner.go` — `Runner.Run`
- Lines: ~103 (start line ~31, next method at ~134)
- Evidence:
  - Creates and starts Docker container, sets up log streaming goroutine, waits with timeout, stops container on errors/timeouts, inspects exit code, handles cleanup and conditional removal.
  - Uses concurrency (goroutine for logs) plus synchronous wait/select logic — multiple responsibilities in one method.
- Why it's a smell:
  - Multiple concerns (create/start, log streaming, wait/timeout, stopping, cleanup) are mixed; concurrent interactions and timeouts increase cognitive load and testing complexity.
- Remediation direction:
  - Extract `createContainer`, `startContainer`, `streamLogs` (already a method but could be split further to return channels), and `waitWithTimeout` helpers.
  - Introduce a small state manager or clearer lifecycle flow with explicit phases and error handling policies.
- Dependencies / assumptions:
  - Interacts with Docker client heavily; unit tests will require fakes/mocks; refactor should keep single responsibility per helper for easier mocking.


## 3) Loader: `FromLabels` and reflection helpers (internal/config/loader/loader.go)
- Location: `internal/config/loader/loader.go` — `FromLabels`, `setFieldRecursive`, `setReflectValue`
- Lines: `FromLabels` ~65 lines (start ~108 to ~172); helpers spread across file
- Evidence:
  - `FromLabels` filters labels, validates keys, enforces scope, parses typed values, and uses reflection to set fields in `schema.ConfigV1`.
  - Uses multiple helper functions (`parseValue`, `setField`, reflection recursion) and aggregates multiple validation errors.
- Why it's a smell:
  - High cognitive coupling between parsing, validation and reflection-based mutation logic. Reflection introduces subtlety and hidden failure modes that are hard to test and reason about.
  - Error aggregation means function must coordinate many failure modes, increasing complexity.
- Remediation direction:
  - Separate concerns: (a) parse and validate label-to-typed-value mapping into an intermediate map, (b) map into struct using an explicit, tested mapping layer (or use tag-based unmarshalling libraries if suitable), (c) keep reflection usage minimal and isolated.
  - Add smaller unit-tested helpers and document behaviors around required/optional fields.
- Dependencies / assumptions:
  - Relies on `schema.Spec` and `schema.FieldSpec` semantics; refactor should preserve validation semantics and error types.


## 4) Loader: `ValidateJobLabels` (internal/config/loader/job_validation.go)
- Location: `internal/config/loader/job_validation.go` — `ValidateJobLabels`
- Lines: ~159 (start ~74 to ~232)
- Evidence:
  - Traverses labeled entities, distinguishes containers vs volumes, validates booleans, cron expressions, mount paths, and performs cross-entity conflict detection using maps.
  - Contains many validation rules, conflict detection logic and error/warning aggregation.
- Why it's a smell:
  - Function encapsulates both per-entity validation and cross-entity consistency checks; long function with multiple nested flows reduces clarity and test granularity.
- Remediation direction:
  - Split into multiple passes/components: per-entity validators (container validator, volume validator), and a separate cross-entity consistency pass. Keep the orchestration function small and declarative.
- Dependencies / assumptions:
  - Uses cron parser; validation semantics likely part of public contract — changes should preserve emitted error codes and messages.


## 5) Planner: `Plan` (internal/app/planner/planner.go)
- Location: `internal/app/planner/planner.go` — `Planner.Plan`
- Lines: ~69 (start ~27 to ~96)
- Evidence:
  - Sorts containers and volumes for deterministic output, checks stacks and builds steps with some branching around stack usage and descriptions.
- Why it's a smell (moderate):
  - Function is slightly over the >50 line threshold and mixes deterministic ordering logic with description formatting; splitting step construction into helpers would clarify intent and reduce line count.
- Remediation direction:
  - Extract functions for `buildStopStep`, `buildRunStep`, and deterministic sort helpers; make descriptions easier to test.


## 6) Compose adapter: `topologicalSort` (internal/adapters/docker/compose/topology.go)
- Location: `internal/adapters/docker/compose/topology.go` — `topologicalSort`
- Lines: ~50-ish
- Evidence:
  - Implements DFS-based topological sort with cycle detection and uses recursion / maps to track state.
- Why it's a smell (cautionary):
  - Algorithm is correct but non-trivial; it should be well-covered by unit tests (I found a test file present). If requirements change (external dependencies), semantics may need to be reconsidered.
- Remediation direction:
  - Keep the algorithm isolated, add more unit tests for edge cases (external deps, cycles), and consider iterative implementation if recursion depth is a concern.

---

# Questions for user
- [blocking] Which findings should be prioritized for immediate refactor during Milestone 3? Options: (A) Executor (`ExecuteJob`) only, (B) Runner (`Run`) only, (C) Loader (`FromLabels` + `ValidateJobLabels`), (D) All of the above in priority order you specify.
- [non-blocking] Is API/behavioral change allowed for these refactors (e.g., introducing new small helper functions that alter internal error types), or must all public behavior/messages remain byte-for-byte identical?
- [non-blocking] Are there existing repository conventions/preferences for "step" abstractions/orchestrators or small state machines that should be followed? (If yes, point me to files or memories.)

---

# Confidence / Notes
- Confidence: High for findings and sizes (line counts are sourced from file locations); medium for suggested remediations (implementation details depend on repo conventions and tests).
- I did not exhaustively scan test coverage for each hotspot; next steps ideally include checking unit tests for the identified symbols and expanding coverage where lacking.

---

# Suggested next scout(s)
- Follow-up "Refactor-handoff" scout focusing on `Executor.ExecuteJob` to propose a concrete decomposition and identify tests to add.
- Coverage scout to list existing tests for each hotspot and identify gaps.
- Performance/robustness scout for `Runner.Run` to consider improved timeout/stop semantics and integration tests with Docker client fakes.

---

## META feedback (optional)
- Noted pattern: many orchestration functions follow similar structure (plan/stop/run/start) — consider introducing a lightweight step orchestration helper usable across executor and planner.
