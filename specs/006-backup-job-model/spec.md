# Feature Specification: Job Model & Planning

**Feature Branch**: `006-backup-job-model`
**Created**: 2025-11-30
**Status**: Draft
**Input**: User description: "Milestone 2: Job Model & Planning - Introduce a first-class job concept and a planner that turns snapshot + config into an executable plan (no Docker side effects)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define Jobs via Labels (Priority: P1)

As a Docker Compose user, I want to define jobs using labels on my containers and volumes, so that Bosun can discover what tasks need to run without me writing configuration files.

**Why this priority**: This is the foundational capability - without job discovery from labels, no other features can work. Users must be able to declare their job intent via the familiar Docker label mechanism.

**Independent Test**: Can be fully tested by deploying a Docker Compose stack with `bosun.job.*` labels and verifying that Bosun discovers and parses the job definitions correctly.

**Acceptance Scenarios**:

1. **Given** a container with label `bosun.job.enabled=true` and `bosun.job.name=daily-maintenance`, **When** Bosun takes a snapshot, **Then** it discovers a job named "daily-maintenance"
2. **Given** a container with `bosun.job.schedule=0 2 * * *`, **When** Bosun parses the job, **Then** it recognizes the cron schedule for 2 AM daily
3. **Given** a container with `bosun.job.worker.image=restic/restic:latest`, **When** Bosun parses the job, **Then** it knows which worker container image to use
4. **Given** a volume with `bosun.job.attach=daily-maintenance:ro` label, **When** Bosun discovers jobs, **Then** that volume is included in the "daily-maintenance" job's attached volumes as read-only
5. **Given** multiple containers sharing `bosun.stack=myapp`, **When** Bosun discovers jobs, **Then** all containers are grouped under the same stack for coordinated stop/start

---

### User Story 2 - View Discovered Jobs (Priority: P1)

As a system administrator, I want to list all discovered jobs in my Docker environment, so that I can verify my label configuration is correct before running any jobs.

**Why this priority**: Users need immediate feedback on their label configuration. This is essential for debugging and validation without any side effects.

**Independent Test**: Can be tested by running `bosun plan list` and verifying the output matches expected job definitions.

**Acceptance Scenarios**:

1. **Given** Docker containers with job labels, **When** I run `bosun plan list`, **Then** I see a list of all discovered jobs with their names
2. **Given** no containers with job labels, **When** I run `bosun plan list`, **Then** I see an empty list with an informative message
3. **Given** multiple stacks with different jobs, **When** I run `bosun plan list`, **Then** each job is listed with its associated stack

---

### User Story 3 - Preview Execution Plan (Priority: P1)

As a system administrator, I want to see the exact steps Bosun would take to execute a job, so that I can verify the plan before any containers are stopped or actions are taken.

**Why this priority**: Dry-run capability is critical for user confidence. Users must understand exactly what will happen before granting Bosun permission to stop services.

**Independent Test**: Can be tested by running `bosun plan show <job>` and verifying the output contains the correct ordered steps.

**Acceptance Scenarios**:

1. **Given** a job named "daily-maintenance" targeting stack "myapp", **When** I run `bosun plan show daily-maintenance`, **Then** I see the ordered steps: stop stack → run worker with volumes → (future: restart stack)
2. **Given** a job with multiple target volumes, **When** I view the plan, **Then** each volume attachment is listed in the worker step
3. **Given** a job targeting a stack with 3 containers, **When** I view the plan, **Then** the stop step lists all 3 containers that will be stopped
4. **Given** I run `bosun plan show daily-maintenance --format=json`, **When** the command completes, **Then** the output is valid JSON suitable for scripting
5. **Given** I run `bosun plan show daily-maintenance --format=yaml`, **When** the command completes, **Then** the output is valid YAML for human readability

---

### User Story 4 - Validate Job Configuration (Priority: P2)

As a user, I want Bosun to validate my job labels and report configuration errors, so that I catch mistakes before attempting to run the job.

