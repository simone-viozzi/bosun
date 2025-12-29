# WIP: Milestone 3 Process Status

**Last Updated**: 2025-12-29
**Feature Branch**: `009-job-execution-mvp`
**GitHub Issue**: #85 (M3: Job Execution MVP)

## Current Phase: IMPLEMENTATION IN PROGRESS

### Recent Progress (2025-12-29)

**Bug Fix - COMPLETED:**
- Fixed worker log capture deadlock in `internal/adapters/docker/worker/runner.go`
  - Moved log capture AFTER `ContainerWait()` (was blocking before)
  - Changed `Follow: false` since container is already stopped
  - Memory `wip_m3_log_capture_bug` deleted (no longer needed)

**T017.1 Signal Handler - COMPLETED:**
- Implemented `SIGINT/SIGTERM` handler in `internal/cmd/job_run.go`
- Context cancellation on signal → executor aborts → stack restart (via defer)
- Exit code 16 (`ExitInterrupted`) on Ctrl+C

**Phase 4 Dry Run - COMPLETED (T020-T024):**
- Added `--dry-run` flag to preview execution plan
- Added `--format` flag (text, json)
- Implemented `printDryRunText()` and `printDryRunJSON()` formatters
- `DryRunJob()` already existed in executor

**Additional Fixes:**
- Fixed deprecated `ImageInspectWithRaw` → `ImageInspect`
- Fixed errcheck lint warning in `captureLogs()`
- Fixed test count in `TestJobSpecByScope` (7 container-scope keys after M3 additions)
- Changed `JobLabelConfig` timeout fields from `string` to `time.Duration`

### Remaining MVP Tasks

| Phase | Task | Description | Status |
|-------|------|-------------|--------|
| 3 | T012 | ComposeController unit tests | ✅ Complete |
| 3 | T015 | WorkerRunner unit tests | ✅ Complete |
| 3 | T018 | Executor unit tests | ✅ Complete |
| 4 | T024 | Dry-run unit test | ✅ Complete |

**MVP Phase 1-4 Implementation COMPLETE** (2025-12-29)


## Created Artifacts

