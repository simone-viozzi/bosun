# WIP Research: Compose Control Strategy (#109)

**Status**: NOT STARTED
**GitHub Issue**: #109
**Spec Reference**: `specs/009-job-execution-mvp/spec.md`
**Blocks**: M3 implementation (FR-001, FR-002, FR-003, FR-004)

## Research Question

Should Bosun control Compose stacks via direct Docker API or by shelling out to `docker compose` CLI?

## Context from Spec

Bosun needs to:
1. **Stop** all containers in a Compose stack before worker runs
2. **Start** all containers after worker completes
3. Respect container dependency order during stop/start
4. Wait for health checks during startup

## Options

### Option A: Direct Docker API

Use `com.docker.compose.project` labels to identify stack containers, call Docker API directly.

**Potential Pros**:
- No CLI dependency
- More control
- Programmatic error handling
- No text parsing

**Potential Cons**:
- Must replicate Compose's start order logic
- Must handle health checks manually
- Must handle dependency ordering

### Option B: Shell to `docker compose` CLI

Use `exec.Command` to call `docker compose down/up`.

**Potential Pros**:
- Compose handles all orchestration logic
- Proven, handles edge cases
- Less code to maintain

**Potential Cons**:
- CLI dependency (requires docker compose installed)
- Parsing text output for errors
- Less control over process

## Research Tasks

- [ ] Research how Compose determines container start order
- [ ] Investigate edge cases where API-only approach might fail
- [ ] Examine what metadata Compose stores beyond labels
- [ ] Review how other tools handle this (Portainer, Dockge)
- [ ] Compare performance: API calls vs subprocess

## Questions to Answer

1. Does Compose store dependency graph somewhere Bosun can read?
2. What happens with `depends_on: condition: service_healthy`?
3. Are there compose-only features that can't be replicated via API?
4. What's the performance difference?

## Expected Output

- **Decision**: A or B (with clear rationale)
- **Implications**: What this means for #115 (ComposeController port) and #118 (adapter)
- **Follow-up**: Any new issues or scope changes

## Related Files

- `internal/ports/compose.go` (to be created - #115)
- `internal/adapters/docker/compose.go` (to be created - #118)

---

*This memory will be updated with research findings and final decision.*
