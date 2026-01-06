# WIP Issue: ExecutionPlan NOT Source of Truth

## Status: FIXED (2026-01-06)

**Resolution**: Implemented step interpreter pattern in executor. Executor now iterates `plan.Steps` and calls `executeStep` for each step type.

## Decision (from wip_smell_milestone3, Decision #4)

**Decided:** "Plan-is-source-of-truth — ExecutionPlan should be authoritative and drive runtime execution."

**Rationale:**
- Single source of truth: planner defines execution sequence, not executor
- DryRun accuracy: "What you preview is what runs"
- Extensibility: per-step policies (retry, timeout, continue-on-error) live in PlanStep
- Follows Command/Interpreter pattern (Terraform, Kubernetes, CI/CD systems)

## Expected Implementation

```
Planner.Plan(job) → ExecutionPlan { Steps: [stop, run_worker, start] }
                          ↓
Executor.Execute(job) → for step := range plan.Steps { executeStep(step) }
```

1. Planner produces complete steps (stop → run_worker → start)
2. Executor implements step interpreter loop
3. Each step type handled by a switch: StopContainers, RunWorker, StartContainers
4. Step metadata (timeout, retry, continue-on-error) respected per-step

## Actual Implementation

```
Planner.Plan(job) → ExecutionPlan { Steps: [stop, run_worker, start] }
                          ↓
Executor.Execute(job) → plan = planner.Plan(job)  // generates plan
                        result.Plan = plan         // stores for display ONLY
                        e.compose.StopStack(...)   // hardcoded stop
                        e.worker.Run(...)          // hardcoded worker
                        defer { e.compose.StartStack(...) }  // hardcoded start
```

1. ✅ Planner produces complete steps (fixed in PR #151)
2. ❌ Executor ignores `plan.Steps` entirely
3. ❌ Executor has hardcoded stop→worker→start sequence
4. ❌ Plan is only used to populate `result.Plan` for display

## Impact

| Concern | Impact |
|---------|--------|
| DryRun divergence | `bosun job run --dry-run` shows plan, but execution may differ |
| Extensibility blocked | Cannot add per-step policies without changing executor |
| Two sources of truth | Planner and executor both define the sequence |
| Testing complexity | Must test planner AND executor separately for same logic |

## Evidence (TODOs added)

- `internal/app/executor/executor.go:57` — Main TODO explaining the gap
- `internal/app/planner/planner.go:53` — useCompose logic is simplistic
- `internal/app/planner/planner.go:87` — Duplicated useCompose decision
- `internal/cmd/job_run.go:418` — Plan rendering in CLI (related coupling)

## Fix Direction

1. Refactor `Executor.Execute` to iterate `plan.Steps`
2. Implement `executeStep(step PlanStep)` with switch on `step.Type`
3. Remove hardcoded `StopStack`/`Run`/`StartStack` calls
4. Optionally extend `PlanStep` with policy fields (Timeout, RetryPolicy, ContinueOnError)

## Related Smells

- Smell #23: Planner vs Executor mismatch (this issue)
- Smell #24: Execution plan incompleteness (partial - start step added but not used)
- Smell #16-18: CLI layering (plan rendering in CLI)

## References

- `wip_smell_milestone3` — Decision #4 and smell tracking
- `wip_smell-design-m3` — Detailed evidence
- Issue #142 — GitHub issue (marked FIXED incorrectly)
