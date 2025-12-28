# WIP: Milestone 3 Process Status

**Last Updated**: 2025-12-28
**Feature Branch**: `009-job-execution-mvp`
**GitHub Issue**: #85 (M3: Job Execution MVP)

## Current Phase: RESEARCH COMPLETE - Ready for Planning

### Workflow Progress

| Chat | Phase | Status | Output |
|------|-------|--------|--------|
| **Chat 1** | `/speckit.specify` | ✅ Complete | `spec.md`, WIP memories |
| **Chat 2** | Research #109 | ✅ Complete | `m3_compose_control_decision.md` |
| **Chat 3** | Research #110 | ✅ Complete | `m3_worker_contract.md` |
| **Chat 4** | Research #117 | ✅ Complete | `m3_failure_handling.md` |
| **Chat 5** | `/speckit.plan` | ⏳ Ready | `plan.md`, `data-model.md` |
| **Chat 6** | `/speckit.tasks` | ⏳ Blocked | `tasks.md` |

## Created Artifacts

### Spec Directory: `specs/009-job-execution-mvp/`
- `spec.md` - Full specification with 5 user stories, 25 functional requirements
- `checklists/requirements.md` - Quality checklist tracking 8 [NEEDS CLARIFICATION] markers

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
- [x] #117 - Failure Handling → **Decision: 30s timeouts, always restart, pre-validate** (to close)

### Future (created from research)
- [ ] #125 - Add Compose v2 library support for complex stacks

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

### Chat 5 - Planning (NEXT)
```
Run `/speckit.plan` for spec 009-job-execution-mvp.

Input: specs/009-job-execution-mvp/spec.md (all clarifications resolved)

Reference memories:
- .serena/memories/m3_compose_control_decision.md
- .serena/memories/m3_worker_contract.md
- .serena/memories/m3_failure_handling.md

Output:
1. plan.md - Implementation plan with component architecture
2. data-model.md - Domain types and interfaces
```

### Chat 6 - Tasks
```
Run `/speckit.tasks` for spec 009-job-execution-mvp.

Input: specs/009-job-execution-mvp/plan.md

Output:
1. tasks.md - GitHub-ready task list for all sub-issues
```

## Dependencies

```
#109 (Compose Strategy) ──┐
#110 (Worker Architecture)├──► Implementation Tasks ──► #85 (M3 Complete)
#117 (Failure Handling) ──┘
        ↑
   (depends on #109)
```

---

*Update this memory as each chat completes.*
