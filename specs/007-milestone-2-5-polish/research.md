# Research: Milestone 2.5 – Polish Job Model, Config Schema & Test Suite

**Date**: 2025-11-30
**Status**: Complete

## Research Tasks

### 1. Current Validation Logic Locations

**Question**: Where is job validation logic currently duplicated?

**Findings**:

| File | Validation Logic |
|------|-----------------|
| `internal/adapters/joblabels/discoverer.go` | `isJobEnabled()`, cron validation, mount mode checks, label constants |
| `internal/config/loader/job_validation.go` | `ValidateJobLabels()`, enabled/name/schedule/conflict checks |
| `internal/config/schema/job_labels.go` | `JobLabelConfig`, `JobVolumeConfig` structs with tag-based defaults |
| `internal/domain/jobs/types.go` | `DefaultSchedule`, `DefaultWorkerImage`, `DefaultMountMode` constants |

**Decision**: Use `internal/config/schema/job_labels.go` as the **single source of truth** for:
- Label keys (via struct tags)
- Default values (via struct tags)
- Type definitions

Extract validation helpers to a new shared package or move them to the schema package.

**Rationale**: The schema package already has the struct tags with all metadata. The discoverer and loader should import and reuse this.

---

### 2. Docker Label Filtering Best Practices

**Question**: How to efficiently filter Docker containers by `com.docker.compose.project` label?

**Findings**:

Docker SDK supports label-based filtering natively:
```go
containers, err := cli.ContainerList(ctx, container.ListOptions{
    Filters: filters.NewArgs(
        filters.Arg("label", "com.docker.compose.project=myproject"),
    ),
})
```

**Decision**:
- Use Docker's native filtering for `--project` (efficient, server-side)
- Use Docker's native filtering for `--stack` with `bosun.stack=value`
- Both are label-based filters, so same mechanism works

**Rationale**: Server-side filtering is more efficient than fetching all containers and filtering client-side.

---

### 3. Exit Code Conventions

**Question**: What exit codes do similar CLI tools use?

**Findings**:

| Tool | Success | Error | Validation |
|------|---------|-------|------------|
| `grep` | 0 | 2 | 1 (no match) |
| `kubectl` | 0 | 1 | - |
| `docker` | 0 | 1 | - |
| `git` | 0 | 1 | 128+ |
| POSIX | 0 | 1-125 | varies |

**Decision**:
- 0 = success
- 1 = runtime error (Docker unavailable, I/O failure)
- 2 = validation failure (invalid labels, missing required fields)

**Rationale**: Distinguishing runtime from validation errors helps users and scripts handle failures appropriately.

---

### 4. Test Deduplication Strategy

**Question**: How to determine which tests are redundant between unit and integration?

**Findings**:

Current test overlap:
- `internal/adapters/joblabels/discoverer_test.go` - tests job discovery from mock snapshots
- `integration/joblabels_test.go` - tests job discovery from real Docker Compose stack
- Both test the same validation scenarios (missing name, invalid cron, etc.)

**Decision**:
- **Unit tests**: Exhaustive edge cases with mocked inputs (no Docker)
- **Integration tests**: Happy path E2E only (real Docker, CLI invocation)
- Remove duplicate edge case tests from integration suite

**Rationale**: Unit tests are faster, more reliable, and easier to maintain for edge cases. Integration tests prove the system works end-to-end but don't need to re-test every edge case.

---

### 5. CLI Branding Alignment

**Question**: What should the CLI description say?

**Findings**:

README says:
> "Bosun is a Docker-label–driven job orchestrator for single hosts that can safely stop stacks, run worker containers with attached resources on a schedule, and start everything back up"

Current CLI says:
> "Bosun - Docker label management tool"

**Decision**: Update to:
- Root: "Bosun - Docker-label-driven backup job orchestrator"
- Short: "Orchestrate backup jobs defined via Docker labels"
- Long: Match README's description

**Rationale**: Consistency between README and CLI helps users understand the tool's purpose.

---

## Alternatives Considered

### Validation Location Alternatives

| Option | Pros | Cons | Decision |
|--------|------|------|----------|
| Keep in `joblabels` adapter | Close to discovery logic | Duplicates loader logic | ❌ Rejected |
| Keep in `loader` package | Close to config loading | Duplicates adapter logic | ❌ Rejected |
| New `validation` package | Clean separation | Another package to maintain | ❌ Rejected |
| **Move to `schema` package** | Single source of truth, tags define everything | Schema becomes larger | ✅ Chosen |

### Filter Flag Naming Alternatives

| Option | Pros | Cons | Decision |
|--------|------|------|----------|
| `--project` only | Simple | Can't filter by bosun.stack | ❌ Rejected |
| `--stack` only | Matches bosun concept | Confusing with compose project | ❌ Rejected |
| **`--project` + `--stack`** | Clear semantics for each | Two flags to learn | ✅ Chosen |
