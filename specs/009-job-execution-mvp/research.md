# Research Summary: Job Execution MVP (M3)

**Feature Branch**: `009-job-execution-mvp`
**Date**: 2025-12-29

This document consolidates all research decisions for M3 implementation.

---

## Research #109: Compose Control Strategy

**Question**: Should Bosun control Compose stacks via direct Docker API or by shelling out to `docker compose` CLI?

### Decision

Use **Docker API + Labels + Topological Sort** (no Compose CLI/library).

### Rationale

| Option | Pros | Cons | Decision |
|--------|------|------|----------|
| Docker API direct | Full control, no CLI dependency, cleaner error handling | Must replicate ordering logic | ✅ CHOSEN |
| `docker compose` CLI | Proven, handles edge cases | Text parsing, subprocess spawning | ❌ |
| Compose v2 library | Full feature set | Large dependency, overkill for M3 | ❌ (Future M6+) |

**Key factors**:
1. **Watchtower pattern**: Proven API-only approach works for similar use cases
2. **Simpler MVP**: Ship faster without Compose library complexity
3. **Clean architecture**: Aligns with hexagonal design (adapters hide Docker SDK details)

### Alternatives Considered

**Compose v2 Library** (`github.com/docker/compose/v2/pkg/compose`):
- Pros: Full `depends_on` conditions, health check waiting, orphan cleanup
- Cons: Large dependency (~50+ transitive packages), overkill for M3 scope
- Decision: Defer to M6+ if users report issues with complex stacks

**Known Limitation** (documented):
> ⚠️ M3 does not support `depends_on: condition: service_healthy`.
> Stacks with health-based dependencies may not restart correctly.

### Implementation Details

**Labels to read**:
- `com.docker.compose.project` → Stack identification
- `com.docker.compose.service` → Service name within stack
- `com.docker.compose.depends_on` → Dependency graph (if present)

**Algorithm**:
```
1. List containers with label com.docker.compose.project=<stack>
2. Build dependency graph from com.docker.compose.depends_on
3. Topological sort (DFS with cycle detection)
4. Stop: Iterate reversed sorted order (dependents first)
5. Run worker container
6. Start: Iterate sorted order (dependencies first)
```

**Source**: `.serena/memories/m3_compose_control_decision.md`

---

## Research #110: Worker Container Architecture

**Question**: What is the minimum contract between Bosun and worker containers?

### Decision

**BYOI (Bring Your Own Image)** with minimal metadata injection.

### Rationale

Workers define their own environment (RESTIC_*, PGHOST, etc.). Bosun injects only metadata.

| Aspect | Decision |
|--------|----------|
| Environment | `BOSUN_*` metadata + label-based pass-through |
| Signals | SIGTERM → 10s grace → SIGKILL |
| Cleanup | Always remove; `--keep-failed` to preserve |
| Base Images | BYOI only for M3; example Dockerfiles in docs |
| Communication | Exit codes only (no stdout parsing) |

### Alternatives Considered

**Callback mechanisms** (HTTP, file-based, stdout parsing):
- Pros: Progress reporting, real-time status
- Cons: Complexity, worker changes required
- Decision: Defer to M5+ (structured output milestone)

**Official base images** (`bosun/worker-postgres:1.0`):
- Pros: Convenience, blessed patterns
- Cons: Maintenance burden, registry setup
- Decision: Defer to M6+ based on user demand

### Implementation Details

**Bosun-injected environment**:
| Variable | Description |
|----------|-------------|
| `BOSUN_JOB_NAME` | Job identifier from labels |
| `BOSUN_RUN_ID` | Unique execution ID (UUID v4) |
| `BOSUN_STACK` | Target compose stack name |
| `BOSUN_DRY_RUN` | `true` or `false` |

**User pass-through** (from labels):
```yaml
labels:
  bosun.job.worker-env.RESTIC_REPOSITORY: "s3:s3.amazonaws.com/mybucket"
  bosun.job.worker-env.RESTIC_PASSWORD_FILE: "/run/secrets/restic-password"
```

**Exit code interpretation**:
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1-125 | Application error |
| 126 | Command not executable |
| 127 | Command not found |
| 137 | SIGKILL (timeout) |
| 143 | SIGTERM (interrupted) |

**Source**: `.serena/memories/m3_worker_contract.md`

---

## Research #117: Failure Handling

**Question**: How should Bosun handle failures during execution?

### Decision

Fail-fast with guaranteed restart attempt.

### Rationale

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Stop timeout | 30s default | Docker 10s too short for DBs; AWS ECS uses 30s |
| Start timeout | 30s default | Consistent with stop |
| Worker timeout | 1h default | Long enough for most backups |
| Restart policy | Always restart | Availability > backup success |
| Pre-validation | Check image before stop | Fail fast, don't stop stack if image missing |

### Alternatives Considered

**Shorter timeouts (10s)**:
- Pros: Faster failure detection
- Cons: Database containers need longer for clean shutdown
- Decision: 30s default, configurable via labels

**Skip restart on worker failure**:
- Pros: Keep evidence for debugging
- Cons: Production stack stays down
- Decision: Always restart; `--keep-stopped` flag for debugging

### Implementation Details

**Timeout configuration**:
| Step | Default | Label Override | CLI Override |
|------|---------|----------------|--------------|
| Stop stack | 30s | `bosun.job.stop-timeout` | `--stop-timeout` |
| Run worker | 1h | `bosun.job.timeout` | `--timeout` |
| Start stack | 30s | `bosun.job.start-timeout` | `--start-timeout` |

**Failure scenarios**:

| Scenario | Bosun Action | Stack State |
|----------|--------------|-------------|
| Stop timeout | Abort, do NOT proceed | Partially stopped |
| Worker fails | Still restart stack | Restarted |
| Start fails | Report error | Partially started |
| Ctrl+C | Attempt restart before exit | Best effort |

**Pre-validation flow**:
```
1. ImageInspect to check local cache
2. If not local, pull image
3. If pull fails, abort BEFORE stopping stack
```

**Source**: `.serena/memories/m3_failure_handling.md`

---

## Clarifications Resolved

| Question | Answer | Source |
|----------|--------|--------|
| Worker failure → restart? | Yes, always restart by default | #117 |
| Maintenance mode | `--keep-stopped` flag | #117 |
| Log persistence | Display only; persistence in M5 | Spec |
| Log streaming | Real-time via Docker attach | Spec |
| Concurrency/locking | No locking in M3; deferred to M4 | Spec |

---

## External Resources

### Watchtower Codebase Analysis
- **Location**: `./watchtower/` in workspace
- **Relevant patterns**: Signal handling, lifecycle timeouts, container ordering
- **Key insight**: API-only approach proven viable for similar use case

### Portainer Codebase Analysis
- **Location**: `./portainer/` in workspace
- **Relevant patterns**: Compose v2 library integration, status polling
- **Key insight**: Library approach adds significant complexity

### Industry Standards
| System | Stop Timeout | Grace Period |
|--------|--------------|--------------|
| Docker default | 10s | 10s |
| Docker Compose | 10s | 10s |
| AWS ECS | 30s | 30s |
| Kubernetes | 30s | 30s |

---

## Open Items for Future Milestones

| Item | Target | Rationale |
|------|--------|-----------|
| Health check waiting | M6+ | Compose v2 library integration |
| Log persistence | M5 | Requires storage adapter |
| Concurrent job locking | M4 | Requires job queue |
| Progress reporting | M5 | Structured output |
| Official base images | M6+ | Community demand driven |
