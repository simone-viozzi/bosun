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

| ID | Title | Status | Evidence | TODO Anchors |
|----|-------|--------|----------|--------------|
| 1 | Confusing `BackupEnabled` property name | Needs-Answers | `wip_smell-general-m3` #1 | — |
| 2 | Duplicate CLI error printing | New | `wip_smell-general-m3` #2 | — |
| 3 | Strict unknown-key rejection (may cause #136) | Needs-Answers | `wip_smell-general-m3` #3 | — |
| 4 | Executor API mismatch / unused `discoverer` param | New | `wip_smell-general-m3` #4 | — |
| 5 | Missing restart step in plan display (#135) | New | `wip_smell-general-m3` #5 | — |
| 6 | Worker runner silent error swallowing / magic 137 | New | `wip_smell-general-m3` #6 | — |
| 7 | JSON/YAML output duplication across CLI commands | New | `wip_smell-general-m3` #7 | — |
| 8 | Brittle integration test harness (needs testcontainers) | New | `wip_smell-general-m3` #8 | — |
| 9 | TODOs without issue links | New | `wip_smell-general-m3` #9 | — |
| 10 | Magic strings for Mode (`ro`/`rw`) and container state | New | `wip_smell-general-m3` #10 | — |
| 11 | `Executor.ExecuteJob` too large (~167 lines, many responsibilities) | New | `wip_smell-complexity-m3` #1 | — |
| 12 | `Runner.Run` too complex (~103 lines, concurrency + lifecycle) | New | `wip_smell-complexity-m3` #2 | — |
| 13 | `FromLabels` + reflection helpers complex | New | `wip_smell-complexity-m3` #3 | — |
| 14 | `ValidateJobLabels` too large (~159 lines) | New | `wip_smell-complexity-m3` #4 | — |
| 15 | `Planner.Plan` slightly over threshold (~69 lines) | New | `wip_smell-complexity-m3` #5 | — |
| 16 | CLI imports adapters directly (layering violation) | Needs-Answers | `wip_smell-layering-m3` #1 | — |
| 17 | Business logic in CLI handlers (discovery, selection) | New | `wip_smell-layering-m3` #2 | — |
| 18 | CLI performs low-level Docker client wiring | New | `wip_smell-layering-m3` #3 | — |
| 19 | Duplicated CLI flag setup & format validation | New | `wip_smell-duplication-m3` #2 | — |
| 20 | Docker client wiring duplicated across CLI/tests | New | `wip_smell-duplication-m3` #3 | — |
| 21 | Near-identical test setup calls | New | `wip_smell-duplication-m3` #5 | — |
| 22 | Duplicate discovery/wiring sequences (snapshot→discover) | New | `wip_smell-duplication-m3` #6 | — |
| 23 | Planner vs Executor mismatch (plan not authoritative) | Needs-Answers | `wip_smell-design-m3` #1 | — |
| 24 | Execution plan incompleteness (no start step, no per-step policies) | Needs-Answers | `wip_smell-design-m3` #4 | — |
| 25 | Error handling & retry policy under-specified | New | `wip_smell-design-m3` #6 | — |
| 26 | `ports` imports domain types (coupling tradeoff) | New | `wip_smell-design-m3` #7 | — |

---

## Open Questions (Blocking)

1. **Smell 3**: Are `bosun.other` / `bosun.network` intended extension labels that should be tolerated (warning) vs rejected as unknown keys?
2. **Smell 16**: Do you want CLI to be forbidden from importing adapters? (Rule: CLI calls only app-level services)
3. **Smell 23**: Should `ExecutionPlan` be authoritative and drive runtime execution, or remain a preview (DryRun only)?
4. **Smell 4**: Which API contract for job execution? (A) `Execute(ctx, jobName)` with discoverer inside executor, or (B) `ExecuteJob(ctx, job)` with discovery external?
5. **Smell 1**: Breaking rename of `BackupEnabled` now, or keep + add alias/docs?

## Open Questions (Non-Blocking)

- **Smell 1**: Preferred rename for `BackupEnabled`? Options: `BackupsEnabled`, `VolumeBackupsEnabled`, `JobEnabled`, or keep + add docs.
- **Smell 2**: Structured (machine-friendly) vs free-form CLI error messages?
- **Smell 4**: Should `Execute`/`DryRun` be implemented, or is `ExecuteJob`/`DryRunJob` the intentional API?
- **Smell 11**: Which complexity hotspot to refactor first? (A) Executor, (B) Runner, (C) Loader, (D) All.
- **Smell 16**: Where should wiring live if not in CLI? (a) `internal/app` factory, (b) `internal/bootstrap`, (c) `cmd/bosun/main.go` only.

---

## Session Log

- **2026-01-04**: Scope confirmed with user. Spawning initial scouts.
- **2026-01-04**: 3 scouts completed (general, complexity, layering). Consolidated 18 smells into index.
- **2026-01-04**: 2 more scouts completed (duplication, design). Total 26 smells. 5 blocking questions identified.
- **2026-01-04**: Created 5 GitHub issues from smell findings, attached as sub-issues to #138:
  - #139 — Fix invalid labels in test compose (P0) ← fixes #136
  - #140 — Rename "backup" terminology to "job" (P1)
  - #141 — CLI imports adapters directly (P2)
  - #142 — ExecutionPlan should be authoritative (P1)
  - #143 — Fix JobExecutor interface mismatch (P1)