### Spec Directory: `specs/009-job-execution-mvp/`
- `spec.md` - Full specification with 5 user stories, 25 functional requirements
- `plan.md` - Implementation plan with architecture, dependencies, phases
- `data-model.md` - Domain types, port interfaces, error types
- `research.md` - Consolidated research decisions
- `quickstart.md` - Developer implementation guide
- `tasks.md` - 57 tasks across 9 phases (24 MVP, 33 post-MVP)
- `contracts/` - Go interface definitions
  - `compose_controller.go` - ComposeController port (#115)
  - `worker_runner.go` - WorkerRunner port (#116)
  - `job_executor.go` - JobExecutor port (#114)

### Decision Memories: `.serena/memories/`
- `m3_compose_control_decision.md` - **Decision: API + Labels + Topological Sort**
- `m3_worker_contract.md` - **Decision: BOSUN_* env vars, SIGTERM→SIGKILL, BYOI**
- `m3_failure_handling.md` - **Decision: 30s timeouts, always restart, pre-validate image**
- `wip_research_watchtower.md` - Watchtower patterns (complete)
- `wip_research_portainer.md` - Portainer patterns (complete)

### WIP Research Memories: `.serena/memories/`
- All WIP research memories cleaned up (converted to decision memories)

## GitHub Sub-Issues (12 total)

### Research (ALL COMPLETE)
- [x] #109 - Compose Control Strategy → **Decision: Docker API + Labels** (closed)
- [x] #110 - Worker Architecture → **Decision: BOSUN_* env, SIGTERM→SIGKILL, BYOI** (closed)
- [x] #117 - Failure Handling → **Decision: 30s timeouts, always restart, pre-validate** (closed)

### Deferred Features (M6+)
- [ ] #125 - Add Compose v2 library support for complex stacks
- [ ] #126 - Pass BOSUN_VOLUMES environment variable to worker containers
- [ ] #127 - Wait for health checks during stack startup

### Port Definitions
- [ ] #115 - ComposeController port interface
- [ ] #116 - WorkerRunner port interface
- [ ] #114 - JobExecutor port interface

### Adapter Implementation
- [ ] #118 - ComposeController adapter
- [ ] #119 - WorkerRunner adapter

### Application Layer
- [ ] #121 - Executor service

### CLI & Integration
- [ ] #122 - `bosun job run` CLI command
- [ ] #123 - M3 Integration tests
- [ ] #120 - M3 Basic docs update

## [NEEDS CLARIFICATION] Summary - ALL RESOLVED ✅

### Research #109 - Compose Control ✅ RESOLVED
| FR | Question | Resolution |
|----|----------|------------|
| FR-003 | Container dependency order | Topological sort on `com.docker.compose.depends_on` |
| FR-004 | Health check waiting | Not in M3 (documented limitation), #125 for future |

### Research #110 - Worker Architecture ✅ RESOLVED
| FR | Question | Resolution |
|----|----------|------------|
| FR-010 | Environment variables | BOSUN_JOB_NAME, BOSUN_RUN_ID, BOSUN_STACK, BOSUN_DRY_RUN + label pass-through |
| FR-011 | Signal protocol | SIGTERM → 10s grace → SIGKILL (Docker standard) |
| FR-012 | Container cleanup | Always remove; `--keep-failed-worker` flag to preserve |

### Research #117 - Failure Handling ✅ RESOLVED
| FR | Question | Resolution |
|----|----------|------------|
| FR-005 | Default timeouts | 30s stop/start, 1h worker, label-configurable |
| FR-014 | Stack restart policy | Always restart; `--keep-stopped` to override |
| FR-023 | Pre-validation | ImageInspect before stop (fail fast) |

### Additional Clarifications ✅ RESOLVED
| Question | Resolution |
|----------|------------|
| Restart-on-failure config | `--keep-stopped` CLI flag |
| Maintenance mode | Use `--keep-stopped` flag |
| Log persistence | Display only in M3, persistence in M5 |
| Log streaming | Real-time via Docker attach |
| Concurrency/locking | No locking in M3, deferred to M4 |

## Next Steps

### Chat 6 - Implementation (NEXT)
```
Run `/speckit.implement` for spec 009-job-execution-mvp.

Input: specs/009-job-execution-mvp/tasks.md (Phases 1-4 = MVP)

Reference:
- specs/009-job-execution-mvp/plan.md (architecture)
- specs/009-job-execution-mvp/data-model.md (types)
- specs/009-job-execution-mvp/contracts/ (interfaces)
- specs/009-job-execution-mvp/quickstart.md (guide)

MVP Tasks (24 total):
- Phase 1: Setup (T001-T005) - Domain types, error types, exit codes
- Phase 2: Ports (T006-T008) - ComposeController, WorkerRunner, JobExecutor interfaces
- Phase 3: US1 Execute Job (T009-T019) - Adapters, Executor, CLI
- Phase 4: US2 Dry Run (T020-T024) - --dry-run, --format flags
```

### Post-MVP (Phases 5-9)
After MVP validation:
- Phase 5: US3 Logs (T025-T029)
- Phase 6: US4 Timeouts (T030-T035)
- Phase 7: US5 Dependencies (T036-T039)
- Phase 8: Integration Tests (T040-T050)
- Phase 9: Polish (T051-T056)

## Analysis Summary (Chat 5)

### Issues Found & Remediated
| ID | Severity | Issue | Resolution |
|----|----------|-------|------------|
| I1 | HIGH | Env var naming (BOSUN_STACK_NAME vs BOSUN_STACK) | Fixed: Use BOSUN_STACK |
| C1 | HIGH | Missing Ctrl+C task | Added T017.1 for signal handler |
| I2 | MEDIUM | Flag naming (--keep-failed-worker) | Fixed: Use --keep-failed |
| I3 | MEDIUM | Exit code collision (1-6) | Fixed: Use 10-16 range |
| A1/A2 | MEDIUM | Health check ambiguity | Marked DEFERRED, created #127 |

### Metrics
- Requirements: 25 (24 covered, 1 deferred)
- Tasks: 57 (24 MVP, 33 post-MVP)
- Coverage: 96%
- Constitution: All 5 principles PASS

## Dependencies

```
Research Complete ──► Planning Complete ──► Implementation Ready
     ↓                      ↓                      ↓
#109, #110, #117      plan.md, tasks.md      Phase 1-4 (MVP)
   (closed)              (done)                (next)
```

### Implementation Order
```
Phase 1 (Setup) ────► Phase 2 (Ports) ────► Phase 3 (US1) ────► Phase 4 (US2)
   T001-T005            T006-T008            T009-T019           T020-T024
   (parallel)           (parallel)           (sequential)        (sequential)
                                                  │
                                                  ▼
                                            MVP COMPLETE
```

---

*Update this memory as each chat completes.*
