# WIP: Milestone 3 Process Status

**Last Updated**: 2025-12-28
**Feature Branch**: `009-job-execution-mvp`
**GitHub Issue**: #85 (M3: Job Execution MVP)

## Current Phase: Research In Progress

### Workflow Progress

| Chat | Phase | Status | Output |
|------|-------|--------|--------|
| **Chat 1** | `/speckit.specify` | ✅ Complete | `spec.md`, WIP memories |
| **Chat 2** | Research #109 | ✅ Complete | `m3_compose_control_decision.md` |
| **Chat 3** | Research #110 | ⏳ Not started | Worker architecture decision |
| **Chat 4** | Research #117 | ⏳ Not started | Failure handling decision |
| **Chat 5** | `/speckit.plan` | ⏳ Blocked | `plan.md`, `data-model.md` |
| **Chat 6** | `/speckit.tasks` | ⏳ Blocked | `tasks.md` |

## Created Artifacts

### Spec Directory: `specs/009-job-execution-mvp/`
- `spec.md` - Full specification with 5 user stories, 25 functional requirements
- `checklists/requirements.md` - Quality checklist tracking 8 [NEEDS CLARIFICATION] markers

### Decision Memories: `.serena/memories/`
- `m3_compose_control_decision.md` - **Decision: API + Labels + Topological Sort**
- `wip_research_watchtower.md` - Watchtower patterns (complete)
- `wip_research_portainer.md` - Portainer patterns (complete)

### WIP Research Memories: `.serena/memories/`
- `wip_research_110_worker_architecture.md` - Context for Chat 3
- `wip_research_117_failure_handling.md` - Context for Chat 4

## GitHub Sub-Issues (12 total)

### Research (must complete first)
- [x] #109 - Compose Control Strategy → **Decision: Docker API + Labels** (closed)
- [ ] #110 - Worker Architecture (signals, base images)
- [ ] #117 - Failure Handling (timeouts, rollback)

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

## [NEEDS CLARIFICATION] Summary (6 remaining, 2 resolved)

### ~~Research #109 - Compose Control~~ ✅ RESOLVED
| FR | Question | Resolution |
|----|----------|------------|
| ~~FR-003~~ | ~~Container dependency order~~ | Topological sort on `com.docker.compose.depends_on` |
| ~~FR-004~~ | ~~Health check waiting~~ | Not in M3 (documented limitation), #125 for future | |
| FR-004 | How to wait for health checks during startup |

### Research #110 - Worker Architecture
| FR | Question |
|----|----------|
| FR-010 | What environment variables to inject (BOSUN_JOB_NAME, etc.) |
| FR-011 | Signal protocol for timeouts (SIGTERM → SIGKILL?) |
| FR-012 | Container cleanup strategy (keep on failure?) |

### Research #117 - Failure Handling
| FR | Question |
|----|----------|
| FR-005 | Default timeouts for stop/start, per-job config |
| FR-014 | Stack restart policy when worker fails |
| FR-023 | Pre-pull/validate worker image before stopping stack |

### Additional (lower priority)
- Restart-on-failure configuration label
- Maintenance mode feature
- Log persistence vs display-only
- Log streaming approach
- Concurrency/locking for M3

## Next Steps

### Chat 2 Prompt
```
Research for M3 spec 009-job-execution-mvp.

Read: .serena/memories/wip_research_109_compose_control.md

Question: Docker API vs docker compose CLI for stack control?

Tasks:
- Research Compose dependency handling
- Review Portainer/Dockge approaches
- Prototype both options if needed
- Make decision with rationale

Output:
1. Update wip_research_109_compose_control.md with findings
2. Create final memory: m3_compose_control_decision.md
```

### Chat 3 Prompt
```
Research for M3 spec 009-job-execution-mvp.

Read: .serena/memories/wip_research_110_worker_architecture.md

Question: Worker contract - env vars, signals, cleanup, base images?

Output:
1. Update wip_research_110_worker_architecture.md with findings
2. Create final memory: m3_worker_contract.md
```

### Chat 4 Prompt
```
Research for M3 spec 009-job-execution-mvp.

Read: .serena/memories/wip_research_117_failure_handling.md
(Uses Chat 2 outcome for API vs CLI context)

Question: Timeouts, failure behavior, restart policy?

Output:
1. Update wip_research_117_failure_handling.md with findings
2. Create final memory: m3_failure_handling.md
3. Update spec.md to resolve [NEEDS CLARIFICATION] markers
```

### Chat 5
Run `/speckit.plan` after all research is complete.

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
