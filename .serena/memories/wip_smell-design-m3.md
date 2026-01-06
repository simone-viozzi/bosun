## Evidence updates (concrete code pointers)

1) Planner vs Executor mismatch (plan not authoritative)
- Evidence:
  - `internal/app/executor/executor.go`:
    - TODO: "Plan is generated but NOT used to drive execution" (lines 28-36) and subsequent plan generation at `plan, err := e.planner.Plan(ctx, job)` (lines ~44-48). The executor then performs a hardcoded sequence: `e.compose.StopStack(...)` (line ~131), `e.worker.Run(...)` (line ~161), and restarts stack in a `defer` that calls `e.compose.StartStack(...)` (lines ~98-116). It does not iterate over `plan.Steps`.
  - `internal/ports/executor.go` (interface docstring): explicit contract that `Execute` should "3. Execute each step in the plan".
  - `internal/app/planner/planner.go` produces `plan.Steps` (stop, run_worker, start) and sets `ExecutionPlan.CreatedAt` (line ~108).
- Call-sites/Tests referencing Plan:
  - `internal/cmd/job_run.go` (printDryRunText) reads `plan.Steps` for display and will be inconsistent if executor ignores steps.
  - `internal/app/executor/executor_test.go` asserts plan lengths in tests but executor runtime doesn't follow plan.

2) Planner `useCompose` logic is simplistic and duplicated
- Evidence:
  - `internal/app/planner/planner.go`:
    - `useComposeForStop` set to true when `len(job.TargetStacks) == 1` (lines ~54-59). TODO comment: "In future, verify all target containers are in this stack" — indicates current logic is incomplete.
    - Equivalent logic repeated for `useComposeForStart` (lines ~88-97) — duplication with similar decision logic.
  - `pr151-review.md` and `specs/010-m35-cleanup` notes flag the duplicated logic and suggest clearer naming.
- Call-sites:
  - `internal/cmd/plan_show.go` and `internal/cmd/job_run.go` rendering use `step.UseCompose` (e.g., printing "Uses: docker compose stop ...").

3) Rendering of ExecutionPlan lives in CLI
- Evidence:
  - `internal/cmd/job_run.go` defines `printDryRunText(plan jobs.ExecutionPlan)` which renders `plan.Steps` and `step.UseCompose` (lines ~433-472).
  - `internal/cmd/plan_show.go` also inspects `plan.Steps` and prints human-readable output.
- Implication: If plan schema changes (new fields/policies), the CLI must be updated; rendering is currently an ad-hoc responsibility of CLI commands.

4) Executor ignores per-step fields (UseCompose, ContainerIDs)
- Evidence:
  - `internal/app/executor/executor.go` uses `run.StackName` and `e.compose.StopStack(ctx, run.StackName, stopOpts)` (line ~131) and constructs `StepResult` structs using ad-hoc descriptions, but does not read `plan.Steps[*].UseCompose` or `ContainerIDs` to decide whether to use compose vs per-container calls.

5) Tests assume plan-driven execution but runtime deviates
- Evidence:
  - `internal/app/planner/planner_test.go` asserts steps order (stop -> run -> start) (e.g., lines ~60-86), while `internal/app/executor/executor_test.go` checks that `DryRun` returns plan, but there are no tests verifying that `Execute` actually iterates over `plan.Steps` to perform actions.

---

## Immediate remediation suggestions (conceptual)
- Make `ExecutionPlan` authoritative (Plan-is-source-of-truth):
  - Planner must generate complete steps (stop/run/start) including `UseCompose` and `ContainerIDs` that are precise (verify container membership to stacks).
  - Executor should implement a step interpreter loop: `for _, step := range plan.Steps { executeStep(ctx, step, opts) }` where `executeStep` performs behavior based on `step.Type` and `step` fields (e.g., if `StepTypeStopContainers` and `UseCompose==true` => `ComposeController.StopStack`, else stop per container via adapter method).
  - Add per-step policy fields to `jobs.PlanStep` (optional): `Timeout`, `RetryPolicy`, `ContinueOnError`.
- Decide API contract for JobExecutor (Execute-by-name vs Execute-with-Job): choose one and implement consistently (current code implements Execute(ctx, job jobs.Job,...), but ports interface & docs reference name-based Execute). Prefer `Execute(ctx, job jobs.Job, ...)` to keep discovery external (simpler testing and separation of concerns).
- Move rendering to a small `internal/presentation` or `internal/cmd/format` helper package so multiple CLI commands (plan_show, job_run dry-run) share formatting and tests; CLI stays thin and only calls formatting helpers.
- Improve `useCompose` decision: Planner should verify container ownership vs target stacks (e.g., call out to snapshot metadata or require caller to provide stack membership) before setting `UseCompose` to true. Alternatively, add `UseCompose` as best-effort, but document behavior and surface warnings when ambiguous.

---

## Blocking questions to record in the WIP memory (exact phrasing)
- [blocking] Should `ExecutionPlan` be authoritative and drive runtime execution (Plan-is-source-of-truth)? (options: "Yes: Executor interprets plan.Steps" / "No: Plan is preview only and Executor uses fixed internal flow")
- [blocking] Which `JobExecutor` API is canonical? (A) `Execute(ctx, jobName string, opts...)` implemented by Executor (executor performs discovery) or (B) `Execute(ctx, job jobs.Job, opts...)` with discovery performed by caller? (Recommend B)
- [blocking] Should plan rendering remain in the CLI or be moved into a shared presentation package? (options: "Keep in CLI; tests update with schema changes" / "Move to `internal/presentation` and use shared formatters")

---

## WIP: actions added (TODO anchors)
- Added TODO anchor in `internal/app/executor/executor.go` at lines ~28-36 (already present) to implement step interpreter: "Fix: Implement step interpreter loop: for _, step := range plan.Steps { executeStep(step) }".
- Add an issue to implement `executeStep` and extend `PlanStep` with per-step fields (timeouts, retry policy).
- Add an issue to centralize plan rendering to `internal/presentation` for shared usage between `plan_show` and `job_run`.
- Add an issue to refine `useCompose` logic in planner and consider enhancing planner input or snapshot to confirm container→stack mapping.
