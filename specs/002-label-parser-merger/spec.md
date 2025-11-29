# Feature Specification: Label Parser and Source Merger

**Feature Branch**: `002-label-parser-merger`
**Created**: 2025-11-29
**Status**: Draft
**Input**: User description: "Label Parser and Source Merger - Parse bosun.* labels into typed config with strict validation (fail on unknown keys) and merge multiple config sources with deterministic precedence (defaults < file < env < labels)"
**Related Issues**: #58, #59

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parse Valid Labels into Typed Config (Priority: P1)

As a Bosun user, I want my `bosun.*` Docker labels to be automatically parsed into strongly-typed configuration values, so that I can configure Bosun behavior through container labels without writing config files.

**Why this priority**: This is the foundational capability - without label parsing, no other config features work. Users expect their Docker labels to "just work" with proper type conversion.

**Independent Test**: Can be fully tested by creating a container with valid `bosun.*` labels and verifying the resulting config struct has correctly typed values (durations as `time.Duration`, sizes as bytes, bools as `bool`, etc.).

**Acceptance Scenarios**:

1. **Given** a container with label `bosun.container.stopGracePeriod=30s`, **When** labels are parsed, **Then** the config contains a `time.Duration` of 30 seconds
2. **Given** a container with label `bosun.container.autoRestart=true`, **When** labels are parsed, **Then** the config contains a `bool` value of `true`
3. **Given** a container with label `bosun.volume.maxSize=1GiB`, **When** labels are parsed, **Then** the config contains an `int64` value of 1073741824 (bytes)
4. **Given** a container with label `bosun.container.logLevel=debug`, **When** labels are parsed for enum type, **Then** the config contains the validated enum value `debug`
5. **Given** a container with label `bosun.global.tags=backup,critical,db`, **When** labels are parsed as list, **Then** the config contains `[]string{"backup", "critical", "db"}`
6. **Given** a container with label `bosun.global.tags=["backup","critical","db"]`, **When** labels are parsed as JSON array, **Then** the config contains the same `[]string` result

---

### User Story 2 - Fail Fast on Unknown Keys (Priority: P1)

As a Bosun user, I want the system to fail immediately when I use an unknown `bosun.*` label key, so that I catch typos and configuration mistakes early rather than having them silently ignored.

**Why this priority**: Critical for user experience and safety. Silent failures lead to frustration and hard-to-debug issues. Failing early with clear error messages is a core design principle.

**Independent Test**: Can be tested by adding an unknown label like `bosun.container.typoedKey=value` and verifying the parser returns a clear error message including the unknown key name.

**Acceptance Scenarios**:

1. **Given** a container with label `bosun.container.unknownKey=value`, **When** labels are parsed, **Then** parsing fails with error message containing "unknown key: bosun.container.unknownKey"
2. **Given** a container with multiple labels including one unknown key, **When** labels are parsed, **Then** all unknown keys are reported in the error message
3. **Given** a volume with label `bosun.volume.invalidSetting=x`, **When** labels are parsed, **Then** parsing fails with a scope-aware error message

---

### User Story 3 - Validate Label Scopes (Priority: P1)

As a Bosun user, I want the system to validate that labels are applied to the correct Docker entity types, so that I don't accidentally put container-specific labels on volumes or networks.

**Why this priority**: Prevents silent misconfiguration. A container label on a volume would be ignored, leading to confusion.

**Independent Test**: Can be tested by applying a container-scoped label to a volume and verifying the parser rejects it with a scope mismatch error.

**Acceptance Scenarios**:

1. **Given** a volume with label `bosun.container.stopGracePeriod=30s`, **When** labels are parsed for scope "volume", **Then** parsing fails with error "key 'bosun.container.stopGracePeriod' not allowed on scope 'volume'"
2. **Given** a container with label `bosun.global.instance=prod`, **When** labels are parsed for scope "container", **Then** parsing succeeds (global scope allowed on any entity)
3. **Given** a network with label `bosun.network.priority=10`, **When** labels are parsed for scope "network", **Then** parsing succeeds

---

### User Story 4 - Validate Type Parsing (Priority: P2)

As a Bosun user, I want clear error messages when I provide values that don't match the expected type, so that I can quickly fix configuration mistakes.

**Why this priority**: Essential for usability but secondary to basic parsing - users need to know why their config is rejected.

**Independent Test**: Can be tested by providing invalid values (e.g., "not-a-number" for an int field) and verifying specific type error messages.

**Acceptance Scenarios**:

1. **Given** a container with label `bosun.container.stopGracePeriod=invalid`, **When** labels are parsed, **Then** parsing fails with error mentioning "invalid duration"
2. **Given** a container with label `bosun.volume.maxSize=notasize`, **When** labels are parsed, **Then** parsing fails with error mentioning "invalid size"
3. **Given** a container with label `bosun.container.autoRestart=maybe`, **When** labels are parsed, **Then** parsing fails with error mentioning "invalid bool"
4. **Given** a container with label `bosun.container.logLevel=invalid`, **When** labels are parsed for enum, **Then** parsing fails with error listing valid enum values

---

