# WIP Research: Compose Failure Handling (#117)

**Status**: NOT STARTED
**GitHub Issue**: #117
**Spec Reference**: `specs/009-job-execution-mvp/spec.md`
**Blocks**: M3 implementation (FR-005, FR-014, FR-023)
**Depends On**: #109 (need to know API vs CLI first)

## Research Question

How should Bosun handle failures during Compose stop/start operations?

## Context from Spec

Bosun executes: **stop stack → run worker → start stack**

Failures can occur at any step. We need clear policies for:
- Timeouts at each step
- What to do when a step fails
- Whether to proceed or abort
- What state to leave the system in

## Failure Scenarios

### Scenario 1: `down` hangs

Container won't stop (stuck process, I/O wait, etc.)

**Options**:
- Timeout → force kill containers
- Timeout → abort job, leave stack running
- Timeout → escalate (SIGTERM → SIGKILL)

### Scenario 2: `down` fails

Compose returns error (container not found, permission denied, etc.)

**Options**:
- Retry once then abort
- Abort immediately
- Continue with partial stop (dangerous)

### Scenario 3: Worker fails (after stack stopped)

Worker exits non-zero or times out.

**Questions**:
- Should stack still be restarted? (probably yes)
- Should this be configurable? (`bosun.job.restart-on-failure`)
- What if user wants "maintenance mode"?

### Scenario 4: `up` fails (after worker completed)

Stack won't restart (image missing, port conflict, health check fails, etc.)

**Options**:
- Retry once then report error
- Report error immediately, leave in failed state
- Attempt rollback (complex, probably out of scope)

### Scenario 5: Partial failure

Some containers restart, others fail.

**Options**:
- Report partial success with details
- Treat as failure
- Retry individual containers (complex)

## Timeout Strategy

### Questions to Answer

1. What is a reasonable default timeout for `down`? (30s? 60s? 5m?)
2. What is a reasonable timeout for `up`? (60s? 2m? depends on health checks?)
3. Should timeouts be configurable per-job via labels?
4. Should there be an overall job timeout vs per-step timeouts?

### Signal Escalation

For stuck containers:
1. Send SIGTERM
2. Wait grace period (e.g., 10s)
3. Send SIGKILL

This is standard Docker behavior - should Bosun use Docker's defaults or configure its own?

## Pre-Validation (FR-023)

Should Bosun validate/pull the worker image BEFORE stopping the stack?

**Pros**:
- Fail fast - don't stop stack if image is missing
- Better user experience

**Cons**:
- Extra latency
- Image could disappear between check and use (race condition)
- More complex

**Options**:
- **A) Pre-pull**: Always pull/verify image before stop
- **B) Fail fast**: Try to create container before stop, fail if image missing
- **C) Optimistic**: Stop first, deal with image failure later

## Rollback Strategy

If `up` fails after worker completed, what state do we leave the system in?

**Options**:
- **A) Leave failed**: User must manually fix
- **B) Retry once**: Try `up` again
- **C) Keep retrying**: Dangerous, could loop forever
- **D) Track last good state**: Complex, requires state storage

For M3, **Option A** (leave failed, report clearly) is probably appropriate.

## Research Tasks

- [ ] Define default timeouts for down/up operations
- [ ] Define timeout configuration mechanism (labels? CLI?)
- [ ] Define restart policy on worker failure
- [ ] Define pre-validation strategy for worker image
- [ ] Define behavior for each failure scenario
- [ ] Consider Ctrl+C handling (graceful shutdown)

## Expected Output

- **Timeout Defaults**: Specific values with rationale
- **Failure Matrix**: What happens in each scenario
- **Restart Policy**: Default + configuration
- **Pre-Validation**: Yes/No with reasoning
- **User Feedback**: What messages/logs user sees

## Related Files

- `internal/ports/compose.go` (error types - #115)
- `internal/adapters/docker/compose.go` (implementation - #118)
- `internal/app/executor/executor.go` (orchestration logic - #121)

---

*This memory will be updated with research findings and final decision.*
