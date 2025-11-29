# Feature Specification: CLI Config Validate Command

**Feature Branch**: `003-cli-config-validate`
**Created**: 2025-11-29
**Status**: Draft
**Input**: User description: "CLI config validate command - A bosun config validate CLI command that validates configuration from labels and files, prints effective config, and fails fast on invalid or unknown keys"
**Related Issues**: #60

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Validate Configuration Before Deployment (Priority: P1)

As a Bosun user, I want to run `bosun config validate` to check my Docker labels and config file for errors before deploying, so that I can catch configuration mistakes early without affecting running services.

**Why this priority**: This is the primary use case - users need a safe, read-only way to verify their configuration. Without this, users must deploy and hope for the best.

**Independent Test**: Can be tested by running `bosun config validate` against a Docker environment with valid labels and verifying it exits 0 with no errors.

**Acceptance Scenarios**:

1. **Given** Docker containers with valid `bosun.*` labels, **When** I run `bosun config validate`, **Then** the command exits 0 with a success message
2. **Given** a valid config file and valid labels, **When** I run `bosun config validate`, **Then** validation passes and shows merged config summary
3. **Given** no config file and no labels, **When** I run `bosun config validate`, **Then** validation passes with default configuration

---

### User Story 2 - Fail Fast on Invalid Configuration (Priority: P1)

As a Bosun user, I want `bosun config validate` to fail immediately with clear error messages when my configuration is invalid, so that I know exactly what to fix.

**Why this priority**: Equal to validation success - the value of validation is in catching problems. Clear error messages are essential for usability.

**Independent Test**: Can be tested by adding an unknown label and verifying the command exits non-zero with an error message naming the unknown key.

**Acceptance Scenarios**:

1. **Given** a container with unknown label `bosun.container.typoKey=value`, **When** I run `bosun config validate`, **Then** command exits non-zero with error "unknown key: bosun.container.typoKey"
2. **Given** a container with invalid duration `bosun.container.stopGracePeriod=notaduration`, **When** I run `bosun config validate`, **Then** command exits non-zero with error about invalid duration format
3. **Given** multiple validation errors across entities, **When** I run `bosun config validate`, **Then** all errors are reported (not just the first one)
4. **Given** a scope mismatch (container label on volume), **When** I run `bosun config validate`, **Then** command exits non-zero with scope error

---

### User Story 3 - Print Effective Configuration (Priority: P2)

As a Bosun user, I want to see the final merged configuration that Bosun will use, so that I can verify my labels and config file are being interpreted correctly.

**Why this priority**: Important for debugging and understanding, but secondary to basic validation. Users need to see what Bosun "sees".

**Independent Test**: Can be tested by running `bosun config validate --print` and verifying JSON output contains expected merged values.

**Acceptance Scenarios**:

1. **Given** valid configuration, **When** I run `bosun config validate --print`, **Then** the merged config is printed as pretty JSON
2. **Given** labels overriding file config, **When** I run `bosun config validate --print`, **Then** output shows label values (higher precedence)
3. **Given** only defaults (no file, no labels), **When** I run `bosun config validate --print`, **Then** output shows all default values

---

### User Story 4 - Select Configuration Source (Priority: P2)

As a Bosun user, I want to choose which configuration sources to validate (labels only, file only, or both), so that I can debug specific parts of my configuration.

**Why this priority**: Useful for troubleshooting but not essential for basic validation workflow.

**Independent Test**: Can be tested by running with `--from labels` and verifying file config is ignored, then `--from file` and verifying labels are ignored.

**Acceptance Scenarios**:

1. **Given** both file and labels exist, **When** I run `bosun config validate --from labels`, **Then** only labels are validated (file ignored)
2. **Given** both file and labels exist, **When** I run `bosun config validate --from file`, **Then** only file is validated (labels ignored)
3. **Given** both file and labels exist, **When** I run `bosun config validate --from auto`, **Then** both sources are merged (default behavior)
4. **Given** `--from file` but no config file exists, **When** I run `bosun config validate`, **Then** command warns that no file found and validates defaults only

---

### User Story 5 - Scope-Specific Validation (Priority: P3)

As a Bosun user, I want to validate configuration for a specific scope (container, volume, network, global), so that I can focus on one entity type at a time.

**Why this priority**: Nice-to-have for advanced users working with complex setups. Most users will validate everything.

**Independent Test**: Can be tested by running `--scope container` and verifying only container-scoped labels are checked.

**Acceptance Scenarios**:

1. **Given** labels on containers and volumes, **When** I run `bosun config validate --scope container`, **Then** only container labels are validated
2. **Given** an invalid volume label, **When** I run `bosun config validate --scope container`, **Then** validation passes (volume error not checked)
3. **Given** global-scoped labels, **When** I run `bosun config validate --scope container`, **Then** global labels are included (global applies everywhere)

---

### Edge Cases

- What happens when Docker is not running? (Should fail with friendly error message)
- What happens when config file path is specified but file is unreadable? (Should fail with clear file access error)
- How does the command handle very large numbers of entities? (Should handle gracefully, perhaps with progress indication)
- What if the user has no `bosun.*` labels at all? (Should succeed with defaults, not fail)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide `bosun config validate` CLI command
- **FR-002**: System MUST exit 0 on successful validation, non-zero on failure
- **FR-003**: System MUST report all validation errors with clear, actionable messages
- **FR-004**: System MUST support `--from` flag with values: `labels`, `file`, `auto` (default: `auto`)
- **FR-005**: System MUST support `--scope` flag with values: `container`, `volume`, `network`, `global` (default: all)
- **FR-006**: System MUST support `--print` flag to output merged config as JSON
- **FR-007**: System MUST use existing label discovery snapshot (read-only, no side effects)
- **FR-008**: System MUST integrate with loader (#58) and merger (#59) for validation
- **FR-009**: System MUST handle Docker unavailability with friendly error message
- **FR-010**: System MUST validate unknown keys, type errors, scope mismatches, and cross-resource collisions

### Key Entities

- **ValidateCommand**: The CLI command struct with flag parsing and execution logic
- **ValidationResult**: Success/failure status with list of errors and merged config
- **ConfigSource**: Enum representing where config came from (labels, file, defaults)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can validate configuration in under 5 seconds for typical setups (< 50 containers)
- **SC-002**: 100% of validation errors include the entity name, key, and reason for failure
- **SC-003**: Command handles Docker unavailability gracefully with helpful error message
- **SC-004**: `--print` output is valid JSON that can be piped to other tools
- **SC-005**: All acceptance scenarios pass as automated tests

## Assumptions

- Loader (#58) and merger (#59) are implemented and provide validation logic
- Label discovery snapshot from existing `bosun labels snapshot` can be reused
- Config file format will be YAML or JSON (to be determined, not blocking this spec)
- Schema version gating is deferred (placeholder only, no-op for now)