### User Story 5 - Merge Config from Multiple Sources (Priority: P2)

As a Bosun user, I want to configure Bosun through multiple sources (defaults, config file, environment, labels) with a clear precedence order, so that I can have sensible defaults while overriding specific settings via labels.

**Why this priority**: Enables flexible configuration patterns. Users need predictable behavior when the same setting is defined in multiple places.

**Independent Test**: Can be tested by providing conflicting values in defaults and labels, verifying labels win, then removing label and verifying file value is used.

**Acceptance Scenarios**:

1. **Given** default `stopGracePeriod=10s` and label `bosun.container.stopGracePeriod=30s`, **When** config is merged, **Then** result is 30s (labels win)
2. **Given** file config `stopGracePeriod=20s` and label `bosun.container.stopGracePeriod=30s`, **When** config is merged, **Then** result is 30s (labels win over file)
3. **Given** default `autoRestart=false` and file `autoRestart=true` with no label, **When** config is merged, **Then** result is `true` (file wins over default)
4. **Given** only defaults with no file or labels, **When** config is merged, **Then** result equals defaults

---

### User Story 6 - Optional Environment Variable Layer (Priority: P3)

As a Bosun operator, I want to optionally enable environment variable configuration between file and labels, so that I can configure Bosun in containerized deployments without modifying labels.

**Why this priority**: Nice-to-have for advanced deployment scenarios. The feature flag approach allows deferring full implementation.

**Independent Test**: Can be tested by enabling the env flag, setting an env var, and verifying it takes precedence over file but not labels.

**Acceptance Scenarios**:

1. **Given** env layer enabled and env `BOSUN_CONTAINER_STOPGRACEPERIOD=25s` with file `stopGracePeriod=20s`, **When** config is merged, **Then** result is 25s
2. **Given** env layer disabled and env `BOSUN_CONTAINER_STOPGRACEPERIOD=25s` with file `stopGracePeriod=20s`, **When** config is merged, **Then** result is 20s (env ignored)
3. **Given** env layer enabled and env `BOSUN_CONTAINER_STOPGRACEPERIOD=25s` with label `bosun.container.stopGracePeriod=30s`, **When** config is merged, **Then** result is 30s (labels still win)

---

### Edge Cases

- What happens when a label value is empty string? (Should use default or fail if required)
- How does the system handle duplicate labels on the same entity? (Docker deduplicates, but what about cross-entity?)
- What happens when a required field has no value in any source? (Should fail with clear error)
- How are malformed JSON arrays handled in list parsing? (Should fall back to CSV or fail with clear error)
- What happens when config file is present but unreadable? (Should fail with clear error, not silently use defaults)

## Requirements *(mandatory)*

### Functional Requirements

#### Label Parsing (#58)

- **FR-001**: System MUST parse string labels into their declared types (string, bool, int, duration, size, enum, list)
- **FR-002**: System MUST fail on unknown `bosun.*` keys with error message including the key name
- **FR-003**: System MUST validate scope (container/volume/network/global) and fail on scope mismatches
- **FR-004**: System MUST accept CSV format (`a,b,c`) for list types
- **FR-005**: System MUST accept JSON array format (`["a","b","c"]`) for list types
- **FR-006**: System MUST use `docker/go-units` for byte size parsing (supports Ki, Mi, Gi, etc.)
- **FR-007**: System MUST use `time.ParseDuration` for duration parsing (supports s, m, h, etc.)
- **FR-008**: System MUST validate enum values against declared allowed values
- **FR-009**: System MUST fail on required fields with no value
- **FR-010**: System MUST report all validation errors (not just first one) for better UX

#### Config Merging (#59)

- **FR-011**: System MUST merge configs with precedence: defaults < file < env (optional) < labels
- **FR-012**: System MUST support optional env layer controlled by feature flag
- **FR-013**: System MUST preserve lower-layer values when higher-layer field is unset
- **FR-014**: System MUST handle nil/absent layers gracefully (skip them in merge)
- **FR-015**: System MUST produce deterministic output for identical inputs

### Key Entities

- **Spec**: Schema metadata describing all valid config keys, their types, scopes, defaults, and constraints (from #57)
- **Cfg (ConfigV1)**: The typed configuration struct containing all settings
- **LabelSet**: Raw `map[string]string` of Docker labels to be parsed
- **MergeSource**: Tagged config from a specific source (defaults, file, env, or labels)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 7 config types (string, bool, int, duration, size, enum, list) parse correctly with valid input
- **SC-002**: 100% of unknown `bosun.*` keys result in clear, actionable error messages
- **SC-003**: Scope validation correctly rejects all cross-scope label applications
- **SC-004**: Config merge produces identical output given identical inputs (deterministic)
- **SC-005**: Unit test coverage for loader and merger packages exceeds 80%
- **SC-006**: All acceptance scenarios pass as automated tests

## Assumptions

- The schema spec from #57 is complete and provides accurate metadata for all config keys
- Docker labels are case-sensitive (matching Go behavior)
- Empty string values for optional fields should use defaults, not fail
- The env layer feature flag defaults to disabled in v1
