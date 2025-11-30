# Feature Specification: Config Integration Tests

**Feature Branch**: `005-config-integration-tests`
**Created**: 2025-11-29
**Status**: Draft
**Input**: User description: "Config Integration Tests - End-to-end integration tests proving typed settings parsing from Docker labels with happy path and failure scenarios"
**Related Issues**: #62

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Verify Typed Settings Happy Path (Priority: P1)

As a Bosun developer, I want integration tests that prove all config types in ConfigV1 (string, bool, int, duration, size, enum) are correctly parsed from real Docker labels, so that I have confidence the system works end-to-end.

**Why this priority**: The primary purpose of this feature - prove the label-to-config pipeline works in real Docker environments, not just unit tests.

**Independent Test**: Can be tested by running the integration test suite against a Docker Compose stack with valid labels and verifying all assertions pass.

**Acceptance Scenarios**:

1. **Given** a container with `bosun.container.autoRestart=true`, **When** labels are collected and parsed, **Then** config field is `bool` value `true`
2. **Given** a container with `bosun.container.stopGracePeriod=30s`, **When** labels are collected and parsed, **Then** config field is `time.Duration` of 30 seconds
3. **Given** a container with `bosun.volume.maxSize=1GiB`, **When** labels are collected and parsed, **Then** config field is `int64` value `1073741824`
4. **Given** a container with `bosun.container.logLevel=debug`, **When** labels are collected and parsed, **Then** config field is enum value `debug`
5. **Given** a container with `bosun.instance=prod`, **When** labels are collected and parsed, **Then** config field is `string` value `"prod"`

---

### User Story 2 - Verify Unknown Key Rejection (Priority: P1)

As a Bosun developer, I want integration tests that prove unknown `bosun.*` keys cause hard failures, so that I can guarantee the strict validation policy works in production.

**Why this priority**: This is the core safety feature - failing on unknown keys prevents silent misconfigurations. Must be proven end-to-end.

**Independent Test**: Can be tested by running a separate test against a Docker service with an invalid label and verifying it fails with the expected error.

**Acceptance Scenarios**:

1. **Given** a service with label `bosun.container.unknownKey=value`, **When** labels are collected and parsed, **Then** parsing fails with error containing "unknown key"
2. **Given** a service with typo `bosun.container.autoRestrat=true`, **When** labels are collected and parsed, **Then** parsing fails with error mentioning the typo'd key
3. **Given** multiple unknown keys on different services, **When** labels are collected and parsed, **Then** all unknown keys are reported in the error

---

### User Story 3 - Verify Scope Validation (Priority: P2)

As a Bosun developer, I want integration tests that prove scope mismatches are rejected, so that container-only labels on volumes cause failures.

**Why this priority**: Important safety check but less critical than unknown key rejection. Scope errors are less common but still need coverage.

**Independent Test**: Can be tested with a Docker Compose stack where a volume has a container-scoped label applied.

**Acceptance Scenarios**:

1. **Given** a volume with label `bosun.container.stopGracePeriod=30s`, **When** labels are collected and parsed for scope "volume", **Then** parsing fails with scope mismatch error
2. **Given** a container with label `bosun.global.instance=prod`, **When** labels are collected and parsed for scope "container", **Then** parsing succeeds (global allowed anywhere)

---

### User Story 4 - Verify Type Validation Errors (Priority: P2)

As a Bosun developer, I want integration tests that prove invalid type values are rejected with clear error messages, so that users get helpful feedback on malformed labels.

**Why this priority**: User experience for error cases. Secondary to happy path but essential for good DX.

**Independent Test**: Can be tested with services that have intentionally malformed values (bad duration, invalid bool, etc.).

**Acceptance Scenarios**:

1. **Given** a container with `bosun.container.stopGracePeriod=notaduration`, **When** labels are parsed, **Then** parsing fails with duration format error
2. **Given** a container with `bosun.container.autoRestart=maybe`, **When** labels are parsed, **Then** parsing fails with bool format error
3. **Given** a container with `bosun.container.logLevel=invalid`, **When** labels are parsed, **Then** parsing fails with enum error listing valid values

