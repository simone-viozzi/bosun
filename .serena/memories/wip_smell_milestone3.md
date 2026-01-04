# WIP Smell Index: milestone3

## Scope Definition

**Label:** `milestone3`

**What's included:**
| Area | Paths | Rationale |
|------|-------|-----------|
| Executor | `internal/app/executor/` | Core M3 feature — job execution orchestration |
| Planner | `internal/app/planner/` | Generates execution plans (related to #135) |
| Docker adapters | `internal/adapters/docker/compose/`, `internal/adapters/docker/worker/` | Compose stop/start, worker runner |
| Config validation | `internal/adapters/dockerlabels/`, `internal/config/` | Validation issues (#136, #134) |
| CLI | `internal/cmd/` | Duplicate errors (#133), job run command |
| Domain | `internal/domain/jobs/` | ExecutionPlan, PlanStep types |
| Ports | `internal/ports/` | Interface contracts |
| Integration tests | `integration/` | Test quality affects M3 reliability |
| Test utilities | `internal/testutil/` | Compose files triggered #136 |

**What's excluded:**
- `specs/` — spec documents, not code
- `docs/` — non-code documentation
- `.github/` — CI/workflows, not application code
- `cmd/bosun/main.go` — trivial entrypoint

**Special directives:**
- Question "by design" patterns — if something looks intentional but smells off, flag it
- Known issues to cross-reference: #133, #134, #135, #136, #138

---

## Smell List

_Append-only. Never renumber IDs._

| ID | Title | Status | Evidence | TODO Anchors | Issue |
|----|-------|--------|----------|--------------|-------|
| 1 | Confusing `BackupEnabled` property name | **Ready-For-Issue** | `wip_smell-general-m3` #1 | `internal/config/schema/config_v1.go:30` | #140 |
| 2 | Duplicate CLI error printing | New | `wip_smell-general-m3` #2 | — | — |
| 3 | Strict unknown-key rejection (may cause #136) | **Ready-For-Issue** | `wip_smell-general-m3` #3 | `internal/config/loader/loader.go:125` | #139 |
| 4 | Executor API mismatch / unused `discoverer` param | **Ready-For-Issue** | `wip_smell-general-m3` #4, `wip_smell-design-m3` #2 | `internal/app/executor/executor.go:24` | #143 |
| 5 | Missing restart step in plan display (#135) | New | `wip_smell-general-m3` #5 | — | — |
| 6 | Worker runner silent error swallowing / magic 137 | New | `wip_smell-general-m3` #6 | — | — |
| 7 | JSON/YAML output duplication across CLI commands | New | `wip_smell-general-m3` #7 | — | — |
| 8 | Brittle integration test harness (needs testcontainers) | New | `wip_smell-general-m3` #8 | — | — |
| 9 | TODOs without issue links | New | `wip_smell-general-m3` #9 | — | — |
| 10 | Magic strings for Mode (`ro`/`rw`) and container state | New | `wip_smell-general-m3` #10 | — | — |
| 11 | `Executor.ExecuteJob` too large (~167 lines, many responsibilities) | New | `wip_smell-complexity-m3` #1 | — | — |
| 12 | `Runner.Run` too complex (~103 lines, concurrency + lifecycle) | New | `wip_smell-complexity-m3` #2 | — | — |
| 13 | `FromLabels` + reflection helpers complex | New | `wip_smell-complexity-m3` #3 | — | — |
| 14 | `ValidateJobLabels` too large (~159 lines) | New | `wip_smell-complexity-m3` #4 | — | — |
| 15 | `Planner.Plan` slightly over threshold (~69 lines) | New | `wip_smell-complexity-m3` #5 | — | — |
| 16 | CLI imports adapters directly (layering violation) | **Ready-For-Issue** | `wip_smell-layering-m3` #1, `wip_smell-design-m3` #10 | `internal/cmd/job_run.go:1` | #141 |
| 17 | Business logic in CLI handlers (discovery, selection) | **Ready-For-Issue** | `wip_smell-layering-m3` #2, `wip_smell-design-m3` #10 | `internal/cmd/job_run.go:1` | #141 |
| 18 | CLI performs low-level Docker client wiring | **Ready-For-Issue** | `wip_smell-layering-m3` #3, `wip_smell-design-m3` #10 | `internal/cmd/job_run.go:1` | #141 |
| 19 | Duplicated CLI flag setup & format validation | New | `wip_smell-duplication-m3` #2 | — | — |
| 20 | Docker client wiring duplicated across CLI/tests | New | `wip_smell-duplication-m3` #3 | — | — |
| 21 | Near-identical test setup calls | New | `wip_smell-duplication-m3` #5 | — | — |
| 22 | Duplicate discovery/wiring sequences (snapshot→discover) | New | `wip_smell-duplication-m3` #6 | — | — |
| 23 | Planner vs Executor mismatch (plan not authoritative) | **Ready-For-Issue** | `wip_smell-design-m3` #1 | `internal/app/planner/planner.go:84` | #142 |
| 24 | Execution plan incompleteness (no start step, no per-step policies) | **Ready-For-Issue** | `wip_smell-design-m3` #4 | `internal/app/planner/planner.go:84` | #142 |
| 25 | Error handling & retry policy under-specified | Confirmed | `wip_smell-design-m3` #6 | — | — |
| 26 | `ports` imports domain types (coupling tradeoff) | Confirmed | `wip_smell-design-m3` #7 | — | — |
| 27 | Plan CreatedAt responsibility mismatch (planner vs caller) | New | `wip_smell-design-m3` #8 | — | — |
| 28 | Duplicate ExecutionResult & StepResult types (ports vs domain) | New | `wip_smell-design-m3` #9 | — | — |
| 29 | Config merge treats false as zero (surprising semantics) | New | `wip_smell-design-m3` #11 | — | — |
| 30 | Integration tests heavy/brittle; limited isolation | New | `wip_smell-design-m3` #13 | — | — |
| 31 | Adapter error typing (compose returns domain errors) | New | `wip_smell-design-m3` #14 | — | — |
| 32 | Worker runner magic exit codes (137) & timeout handling | New | `wip_smell-design-m3` #15 | — | — |

---

## Open Questions (Blocking)

1. ~~**Smell 3**: Unknown label key handling~~ → **DECIDED: Option B (lenient/warning)** — see Decisions section
2. ~~**Smell 16–18** (#141): Should CLI be thin?~~ → **DECIDED: Yes (thin CLI)** — see Decisions section
3. ~~**Smell 23–24** (#142): Should `ExecutionPlan` be authoritative?~~ → **DECIDED: Plan-is-source-of-truth** — see Decisions section
4. ~~**Smell 4** (#143): Which API contract for job execution?~~ → **DECIDED: Option (B)** — see Decisions section
5. ~~**Smell 1** (#140): Breaking rename of `BackupEnabled` now?~~ → **DECIDED: Rename now (breaking)** — see Decisions section

## Open Questions (Non-Blocking)

- **Smell 1**: Preferred rename for `BackupEnabled`? Options: `BackupsEnabled`, `VolumeBackupsEnabled`, `JobEnabled`, or keep + add docs.
- **Smell 2**: Structured (machine-friendly) vs free-form CLI error messages?
- **Smell 11**: Which complexity hotspot to refactor first? (A) Executor, (B) Runner, (C) Loader, (D) All.
- **Smell 24**: Do we want per-step policy metadata in `PlanStep` (timeouts, retry counts, restart behavior) M4 or defer?
- **Smell 25**: Should `Executor` attempt to pull worker images on demand (instead of failing fast)?
- **Smell 26, 28**: Should `ports` import `internal/domain` types (coupling), or introduce DTOs to decouple?
- **Smell 27**: Which component owns `Plan.CreatedAt`? Planner (current) or caller (for determinism)?
- **Smell 29**: Is changing schema to use pointer bools for optional fields acceptable?
- **Smell 30**: Agree to add more unit-level tests and a fast harness (testcontainers)?
- **Smell 31**: Adapter error types: return ports-level errors or domain errors (current)?
- **Smell 32**: Introduce small retry/backoff for container operations in adapters?

---

## Decisions

### Decision #1: Thin CLI (Smell #16–18, Issue #141)
**Decision:** Yes — CLI should be thin (only parse/present), wiring/discovery moves to `internal/app` factory or `internal/bootstrap`.

**Rationale (best practice confirmed):**
- Hexagonal architecture: CLI is a driving adapter, should not contain business logic
- *"Controller layer shouldn't contain business logic"* (Clean Architecture)
- *"Hexagonal architecture divides a system into loosely-coupled components... user interface code [should not contain] business logic"* (Wikipedia)
- Enables unit testing CLI without Docker, reusable services for future entrypoints (API, scheduler)

**Action:** Refactor CLI to call app-layer services; move discovery + wiring to `internal/app/factory.go` or `internal/bootstrap/`.

---

### Decision #2: ExecuteJob API Contract (Smell #4, Issue #143)
**Decision:** Option (B) — `ExecuteJob(ctx, job)` with discovery external.

**Rationale (best practice confirmed):**
- Single Responsibility: Executor executes, Discoverer discovers
- Testability: Unit test executor with any `jobs.Job` without discoverer mock
- Explicit Dependencies: Caller controls what job is executed
- *"Keep dependencies explicit through constructors"* (DI best practices)

**Action:** Remove unused `discoverer` param from `Executor.New()`, remove stub `Execute(ctx, jobName)` method, update `JobExecutor` interface to match.

---

### Decision #3: BackupEnabled Rename (Smell #1, Issue #140)
**Decision:** Rename now (breaking change acceptable).

**Rationale:**
- No backwards compatibility requirement in current version
- Clean up naming debt early, before users adopt the confusing name
- Rename to `JobEnabled` or similar (TBD in implementation)

**Action:** Rename `BackupEnabled` across schema, labels, docs, and tests.

---

### Decision #5: Unknown Label Key Handling (Smell #3, Issue #139)
**Decision:** Option B — Lenient (warning), don't fail on unknown `bosun.*` keys.

**Rationale:**
- **Graceful degradation**: Users can add custom labels without breaking validation
- **Rolling deployments**: New labels in newer versions won't break older bosun
- **Extension-friendly**: Other tools may add `bosun.*` prefixed labels
- **Precedent**: Docker, Kubernetes labels are permissive (unknown labels ignored)

**Context:**
- `bosun.network.priority` is a *valid* schema field
- `bosun.network` (without `.priority`) and `bosun.other` in test files are *not* in schema → currently rejected
- These test labels likely added accidentally or for other purposes

**Action:**
1. Change loader to log warnings for unknown `bosun.*` keys instead of errors
2. Optionally: add `--strict` flag/config for users who want strict validation
3. Fix test compose files to either remove invalid labels or use valid ones

---

### Decision #4: ExecutionPlan Authority (Smell #23–24, Issue #142)
**Decision:** Plan-is-source-of-truth — `ExecutionPlan` should be authoritative and drive runtime execution.

**Rationale (best practice confirmed):**
- **Single source of truth**: One place (planner) defines the execution sequence, not two (planner + executor)
- **DryRun accuracy**: "What you preview is what runs" — no divergence between DryRun and actual execution
- **Extensibility**: Per-step policies (retry, timeout, continue-on-error) live in `PlanStep`, executor interprets them
- **Testability**: Test planner for correctness; executor is a generic step interpreter
- **Follows Command/Interpreter pattern**: Used by Terraform (plan→apply), DB query planners, CI/CD systems, Kubernetes

**Implementation direction:**
1. Planner produces *complete* steps including start/restart step (remove TODO)
2. Executor implements a step interpreter loop over `plan.Steps`
3. Extend `PlanStep` with optional policy fields: `RetryPolicy`, `Timeout`, `ContinueOnError`
4. Remove hardcoded sequence from `ExecuteJob`; replace with `for step := range plan.Steps { executeStep(step) }`

---

## Session Log

- **2026-01-04**: Scope confirmed with user. Spawning initial scouts.
- **2026-01-04**: 3 scouts completed (general, complexity, layering). Consolidated 18 smells into index.
- **2026-01-04**: 2 more scouts completed (duplication, design). Total 26 smells. 5 blocking questions identified.
- **2026-01-04**: Created 5 GitHub issues from smell findings, attached as sub-issues to #138:
  - #139 — Fix invalid labels in test compose (P0) ← fixes #136
  - #140 — Rename "backup" terminology to "job" (P1)
  - #141 — CLI imports adapters directly (P2) — covers smells #16, #17, #18
  - #142 — ExecutionPlan should be authoritative (P1) — covers smells #23, #24
  - #143 — Fix JobExecutor interface mismatch (P1) — covers smell #4
- **2026-01-04**: Expanded design scouting to full project. Design scout verified prior smells (#4, #23, #24, #25, #26) as **Confirmed**. Added 6 new design smells (#27–#32). Total: **32 smells** tracked.
- **2026-01-04**: Recorded 5 decisions based on user input and best practice research:
  - Decision #1: Thin CLI (smells #16–18, #141) — Yes, move wiring to app layer
  - Decision #2: ExecuteJob API (smell #4, #143) — Option B, discovery external
  - Decision #3: BackupEnabled rename (smell #1, #140) — Rename now (breaking OK)
  - Decision #4: ExecutionPlan authority (smells #23–24, #142) — Plan-is-source-of-truth
  - Decision #5: Unknown label keys (smell #3, #139) — Lenient/warning, not error
- **2026-01-04**: Added TODO anchors to code for all 5 decided smells:
  - `internal/config/schema/config_v1.go:30` — #140 (BackupEnabled rename)
  - `internal/config/loader/loader.go:125` — #139 (unknown keys)
  - `internal/app/executor/executor.go:24` — #143 (executor API)
  - `internal/cmd/job_run.go:1` — #141 (CLI layering)
  - `internal/app/planner/planner.go:84` — #142 (plan authority)

---

## Issue Review Summary

### Existing Issues (Ready for Implementation)

| Issue | Priority | Smells | Decision | Action |
|-------|----------|--------|----------|--------|
| **#139** | P0 | #3 | Lenient/warning | Change loader to warn on unknown `bosun.*` keys instead of error |
| **#140** | P1 | #1 | Rename now | Rename `BackupEnabled` → `JobEnabled` across schema, labels, docs |
| **#141** | P2 | #16,17,18 | Thin CLI | Move wiring to `internal/app/factory` or `internal/bootstrap` |
| **#142** | P1 | #23,24 | Plan-is-truth | Planner adds start step; Executor interprets `plan.Steps` |
| **#143** | P1 | #4 | Option B | Remove unused `discoverer` param; keep `ExecuteJob(ctx, job)` |

### New Findings (Not Yet Issued)

| Smell | Title | Priority | Suggested Action |
|-------|-------|----------|------------------|
| #2 | Duplicate CLI error printing | Low | Centralize error formatting |
| #5 | Missing restart step in plan display | Med | Covered by #142 |
| #6 | Worker runner silent error swallowing | Med | Return/log errors; name magic codes |
| #7 | JSON/YAML output duplication | Low | Extract output helpers |
| #8, #30 | Brittle integration tests | Med | Adopt testcontainers |
| #9 | TODOs without issue links | Low | Triage and link |
| #10 | Magic strings (Mode, state) | Low | Introduce typed constants |
| #11 | `ExecuteJob` too large | Med | Extract helper methods |
| #12 | `Runner.Run` too complex | Med | Extract lifecycle methods |
| #13, #14 | Loader complexity | Low | Defer unless causing bugs |
| #19–22 | CLI/test duplication | Low | Extract shared helpers |
| #25 | Error handling under-specified | Med | Define retry/backoff policy |
| #26, #28 | ports↔domain coupling | Low | Decide DTO vs coupling |
| #27 | CreatedAt ownership | Low | Pick planner or caller |
| #29 | Merge treats false=zero | Med | Use pointer bools or doc |
| #31 | Adapter returns domain errors | Low | Decide ports vs domain errors |
| #32 | Worker magic exit codes | Med | Name constants, add retry |
