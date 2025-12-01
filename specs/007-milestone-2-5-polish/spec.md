# Feature Specification: Milestone 2.5 – Polish Job Model, Config Schema & Test Suite

**Feature Branch**: `007-milestone-2-5-polish`
**Created**: 2025-11-30
**Status**: Draft
**Input**: User description: "Milestone 2.5 – Polish Job Model, Config Schema & Test Suite - cleanup and polish pass on job model, job labels, planner, CLI and prune redundant tests"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Consistent Job Validation Experience (Priority: P1)

As a Bosun user, I want job validation to behave consistently whether I use discovery or config validation commands, so that I don't encounter unexpected differences in how jobs are validated.

**Why this priority**: The current duplication between `joblabels/discoverer.go` and `config/loader/job_validation.go` creates risk of divergent behavior. Users need a single, predictable validation experience.

**Independent Test**: Run `bosun plan list` and `bosun config validate` on the same Docker Compose stack with various label combinations – both should report the same validation errors/warnings.

**Acceptance Scenarios**:

1. **Given** a container with `bosun.job.enabled=true` but missing `bosun.job.name`, **When** I run either `bosun plan list` or `bosun config validate`, **Then** both produce the same "missing job name" error.
2. **Given** a volume with `bosun.job.attach=nonexistent-job`, **When** I run validation, **Then** I get an "orphaned volume" warning from both validation paths.
3. **Given** two containers defining conflicting schedules for the same job name, **When** I run validation, **Then** I get a consistent conflict error message.

---

### User Story 2 - Isolated Plan Testing via Project Filter (Priority: P1)

As a Bosun user running multiple Docker Compose stacks, I want to filter plan operations by compose project so that I only see jobs from the stack I'm interested in.

**Why this priority**: Integration tests currently can't run in isolation because `bosun plan list` shows all jobs on the Docker daemon. This blocks reliable CI and parallel test execution.

**Independent Test**: Deploy two separate compose stacks, use `bosun plan list --project stack-a`, verify only jobs from `stack-a` appear.

**Acceptance Scenarios**:

1. **Given** two running compose projects `app-a` and `app-b` with jobs, **When** I run `bosun plan list --project app-a`, **Then** I only see jobs defined in `app-a` containers.
2. **Given** a running compose project `my-app`, **When** I run `bosun plan show my-job --project my-app`, **Then** the plan only considers containers from `my-app`.
3. **Given** multiple projects, **When** I run `bosun plan list` without `--project`, **Then** I see all jobs (backward compatible).

---

### User Story 3 - Clear CLI Help and Branding (Priority: P2)

As a new Bosun user, I want CLI help text to clearly describe what Bosun does (backup job orchestration) so that I understand the tool's purpose immediately.

**Why this priority**: Current CLI says "Docker label management tool" which undersells the job orchestration and backup focus. Clear branding reduces user confusion.

**Independent Test**: Run `bosun --help` and subcommand help, verify descriptions match README vision and mention backup/job orchestration.

**Acceptance Scenarios**:

1. **Given** I run `bosun --help`, **When** I read the description, **Then** it mentions "job orchestrator" or "backup" and is consistent with README.
2. **Given** I run `bosun plan --help`, **When** I read the description, **Then** it clearly explains that plans describe backup job execution steps.
3. **Given** I run `bosun config --help`, **When** I read the description, **Then** it explains configuration validation for backup jobs.

---

### User Story 4 - Consistent Error Messages (Priority: P2)

As a Bosun user encountering errors, I want consistent, actionable error messages so that I can quickly understand what went wrong and how to fix it.

**Why this priority**: Inconsistent error formats make troubleshooting harder. Users benefit from a predictable error structure.

**Independent Test**: Trigger various error conditions across commands and verify consistent format.

**Acceptance Scenarios**:

1. **Given** I run `bosun plan show nonexistent-job`, **When** the job doesn't exist, **Then** the error message suggests running `bosun plan list` to see available jobs.
2. **Given** any command fails, **When** I see the error output, **Then** it follows a consistent format with context and actionable guidance.
3. **Given** I run a command with invalid flags, **When** it fails, **Then** the exit code is non-zero and consistent with other commands.

---

### User Story 5 - Reduced Test Redundancy (Priority: P3)

As a Bosun developer, I want a clear separation between unit and integration tests so that I can add new tests without duplicating coverage or slowing the suite.

**Why this priority**: Test duplication increases maintenance burden and risks drift between test layers. Clear responsibilities make the codebase easier to evolve.

**Independent Test**: Review test inventory, verify no scenario is fully tested in both unit and integration tests.

**Acceptance Scenarios**:

1. **Given** job discovery edge cases, **When** I look at the test suite, **Then** detailed edge cases are in unit tests, integration tests only verify E2E happy paths.
2. **Given** the planner logic, **When** I look at tests, **Then** planning decisions are unit-tested, integration tests only verify CLI output format.
3. **Given** I want to add a new edge case test, **When** I read `internal/testutil/doc.go`, **Then** I understand where the test belongs.

---

### User Story 6 - Up-to-Date Documentation (Priority: P3)

As a Bosun user, I want README and generated docs to accurately reflect current CLI commands and label semantics so that I can trust the documentation.

**Why this priority**: Stale docs erode user trust and cause support burden. Accurate docs enable self-service.

**Independent Test**: Compare README examples with actual CLI behavior, verify `go generate` produces current docs.

**Acceptance Scenarios**:

1. **Given** I follow README examples, **When** I run the commands, **Then** they work as documented.
2. **Given** I run `go generate ./internal/config/schema/...`, **When** I check `docs/config.md`, **Then** it includes job-related labels (`bosun.job.*`).
3. **Given** TODOs in the codebase, **When** I review `TODO.md`, **Then** items are either linked to issues or marked done.