---

### User Story 5 - Verify Config Merge End-to-End (Priority: P2)

As a Bosun developer, I want integration tests that prove the full pipeline (discover → parse → merge) works correctly, so that I can verify labels override defaults as expected.

**Why this priority**: Proves the complete flow works, building on individual component tests.

**Independent Test**: Can be tested by verifying merged config contains label values where provided and defaults where not.

**Acceptance Scenarios**:

1. **Given** defaults `stopGracePeriod=10s` and label `bosun.container.stopGracePeriod=30s`, **When** full pipeline runs, **Then** merged config has 30s
2. **Given** label for one field and no label for another, **When** full pipeline runs, **Then** merged config has label value for first and default for second
3. *(Deferred)* **Given** multiple containers with different label values, **When** full pipeline runs, **Then** each container's config reflects its own labels

---

### Edge Cases

- What happens when Docker Compose stack fails to start? (Test should fail with clear setup error)
- How do tests clean up containers/volumes after running? (Should use t.Cleanup() for automatic teardown)
- What if integration test runs without Docker available? (Tests fail with clear error; CI must have Docker)
- How to isolate test stacks from each other? (Use unique project names per test)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Tests MUST use real Docker containers via Docker Compose
- **FR-002**: Tests MUST verify all 6 config types in ConfigV1 parse correctly (string, bool, int, duration, size, enum)
- **FR-003**: Tests MUST verify unknown keys cause hard failure with descriptive error
- **FR-004**: Tests MUST verify scope mismatches cause hard failure
- **FR-005**: Tests MUST verify type parse errors include helpful messages
- **FR-006**: Tests MUST verify the full discover → parse → merge pipeline
- **FR-007**: Tests MUST clean up Docker resources after each test
- **FR-008**: Tests MUST use unique project names to avoid conflicts
- **FR-009**: Tests MUST be runnable via `make it` or `go test -tags=integration`
- **FR-010**: Tests MUST fail with clear error message when Docker is unavailable (CI environments must have Docker)

### Key Entities

- **Test Compose File**: YAML file defining test services with various labels
- **Test Harness**: Utility for starting/stopping Docker Compose stacks
- **Assertion Helpers**: Functions to verify typed config values

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All happy path tests pass with valid labeled containers
- **SC-002**: Unknown key test fails with expected error message
- **SC-003**: Scope validation test fails with expected error message
- **SC-004**: Type validation tests fail with expected error messages
- **SC-005**: Full pipeline test produces correct merged config
- **SC-006**: Tests complete in under 60 seconds on typical hardware (aspirational; 3-min timeout as safety margin)
- **SC-007**: Tests clean up all Docker resources (no leaked containers/volumes)

## Clarifications

### Session 2025-11-30

- Q: How should the list type requirement be handled given ConfigV1 doesn't include a list field? → A: Remove list type from FR-002 scope (6 types) since ConfigV1 doesn't have one; unit tests cover list parsing
- Q: What behavior should trigger when Docker is unavailable? → A: Let tests fail with clear error message; CI should always have Docker
- Q: Should the invalid bool test case be added to complete US-4 coverage? → A: Yes, add `bosun.container.autoRestart=maybe` to `validate-invalid.yaml` to test bool parse error
- Q: Should multi-container testing be added, or is it out of scope for this feature? → A: Mark as future work; add follow-up issue for multi-container testing
- Q: Is the 60-second target realistic, and should tests actively verify/enforce this? → A: Keep 60s as aspirational goal; 3-min timeout is safety margin; observe in CI

## Assumptions

- Existing test harness in `internal/testutil/` can be reused
- Loader (#58) and merger (#59) are implemented before these tests
- Integration tests run in CI environment with Docker available
- Docker Compose v2 syntax is supported
