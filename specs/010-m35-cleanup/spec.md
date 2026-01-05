# Feature Specification: Milestone 3.5 - Post-M3 Cleanup & Bug Fixes

**Feature Branch**: `010-m35-cleanup`
**Created**: 2026-01-05
**Status**: Draft
**Input**: User description: "Milestone 3.5: Post-M3 Cleanup & Bug Fixes - Address critical bugs and incomplete tasks from M3 (Job Execution MVP)"

## Related Documentation

> **Important**: This spec consolidates issues and decisions already documented in project memories.
> For detailed context, evidence, and rationale, refer to:
>
> - **[wip_m35.md](../../.serena/memories/wip_m35.md)** — Issue descriptions from GitHub (#133–#148)
> - **[wip_smell_milestone3.md](../../.serena/memories/wip_smell_milestone3.md)** — Smell analysis, decisions, and implementation guidance

## Summary

Address critical bugs and incomplete tasks discovered after M3 (Job Execution MVP) completion. This cleanup milestone ensures M3 functionality is production-ready before starting M4 (Scheduling Engine).

**Goal**: All M3 bugs fixed, documentation complete, users can reliably use `bosun job run` and understand how to use it.

## User Scenarios & Testing

### User Story 1 - Validate Config Without False Positives (Priority: P0)

As an operator, I want `bosun config validate` to succeed on valid compose files without flagging unknown labels as errors, so that I can confidently validate my configuration.

**Why this priority**: Config validation failing on valid files (#136) blocks basic workflows and creates distrust in the tool.

**Independent Test**: Run `bosun config validate --project joblabels` on test compose files — should pass without errors for valid labels, show warnings for unknown `bosun.*` labels.

**Acceptance Scenarios**:

1. **Given** a compose file with valid `bosun.*` labels, **When** I run `bosun config validate`, **Then** validation passes with no errors
2. **Given** a compose file with unknown `bosun.*` labels (e.g., `bosun.custom`), **When** I run `bosun config validate`, **Then** validation passes but shows a warning about unknown keys
3. **Given** `--strict` mode enabled, **When** compose file has unknown `bosun.*` labels, **Then** validation fails with clear error messages

**Decision Applied**: Lenient/warning mode for unknown keys (Decision #5 in `wip_smell_milestone3.md`)

---

### User Story 2 - See Complete Execution Plan (Priority: P0)

As an operator, I want `bosun plan show` to display ALL execution steps including container restart, so that I understand exactly what will happen when I run a job.

**Why this priority**: Missing restart step (#135) means the plan preview doesn't match actual execution — violates "what you see is what runs" principle.

**Independent Test**: Run `bosun plan show <job> --format json` — output should include `start_containers` step after `run_worker` step.

**Acceptance Scenarios**:

1. **Given** a job that stops containers, **When** I run `bosun plan show <job>`, **Then** I see stop → worker → start steps in order
2. **Given** `--dry-run` mode, **When** I preview execution, **Then** the displayed steps match what `job run` would actually execute
3. **Given** a plan with JSON format, **When** I view the plan, **Then** each step has `type`, `description`, and relevant metadata

**Current Problem** (from `wip_smell-design-m3` Finding #1):
- Planner creates `ExecutionPlan` with steps but explicitly omits `start_containers` step (TODO comment)
- Executor ignores `plan.Steps` entirely and uses hardcoded sequence: stop → worker → defer restart
- This duplicates stop/start semantics and causes divergence between preview and actual execution

**Decision Applied**: ExecutionPlan is authoritative — plan drives execution (Decision #4 in `wip_smell_milestone3.md`)

---

### User Story 3 - Clean Generic Terminology (Priority: P1)

As an operator, I want bosun to use generic job terminology (not "backup"-specific), so that I understand bosun is a general-purpose job orchestrator that can be used for backups, maintenance, migrations, health checks, and any other container-based tasks.

**Why this priority**: Confusing terminology (#134, #140) implies bosun is backup-specific rather than a generic job orchestrator. This limits perceived applicability and creates cognitive confusion throughout the codebase.

**Independent Test**:
- Run `bosun config validate --print` — output should show `Enabled`, not `BackupEnabled`
- Run `grep -r "backup" internal/ --include="*.go" | grep -v _test.go` — should return only example/doc contexts

**Acceptance Scenarios**:

1. **Given** I run config validation with `--print`, **When** viewing output, **Then** property names use generic terms (`Enabled` not `BackupEnabled`)
2. **Given** CLI help text, **When** I read command descriptions, **Then** they describe "jobs" not "backup jobs"
3. **Given** package documentation (`doc.go` files), **When** reading them, **Then** they describe generic job orchestration
4. **Given** error messages from executor/planner, **When** failures occur, **Then** messages refer to "job" not "backup"

**Scope of Cleanup** (from `wip_m35.md` Issue #140):
- Config schema: `BackupEnabled` → `Enabled`
- Label schema: `bosun.volume.backupEnabled` → `bosun.volume.enabled`
- Comments in: `job_labels.go`, `run.go`, `doc.go` files
- CLI descriptions in: `plan_list.go`, `plan_show.go`, `job_run.go`
- Port docs: `ports/doc.go`
- Documentation: `docs/config.md`

**Note**: Test files MAY retain "backup" as a valid use-case example (e.g., `"daily-backup"` job name).

**Decision Applied**: Rename now as breaking change (Decision #3 in `wip_smell_milestone3.md`)

---

### User Story 4 - Execute Jobs via Plan Interpreter (Priority: P1)

As an operator, I want the job executor to interpret the execution plan (not hardcode the flow), so that DryRun and actual execution are guaranteed to match.

**Why this priority**: Interface mismatch (#143) and executor design issues (#142) create technical debt that complicates testing and causes preview/execution divergence.

**Independent Test**: Unit tests for executor with mocked ports pass; executor interprets plan steps.

**Acceptance Scenarios**:

1. **Given** `JobExecutor` interface, **When** I check the implementation, **Then** all methods are implemented (no stubs returning "not implemented")
2. **Given** executor constructor, **When** I create an executor, **Then** all parameters are used (no unused `discoverer` param)
3. **Given** a job to execute, **When** executor runs, **Then** it iterates over `plan.Steps` and executes each in order
4. **Given** an execution plan with stop → worker → start, **When** executor runs, **Then** each step result is recorded individually

**Current Problem** (from `wip_smell-design-m3` Findings #1, #2):
- `Executor.New` accepts `JobDiscoverer` but never uses it
- `Execute(ctx, jobName)` returns "not implemented" error
- `ExecuteJob` hardcodes: `compose.StopStack()` → `worker.Run()` → `defer compose.StartStack()`
- Plan is only used for logging, NOT for driving execution

**Target Architecture**:
```
Planner.Plan(job) → ExecutionPlan{Steps: [stop, worker, start]}
                           ↓
Executor.Execute(job) → for step := range plan.Steps { executeStep(step) }
```

**Decision Applied**:
- Option B — `ExecuteJob(ctx, job)` with external discovery (Decision #2)
- Plan-is-source-of-truth — executor interprets plan steps (Decision #4)

---

### User Story 5 - Clean CLI Error Output (Priority: P3)

As a user, I want CLI errors to appear only once, so that I can quickly understand and fix issues.

**Why this priority**: Duplicate errors (#133) are annoying but don't block functionality.

**Independent Test**: Run `bosun invalidcommand` — error message should appear exactly once.

**Acceptance Scenarios**:

1. **Given** an invalid command, **When** I run bosun, **Then** the error message appears exactly once
2. **Given** any CLI error, **When** it's displayed, **Then** it includes helpful context (suggestion, help reference)

---

### User Story 6 - Learn Job Execution from README (Priority: P3)

As a new user, I want the README to document basic job execution, so that I can get started without reading source code.

**Why this priority**: Documentation (#120) is important for adoption but existing users can work without it.

**Independent Test**: Follow README instructions on a fresh setup — should be able to run a job successfully.

**Acceptance Scenarios**:

1. **Given** the README, **When** I read the "Running Jobs" section, **Then** I find clear usage instructions for `bosun job run`
2. **Given** README examples, **When** I copy-paste them, **Then** they work with the test compose files
3. **Given** README documentation, **When** I look for `--dry-run`, **Then** I find explanation and examples

---

### Edge Cases

- What happens when a job has no containers to stop? (Worker-only execution)
- How does system handle worker image not found? (Clear error before execution starts)
- What if compose project doesn't exist? (Validation error with helpful message)
- How does restart behave if some containers failed to stop? (Partial restart with error reporting)

## Requirements

### Functional Requirements

#### Config Validation (Issues #136, #139)

- **FR-001**: System MUST log warnings (not errors) for unknown `bosun.*` label keys by default
- **FR-002**: System MUST provide `--strict` flag to treat unknown keys as errors
- **FR-003**: Test compose files MUST use valid label schema (remove `bosun.other`, fix `bosun.network`)

#### Execution Plan Completeness (Issues #135, #142)

- **FR-004**: Planner MUST include `start_containers` step when containers are stopped
- **FR-005**: Executor MUST execute steps from the plan (not hardcode execution flow)
- **FR-006**: DryRun output MUST match actual execution flow exactly

#### Terminology Cleanup (Issues #134, #140)

- **FR-007**: Config schema MUST rename `BackupEnabled` to `Enabled` (or `JobEnabled`)
- **FR-008**: Label schema MUST rename `bosun.volume.backupEnabled` to `bosun.volume.enabled`
- **FR-009**: Production code MUST NOT contain "backup" terminology except in examples/docs

#### Interface Cleanup (Issue #143)

- **FR-010**: `JobExecutor` interface MUST match implementation signature
- **FR-011**: Executor MUST NOT have unused constructor parameters
- **FR-012**: Executor MUST NOT have stub methods returning "not implemented"

#### CLI Output (Issue #133)

- **FR-013**: CLI errors MUST appear exactly once (no duplicate printing)

#### Documentation (Issue #120)

- **FR-014**: README MUST have "Running Jobs" section with usage examples
- **FR-015**: README MUST document `--dry-run` flag
- **FR-016**: README examples MUST be tested and accurate

### Key Entities

Refer to existing domain types in `wip_smell_milestone3.md` (Decisions section):

- **ExecutionPlan**: Steps to execute for a job (now authoritative)
- **PlanStep**: Individual step with type and metadata (gains `start_containers` type)
- **JobExecutor**: Port interface for job execution (signature updated)

## Success Criteria

### Measurable Outcomes

- **SC-001**: `bosun config validate` passes on all test compose files without errors
- **SC-002**: `bosun plan show` displays all execution steps including restart (verified via JSON output)
- **SC-003**: `grep -r "BackupEnabled" internal/` returns zero matches in production code
- **SC-004**: `grep -r "internal/adapters" internal/cmd/` returns zero matches (CLI doesn't import adapters)
- **SC-005**: All unit and integration tests pass
- **SC-006**: README "Running Jobs" section exists with working examples

## Assumptions

- Breaking changes to config property names are acceptable (no backward compatibility required yet)
- Deprecation warnings for old label names are sufficient migration path
- M3 functionality (job execution) is otherwise complete and stable
- Comprehensive documentation deferred to M6

## Out of Scope

- M4 features (scheduling, daemon mode)
- M5 features (observability, robustness)
- M6 features (comprehensive documentation)
- Per-step retry policies in execution plan (deferred)
- Complete hexagonal architecture refactoring (only CLI layer addressed)

## Implementation Notes

> **For detailed implementation guidance, decisions, and code locations, see:**
>
> - `wip_smell_milestone3.md` — Decisions #1-#5 with rationale and action items
> - `wip_m35.md` — Issue descriptions with specific file/line references

### Priority Order

1. **P0**: #136/#139 (config validation) + #135/#142 (execution plan) — blocking basic usage
2. **P1**: #134/#140 (terminology) + #143 (interface) — technical debt
3. **P3**: #133 (duplicate errors) + #120 (docs) — quality of life

### Key Code Locations

| Issue | Primary Files |
|-------|---------------|
| #139 | `internal/config/loader/loader.go:125` |
| #140 | `internal/config/schema/config_v1.go:30` |
| #141 | `internal/cmd/job_run.go` (and other cmd files) |
| #142 | `internal/app/planner/planner.go:84` |
| #143 | `internal/app/executor/executor.go:24` |

### Relevant Memories by User Story

#### US1 - Config Validation (Issues #136, #139)
| Memory | Relevance |
|--------|----------|
| `pkg_config_loader` | `FromLabels` validation, `AddUnknownKey` error handling |
| `pkg_config_schema` | `Spec` definition, `FieldSpec` for valid keys |
| `wip_smell-design-m3` | Finding #12: unknown key rejection policy |
| `wip_smell-general-m3` | Finding #3: strict unknown-key rejection |

#### US2 - Complete Execution Plan (Issues #135, #142)
| Memory | Relevance |
|--------|----------|
| `pkg_app_planner` | `Plan()` generates steps, missing `StartContainers` (TODO) |
| `pkg_domain_jobs` | `ExecutionPlan`, `PlanStep`, `StepTypeStartContainers` (exists but unused) |
| `wip_smell-design-m3` | Finding #1: planner vs executor mismatch |
| `wip_smell-design-m3` | Finding #4: execution plan incompleteness |

#### US3 - Generic Terminology (Issues #134, #140)
| Memory | Relevance |
|--------|----------|
| `arch_overview` | States "Not a backup tool" — workers can perform any task |
| `pkg_config_schema` | `VolumeConfig.BackupEnabled` field location |
| `wip_smell-design-m3` | Finding #3: `BackupEnabled` naming confusion |
| `wip_smell-general-m3` | Finding #1: confusing property name |

#### US4 - Executor Interface (Issues #142, #143)
| Memory | Relevance |
|--------|----------|
| `pkg_app_executor` | Current execution flow (hardcoded, not plan-driven) |
| `pkg_ports` | `JobExecutor` interface contract |
| `wip_smell-design-m3` | Finding #2: API mismatch, unused `discoverer` |
| `wip_smell-design-m3` | Finding #5: `ExecuteJob` complexity (~160 lines) |
| `wip_smell-layering-m3` | CLI imports adapters directly |

#### US5 - CLI Errors (Issue #133)
| Memory | Relevance |
|--------|----------|
| `pkg_cli` | Command structure, exit codes |
| `wip_smell-general-m3` | Finding #2: duplicate CLI error printing |

#### US6 - Documentation (Issue #120)
| Memory | Relevance |
|--------|----------|
| `pkg_cli` | `bosun job run` flags and behavior |
| `arch_overview` | High-level architecture for README context |