**Why this priority**: Extends existing validation infrastructure to cover job-specific labels, providing early feedback on misconfiguration.

**Independent Test**: Can be tested by deploying stacks with invalid job labels and running `bosun config validate` to see helpful error messages.

**Acceptance Scenarios**:

1. **Given** a container with `bosun.job.enabled=yes` (invalid boolean), **When** I run `bosun config validate`, **Then** I see a type error for that label
2. **Given** a container with `bosun.job.schedule=invalid-cron`, **When** I run `bosun config validate`, **Then** I see a schedule parsing error
3. **Given** a job with `bosun.job.enabled=true` but no `bosun.job.name`, **When** I run `bosun config validate`, **Then** I see a "required field missing" error
4. **Given** a valid job configuration, **When** I run `bosun config validate`, **Then** validation passes with no errors

---

### Edge Cases

- What happens when a job references a stack with no running containers?
  - The planner should still discover the job but note that no containers will be stopped
- What happens when multiple jobs target the same stack?
  - Each job should be independent; the user is responsible for schedule coordination
- What happens when a volume has no `bosun.stack` label but a container does?
  - Volumes are matched to stacks via Docker Compose's `com.docker.compose.project` metadata as fallback
- What happens when `bosun.stack` and `com.docker.compose.project` differ?
  - `bosun.stack` label takes precedence (explicit user intent overrides automatic detection)
- How does the system handle orphaned volumes not attached to any container?
  - They can be included via explicit `bosun.stack` labels on the volume itself
- What happens when a container has `bosun.job.enabled=true` but the stack has no volumes?
  - The job is valid but the plan will have an empty volume attachment list (warning may be shown depending on job type)
- What happens when two containers define the same job with conflicting field values (e.g., different schedules)?
  - Validation error listing both containers and the conflicting field
- Can a container participate in multiple jobs?
  - Yes, a container can have labels for multiple jobs (e.g., `bosun.job.name=daily-backup,weekly-archive` as a list)
- What happens when containers from different stacks contribute to the same job?
  - All contributing stacks will be stopped; volumes from all stacks are included in the worker
- What happens when a volume has `bosun.job.attach=nonexistent-job`?
  - Validation warning (not error) - the volume is orphaned but doesn't block other jobs
- What is the default mount mode for attached volumes?
  - Read-only (`ro`) for safety; user must explicitly specify `rw` if write access is needed
- What happens when a job stops a container that has dependents not in the job?
  - Validation error listing the orphaned dependents. User must either add the dependents to the job or remove the dependency source.
- What if `com.docker.compose.depends_on` label is missing (non-Compose containers)?
  - No dependency validation is performed for that container; user is responsible for ensuring safe stop order

## Requirements *(mandatory)*

### Functional Requirements

#### Domain Model

- **FR-001**: System MUST define a `Job` entity representing a job with: name, schedule, target containers, worker image, and options
- **FR-002**: System MUST define an `ExecutionPlan` entity representing ordered steps to execute a job
- **FR-003**: System MUST define `PlanStep` as an abstraction for individual actions (stop stack, run worker, etc.)
- **FR-004**: Domain entities MUST be pure Go types with no external dependencies (Docker SDK, etc.)

#### Label Schema

- **FR-005**: System MUST recognize `bosun.job.enabled` label (boolean) on containers only to mark a container as a job definition; `bosun.job.*` labels on volumes or networks MUST be ignored
- **FR-006**: System MUST recognize `bosun.job.name` label (string, **required**) to identify the job
- **FR-007**: System MUST recognize `bosun.job.schedule` label (cron expression string, optional, default: `0 0 * * *` daily at midnight) to define when the job runs
- **FR-008**: System MUST recognize `bosun.job.worker.image` label (string, optional, default: built-in minimal image) to specify the worker container image
- **FR-009**: System MUST recognize `bosun.job.attach` label on **volumes** (not containers) to attach the volume to a job. Format: `<job-name>` or `<job-name>:<mode>` where mode is `ro` (read-only, default) or `rw` (read-write). Multiple jobs can be specified as comma-separated list.
- **FR-010**: System MUST recognize `bosun.stack` label (string) to group entities into logical stacks for coordinated operations

