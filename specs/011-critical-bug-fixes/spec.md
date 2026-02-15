# Feature Specification: Milestone 3.75 — Critical Bug Fixes

**Feature Branch**: `011-critical-bug-fixes`
**Created**: 2026-02-15
**Status**: Draft
**Input**: GitHub Milestone 3.75 ([#163](https://github.com/simone-viozzi/bosun/issues/163)) and sub-issues [#153](https://github.com/simone-viozzi/bosun/issues/153), [#144](https://github.com/simone-viozzi/bosun/issues/144), [#161](https://github.com/simone-viozzi/bosun/issues/161)

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Worker Logs Are Readable (Priority: P0)

As a user running `bosun job run`, I expect the worker output to display clean, human-readable text. Currently, worker logs are corrupted with binary header prefixes (`\x01\x00\x00\x00...`) because Docker's multiplexed log stream is copied raw instead of being properly demultiplexed.

**Why this priority**: This is the most visible and disruptive bug. Every user running a job sees garbled output, making bosun appear broken. It directly undermines trust in the tool.

**Independent Test**: Run any job via `bosun job run` and verify the output contains no binary header bytes — only clean stdout/stderr text.

**Acceptance Scenarios**:

1. **Given** a worker container producing stdout output, **When** the user runs `bosun job run`, **Then** the output displays clean text with no binary prefixes.
2. **Given** a worker container producing stderr output, **When** the user runs `bosun job run`, **Then** stderr output displays correctly and is distinguishable from stdout.
3. **Given** a worker container producing mixed stdout and stderr, **When** the user runs `bosun job run`, **Then** both streams display correctly without interleaving corruption.
4. **Given** a worker container that produces only stdout (no stderr), **When** the user views the logs, **Then** logs display correctly without errors.
5. **Given** the test helper (`DumpLogs`) is used during development, **When** a developer runs tests, **Then** container logs in test output are also clean and readable.

---

### User Story 2 — Worker Errors Are Visible (Priority: P1)

As a user, when a worker encounters an error (e.g., container fails to start, log streaming fails, container inspection fails), I expect to see a meaningful error message rather than silent failure. Currently, the worker runner silently swallows errors, making it impossible to diagnose issues.

**Why this priority**: Silent error swallowing is the second most impactful reliability issue. Users have no way to diagnose problems when errors are dropped, leading to frustration and wasted debugging time.

**Independent Test**: Simulate a container failure and verify the user sees a descriptive error message indicating what went wrong.

**Acceptance Scenarios**:

1. **Given** log streaming encounters an error, **When** the worker runner processes it, **Then** the error is logged or returned to the caller with context (which container, which operation).
2. **Given** container inspection fails after a stop, **When** the worker runner tries to get the exit code, **Then** the error is reported rather than silently returning a hardcoded value.
3. **Given** a container exits with a signal (e.g., SIGKILL), **When** the worker runner reports the exit code, **Then** the exit code has a clear, documented meaning (e.g., "killed by signal 9") rather than being an unexplained magic number.

---

### User Story 3 — Clean Repository Root (Priority: P2)

As a developer working on bosun, I expect the repository root to be clean and uncluttered. Currently, coverage output files (`.html`, `.out`, `.txt`) accumulate in the root directory, making it harder to navigate the project.

**Why this priority**: This is a developer-experience improvement. While the files are gitignored, they still clutter the local workspace and make directory listings noisy.

**Independent Test**: Run `make test` or coverage commands and verify that all generated coverage files are written to a dedicated subdirectory rather than the repository root.

**Acceptance Scenarios**:

1. **Given** a developer runs the test suite with coverage, **When** coverage files are generated, **Then** they are written to a dedicated subdirectory (not the repository root).
2. **Given** existing coverage files in the repository root, **When** the migration is complete, **Then** the build system references the new location and the root is clean.
3. **Given** CI or local scripts reference coverage file paths, **When** the paths change, **Then** all references are updated consistently.

---

### Edge Cases

- What happens when the Docker container's TTY setting is enabled (vs. disabled) — does log demultiplexing behave correctly in both modes?
- What happens when a container exits before log streaming begins?
- What happens when log streaming is interrupted mid-stream (e.g., context cancellation)?
- What happens when container inspection fails for reasons other than the container being gone?
- What happens when the coverage output directory does not exist at test runtime?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Worker log output MUST be properly demultiplexed so that users see clean text without binary header bytes.
- **FR-002**: The test helper log output MUST also be properly demultiplexed, consistent with production code.
- **FR-003**: Worker runner MUST log or return errors from log streaming rather than silently dropping them.
- **FR-004**: Worker runner MUST log or return errors from container inspection rather than returning hardcoded values.
- **FR-005**: Exit codes from container operations MUST use named constants with documented meanings (e.g., 128+signal conventions).
- **FR-006**: All coverage output files MUST be written to a dedicated subdirectory rather than the repository root.
- **FR-007**: Build scripts and CI configuration MUST reference the new coverage file paths after migration.
- **FR-008**: Existing gitignore rules MUST be updated to reflect the new coverage file location.

### Key Entities

- **Worker Runner**: The component responsible for starting containers, streaming logs, and reporting results. It is the primary target of the log corruption and error handling fixes.
- **Container Logs**: The multiplexed stdout/stderr stream from Docker containers that must be demultiplexed before display.
- **Exit Codes**: Numeric values returned by container operations that indicate how a process terminated. Must be documented and use named constants.
- **Coverage Files**: Test coverage output files (`.html`, `.out`, `.txt`) generated during test runs.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of worker log output is clean, human-readable text — no binary header bytes appear in any output.
- **SC-002**: All errors in the worker runner path are either logged or returned — no silent error drops exist (verifiable via code inspection: no bare `return` after error checks).
- **SC-003**: All numeric exit codes in the worker runner have corresponding named constants with documentation.
- **SC-004**: Zero coverage-related files exist in the repository root after running the test suite.
- **SC-005**: All existing tests (unit and integration) pass after the fixes.
- **SC-006**: M3 functionality is production-ready, unblocking M4 implementation.

## Scope Boundaries

### In-scope

- Fix corrupted worker logs ([#153](https://github.com/simone-viozzi/bosun/issues/153), P0)
- Fix worker error handling and magic exit codes ([#144](https://github.com/simone-viozzi/bosun/issues/144), P1)
- Move coverage files to subdirectory ([#161](https://github.com/simone-viozzi/bosun/issues/161), P2)

### Out-of-scope

- M4 features (scheduling, daemon mode) — implementation blocked by this milestone
- Architecture refactoring (#141) — deferred
- Test harness improvements (#145) — deferred
- Complexity reduction (#146-#148, #154-#156) — deferred
- Adding retry/backoff logic to the worker runner
- Changing worker lifecycle behavior

## Risks & Assumptions

### Risks

- **Risk**: Additional bugs may be discovered during fixes. **Mitigation**: Add to this milestone or defer to M3.8 if scope expands.
- **Risk**: Changing log demultiplexing may impact test infrastructure. **Mitigation**: Update both production and test helper code together.

### Assumptions

- The demultiplexing library is already available as a transitive dependency (no new external dependencies needed).
- The coverage file migration requires only build-script/Makefile changes and gitignore updates, not code changes.
- Exit code conventions follow standard Unix signal mapping (128 + signal number).

## Dependencies & Ordering

1. **#153** — Corrupted worker logs (P0, critical user-facing bug) — must be fixed first
2. **#144** — Worker error handling (P1, reliability improvement) — builds on #153 since both touch the same code paths
3. **#161** — Coverage files pollution (P2, developer experience) — independent, can be done in parallel

### Cross-links

- **Parallel track**: #108 (M4 concurrency research) can proceed while M3.75 is in progress
- **Blocks**: M4 implementation cannot begin until M3.75 is complete
- **Blocks**: M4 implementation also blocked by #108 completion