---

### Edge Cases

- ~~What happens when a job name is defined in both config file and container labels?~~ → Resolved: labels take precedence per spec 002.
- ~~How does system handle containers that start/stop during a snapshot?~~ → Deferred: out of scope for polish milestone; see issue #99 for future robustness work.
- ~~What happens when `--project` filter matches no containers?~~ → Resolved: exit 0, empty list, stderr message.
- ~~How does mount mode validation handle case sensitivity (`RO` vs `ro`)?~~ → Resolved: case-insensitive, normalize to lowercase.

## Clarifications

### Session 2025-11-30

- Q: Should `--project` filter by Docker Compose project only, or also match `bosun.stack` labels? → A: Filter by `com.docker.compose.project` only; add separate `--stack` flag for `bosun.stack` labels.
- Q: When a filter (`--project` or `--stack`) matches no containers, what should the CLI do? → A: Return empty list, exit code 0, print "No jobs found" to stderr.
- Q: Should mount mode validation be case-insensitive (`RO` vs `ro`)? → A: Case-insensitive input, normalize to lowercase (`RO` → `ro`).
- Q: What exit codes should be used for validation failures vs runtime errors? → A: 0=success, 1=runtime error, 2=validation failure.
- Q: When a job name is defined in both config file and container labels, what should happen? → A: Labels take precedence per spec 002 (FR-011: defaults < file < env < labels).

## Requirements *(mandatory)*

### Functional Requirements

#### Job Validation Unification

- **FR-001**: System MUST define job enablement rules in exactly one canonical location and reuse them elsewhere.
- **FR-002**: System MUST define required fields (e.g., `bosun.job.name` when enabled) in exactly one location.
- **FR-003**: System MUST use shared constants for all label keys (`bosun.job.enabled`, `bosun.job.name`, etc.) – no string literals in validation code.
- **FR-004**: System MUST align default values (`DefaultSchedule`, `DefaultWorkerImage`, `DefaultMountMode`) between `internal/domain/jobs` and `internal/config/schema/job_labels.go`.
- **FR-004a**: Mount mode validation MUST be case-insensitive; input MUST be normalized to lowercase (`RO` → `ro`, `RW` → `rw`).

#### Project/Stack Filtering

- **FR-005**: System MUST implement `Selector.ProjectFilter` to filter Docker snapshots by `com.docker.compose.project` label (Docker Compose project name).
- **FR-005a**: System MUST implement `Selector.StackFilter` to filter Docker snapshots by `bosun.stack` label.
- **FR-006**: `bosun plan list` MUST support a `--project` flag to filter jobs by Docker Compose project.
- **FR-006a**: `bosun plan list` MUST support a `--stack` flag to filter jobs by `bosun.stack` label.
- **FR-007**: `bosun plan show` MUST support `--project` and `--stack` flags to limit plan scope.
- **FR-008**: `bosun labels snapshot` MUST support `--project` and `--stack` flags for consistency.
- **FR-009**: `--project` filtering MUST use Docker's native label filters (`com.docker.compose.project`) for efficiency.
- **FR-009a**: `--stack` filtering MUST filter by `bosun.stack` label value.
- **FR-009b**: When `--project` or `--stack` filter matches no containers, command MUST return exit code 0 with empty result and print "No jobs found" to stderr.

#### CLI UX

- **FR-010**: `bosun --help` MUST describe Bosun as a job orchestrator for backup automation, consistent with README.
- **FR-011**: All CLI commands MUST use consistent exit codes: 0=success, 1=runtime error (Docker unavailable, I/O failure), 2=validation failure (invalid labels, missing required fields).
- **FR-012**: Error messages MUST include context (which command, what entity) and suggest next steps where applicable.

#### Test Suite

- **FR-013**: System MUST document testing philosophy in `internal/testutil/doc.go` or `docs/testing.md`.
- **FR-014**: Integration tests MUST use project filtering to isolate test fixtures from other Docker state.
- **FR-015**: Each compose fixture in `internal/testutil/compose/` MUST have a header comment explaining its purpose.

#### Documentation

- **FR-016**: README MUST accurately reflect current CLI commands and label semantics.
- **FR-017**: `go generate ./internal/config/schema/...` MUST produce up-to-date docs including job labels.
- **FR-018**: Existing TODOs MUST be either converted to GitHub issues or marked as resolved.

### Key Entities *(include if feature involves data)*

- **JobLabelConfig**: Schema struct defining container-level job labels with defaults and validation rules. Source of truth for label keys and defaults.
- **JobVolumeConfig**: Schema struct defining volume-level job labels (attach, mount path, mount mode).
- **Selector**: Port struct for filtering label snapshots, enhanced with `ProjectFilter` for compose project isolation.
- **ValidationError**: Unified error type for job label validation, used by both config loader and job discoverer.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero duplication of job validation logic – validation rules exist in exactly one file and are imported elsewhere.
- **SC-002**: Integration tests pass reliably in CI without interference from other Docker containers on the same host.
- **SC-003**: `bosun plan list --project X` returns only jobs from compose project X within 2 seconds for up to 100 containers.
- **SC-004**: CLI help text for all commands references "backup", "job", or "orchestration" where appropriate.
- **SC-005**: Test suite runs in under 60 seconds for unit tests, under 5 minutes for integration tests.
- **SC-006**: README examples work without modification on a fresh Docker environment.
- **SC-007**: All TODOs in `TODO.md` are either converted to GitHub issues (with links) or removed as completed.
- **SC-008**: Compose fixture files each have a purpose comment, enabling new contributors to understand test setup.