#### Planner

- **FR-011**: System MUST provide a pure function that transforms a Snapshot and Config into a list of ExecutionPlans
- **FR-012**: Planner MUST be deterministic: same inputs MUST produce identical outputs
- **FR-013**: Planner MUST NOT make any Docker API calls or cause side effects
- **FR-014**: Plan MUST specify which containers need to be stopped (only those that declared the job, not entire stack by default)
- **FR-014a**: Planner MUST validate container stop list against Compose dependency graph (from `com.docker.compose.depends_on` label). If stopping containers would orphan dependents not in the job, validation MUST fail.
- **FR-014b**: When a job includes all containers in a stack, the plan SHOULD note that `docker compose stop` can be used to respect dependency order.
- **FR-015**: Plan MUST specify which volumes need to be attached to the worker container (from volumes with `bosun.job.attach` label)
- **FR-016**: Plan MUST specify the worker container image and command to execute
- **FR-017**: Plans MUST be serializable to JSON and YAML formats

#### CLI Commands

- **FR-018**: System MUST provide `bosun plan list` command to display all discovered jobs
- **FR-019**: System MUST provide `bosun plan show <job-name>` command to display the computed plan for a specific job
- **FR-020**: `bosun plan show` MUST support `--format` flag with `json`, `yaml`, and `text` (default) options
- **FR-021**: Both commands MUST exit with code 0 on success, non-zero on error
- **FR-022**: Commands MUST provide helpful error messages when Docker is unavailable or no jobs are found

#### Validation

- **FR-023**: System MUST validate job labels using the existing validation infrastructure
- **FR-024**: System MUST report type errors for invalid label values (e.g., non-boolean for `enabled`)
- **FR-025**: System MUST report missing required field when `bosun.job.enabled=true` but no `bosun.job.name` is specified
- **FR-026**: System MUST report validation error when multiple containers contribute to the same job with conflicting values for the same field
- **FR-027**: System MUST support a container participating in multiple jobs via list syntax in `bosun.job.name` (e.g., `daily-backup,weekly-archive`)

### Key Entities

- **Job**: Represents a discovered job definition
  - Name: Human-readable identifier (unique globally, **required**)
  - Schedule: Cron expression for execution timing (default: `0 0 * * *` daily at midnight)
  - TargetContainers: List of containers that participate in this job (only these will be stopped, not entire stack)
  - TargetStacks: Derived list of stacks involved (for grouping/display; actual stop is per-container)
  - WorkerImage: Container image for running the job (default: built-in minimal image that logs success)
  - AttachVolumes: List of volumes to mount in the worker, discovered from volumes with `bosun.job.attach=<this-job>` label. Each entry includes volume name and mount mode (ro/rw).
  - Options: Job-specific configuration overrides
  - SourceContainers: List of container IDs that contributed to this job's configuration (for traceability)

- **ExecutionPlan**: Ordered sequence of steps to execute a job
  - JobName: Reference to the originating job
  - Steps: Ordered list of PlanStep
  - CreatedAt: Timestamp of plan generation

- **PlanStep**: Individual action in an execution plan
  - Type: Kind of action (StopStack, RunWorker, etc.)
  - Description: Human-readable description
  - Details: Type-specific data (container IDs, volume names, etc.)

