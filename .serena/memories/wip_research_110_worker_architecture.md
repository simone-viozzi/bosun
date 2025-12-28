# WIP Research: Worker Container Architecture (#110)

**Status**: NOT STARTED
**GitHub Issue**: #110
**Spec Reference**: `specs/009-job-execution-mvp/spec.md`
**Blocks**: M3 implementation (FR-010, FR-011, FR-012)

## Research Question

What is the minimum contract between Bosun and worker containers? Should Bosun provide base images?

## Context from Spec

Bosun needs to:
1. Run a worker container with volumes attached
2. Pass job metadata via environment variables
3. Capture exit code, stdout, stderr
4. Handle timeouts gracefully
5. Clean up container after execution

## Design Areas

### 1. Environment Variables (FR-010)

What metadata should Bosun inject into workers?

**Candidates**:
- `BOSUN_JOB_NAME` - Job identifier
- `BOSUN_JOB_ID` - Unique run ID
- `BOSUN_VOLUMES` - Comma-separated volume names
- `BOSUN_STACK` - Target stack name
- `BOSUN_TIMEOUT` - Timeout in seconds
- `BOSUN_DRY_RUN` - "true" or "false"

### 2. Signal Protocol (FR-011)

How should Bosun handle worker timeouts?

**Options**:
- **Simple**: SIGKILL immediately on timeout
- **Graceful**: SIGTERM → wait grace period → SIGKILL
- **Warning**: Send signal/env update before timeout (complex)

### 3. Container Cleanup (FR-012)

Should failed workers be kept for debugging?

**Options**:
- **Always remove**: Clean, simple
- **Keep on failure**: Useful for debugging, but clutters system
- **Configurable**: `--keep-failed` flag or label

### 4. Base Images

Should Bosun provide official base images?

**Options**:
- **A) BYOI only**: Users bring their own images, minimal contract
- **B) Base images**: Bosun provides images with helpers (signal client, common tools)
- **C) Hybrid**: BYOI works, base images add convenience

### 5. Worker Communication

How can workers send signals back to Bosun?

**Options**:
- Exit codes only (simplest)
- stdout parsing (e.g., `BOSUN: PROGRESS 50%`)
- HTTP callback (complex, requires Bosun to listen)
- File-based (write to mounted volume)

## Research Tasks

- [ ] Define minimum environment variable set
- [ ] Decide on signal escalation for timeouts
- [ ] Decide on container cleanup policy
- [ ] Evaluate base image value vs complexity
- [ ] Design worker communication (if any beyond exit code)

## Example Workers to Consider

- PostgreSQL pg_dump worker
- MySQL/MariaDB dump worker
- Restic backup worker
- Generic file/volume copy worker
- Redis RDB snapshot worker

## Expected Output

- **Decision**: Minimum contract definition
- **Environment Variables**: List with descriptions
- **Signal Protocol**: SIGTERM/SIGKILL behavior
- **Cleanup Policy**: Default + configuration options
- **Base Images**: Yes/No for M3 (can defer to M6)

## Related Files

- `internal/ports/worker.go` (to be created - #116)
- `internal/adapters/docker/worker.go` (to be created - #119)

---

*This memory will be updated with research findings and final decision.*