- **Stack**: Logical grouping of Docker entities
  - Name: Stack identifier (from `bosun.stack` or `com.docker.compose.project`)
  - Containers: List of container IDs in this stack
  - Volumes: List of volumes associated with this stack
  - Networks: List of networks associated with this stack

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can define a complete job using only Docker labels (no external config files required)
- **SC-002**: `bosun plan list` returns discovered jobs within 5 seconds for environments with up to 100 containers
- **SC-003**: `bosun plan show` generates a complete plan within 2 seconds for a stack with up to 20 volumes
- **SC-004**: Planner produces identical output when given the same snapshot, verified by JSON comparison
- **SC-005**: 100% of existing validation test patterns work with new job labels
- **SC-006**: Unit tests achieve 80%+ code coverage for domain model and planner
- **SC-007**: Users can understand the plan output without documentation (self-describing step descriptions)
- **SC-008**: No Docker API calls are made during plan list or plan show commands (except for snapshot retrieval)

## Clarifications

### Session 2025-11-30

- Q: When both `bosun.stack` label and `com.docker.compose.project` metadata are present, which takes precedence? → A: `bosun.stack` label always overrides `com.docker.compose.project` (explicit over implicit)
- Q: Can job definitions (`bosun.job.*` labels) be placed on any entity type or only containers? → A: Job definitions are valid only on containers
- Q: What is the relationship between jobs and containers? → A: Multiple containers can contribute to the same job (merged by `bosun.job.name`). A container can participate in multiple jobs. The job targets all containers that declare it, not necessarily the whole stack.
- Q: How are merge conflicts handled when multiple containers set the same job field differently? → A: Validation error (strict mode) - forces explicit coordination between services
- Q: Is Bosun specifically a backup tool? → A: No. Bosun is a general-purpose job runner for Docker Compose stacks. Backup is one use case, but jobs can perform any task (maintenance, migrations, health checks, etc.). Domain entities use generic names: `Job` and `ExecutionPlan` (not BackupJob/BackupPlan)
- Q: Which job fields are required vs optional? → A: Only `bosun.job.name` is required. Defaults: `worker.image` defaults to a minimal "hello-world" style image (confirms execution, can serve as base for custom jobs); `schedule` defaults to daily at midnight (`0 0 * * *`); `attach.volume` is optional (empty by default)
- Q: What should the default worker image be? → A: A Bosun-provided minimal image that logs success and can serve as a base for custom jobs. For this milestone, the image is local/built-in; publishing to a registry (e.g., `ghcr.io/simone-viozzi/bosun-worker`) is deferred to a later milestone.
- Q: How are volumes discovered and attached to jobs? → A: No auto-discovery. Volumes must explicitly declare which job(s) they attach to via `bosun.job.attach` label on the volume itself (not on containers). The label format is `bosun.job.attach=<job-name>` or `bosun.job.attach=<job-name>:ro` / `bosun.job.attach=<job-name>:rw` to specify mount mode (default: `ro` for safety).
- Q: When a job targets specific containers, should it stop only those containers or the entire stack? → A: Stop only the containers that declared the job (fine-grained control). However, the planner MUST validate against the Compose dependency graph (from `com.docker.compose.depends_on` label). If stopping a container would orphan dependents not included in the job, validation MUST fail with an error listing the missing dependents. When stopping multiple containers, use `docker compose stop` to respect dependency order.
- Q: How do we get the Compose dependency graph at runtime? → A: Docker Compose v2 stores `depends_on` in the `com.docker.compose.depends_on` container label. Format: `<service>:<condition>:<required>,...` (e.g., `db:service_started:true,redis:service_healthy:false`). This allows dependency validation without parsing compose.yml.

## Assumptions

- Cron schedule parsing will use a standard Go cron library (specific library selection deferred to implementation)
- Worker container execution is out of scope for this milestone (Milestone 3+)
- Stack restart after backup completion is out of scope for this milestone
- Schedule-based automatic execution is out of scope for this milestone
- Users are familiar with Docker Compose project naming conventions
- The existing snapshot mechanism provides sufficient metadata for stack discovery

## Out of Scope

- Actually stopping/starting Docker Compose stacks
- Actually running worker containers
- Scheduling / cron execution
- Job result verification or monitoring
- Notification systems
- Multi-host / Docker Swarm support
