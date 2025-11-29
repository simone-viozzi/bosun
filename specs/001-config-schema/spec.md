# Feature Specification: Code-First Config Schema with Tags

**Feature Branch**: `001-config-schema`
**Created**: 2025-11-29
**Status**: Draft
**Input**: User description: "feat(config): introduce code-first schema + tags (export-ready) - Define canonical config schema in Go structs with rich bosun: tags for parsing labels into strongly-typed config and later exporting JSON Schema/Markdown"
**Related Issue**: [#57](https://github.com/simone-viozzi/bosun/issues/57)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define Typed Configuration Fields (Priority: P1)

As a Bosun developer, I want to define configuration fields in Go structs with rich metadata tags so that the schema becomes the single source of truth for all config parsing and validation.

**Why this priority**: This is the foundational capability that enables all other features. Without typed struct definitions with tags, we cannot parse labels or export documentation.

**Independent Test**: Can be fully tested by defining a struct with `bosun:` tags and verifying that tag parsing extracts all metadata correctly. Delivers immediate value as the schema definition.

**Acceptance Scenarios**:

1. **Given** a Go struct with `bosun:` tags defining key, scope, type, default, and doc, **When** the tags are parsed, **Then** all metadata is correctly extracted into a machine-readable spec.
2. **Given** a struct field with `bosun:"key=bosun.container.stopGracePeriod,scope=container,type=duration,default=30s,doc='Grace period before force stop'"`, **When** parsed, **Then** the spec contains key="bosun.container.stopGracePeriod", scope="container", type="duration", default="30s", doc="Grace period before force stop".
3. **Given** a struct field with `enum=a|b|c` in its tag, **When** parsed, **Then** the spec contains enum=["a","b","c"].
4. **Given** a struct field with `required=true` in its tag, **When** parsed, **Then** the spec marks that field as required.

---

### User Story 2 - Parse Tags into Machine-Readable Spec (Priority: P1)

As a Bosun developer, I want a generic helper `ParseTags[T any]() (Spec, error)` that reflects over a struct type and produces a spec map keyed by label key, so that downstream components (loader, validator, doc generator) can consume the schema programmatically.

**Why this priority**: Critical infrastructure that connects the struct definitions to all consumers (label parsing, validation, doc export).

**Independent Test**: Can be tested by calling `ParseTags[ConfigV1]()` and asserting the returned Spec contains all expected entries with correct metadata.

**Acceptance Scenarios**:

1. **Given** a config struct type T with multiple tagged fields, **When** `ParseTags[T]()` is called, **Then** it returns a Spec map where each key is the `key=` value from the tag.
2. **Given** a struct with nested embedded structs, **When** `ParseTags[T]()` is called, **Then** fields from embedded structs are also included in the Spec.
3. **Given** a struct field missing the `key=` tag component, **When** `ParseTags[T]()` is called, **Then** it returns an error indicating the missing key.
4. **Given** a struct field with an invalid scope value (not container|volume|network|global), **When** `ParseTags[T]()` is called, **Then** it returns an error.

---

### User Story 3 - Hydrate Defaults from Tags (Priority: P2)

As a Bosun developer, I want a helper `DefaultOf[T any]() T` that returns a new instance of T with all fields set to their default values from the `default=` tag, so that I can easily obtain a baseline configuration.

**Why this priority**: Important for the config merge layer but not required for initial schema definition and parsing.

**Independent Test**: Can be tested by calling `DefaultOf[ConfigV1]()` and verifying each field matches its declared default value.

**Acceptance Scenarios**:

1. **Given** a struct field with `default=30s` and type `time.Duration`, **When** `DefaultOf[T]()` is called, **Then** that field is set to 30 seconds.
2. **Given** a struct field with `default=true` and type `bool`, **When** `DefaultOf[T]()` is called, **Then** that field is set to true.
3. **Given** a struct field with `default=1024` and type `int`, **When** `DefaultOf[T]()` is called, **Then** that field is set to 1024.
4. **Given** a struct field with no `default=` tag, **When** `DefaultOf[T]()` is called, **Then** that field retains its Go zero value.
5. **Given** a struct field with `default=a,b,c` and type `[]string`, **When** `DefaultOf[T]()` is called, **Then** that field is set to `["a","b","c"]`.

---

### User Story 4 - Support All Required Types (Priority: P1)

As a Bosun developer, I want the schema to support string, bool, int, time.Duration, byte size, enum, and []string types so that I can model real-world configuration needs.

**Why this priority**: Type diversity is essential for real configuration scenarios and must be designed upfront.

**Independent Test**: Can be tested by defining a struct with one field of each type and verifying `ParseTags` correctly identifies and validates each type.

**Acceptance Scenarios**:

1. **Given** a field with `type=string`, **When** parsed, **Then** the spec records type as "string".
2. **Given** a field with `type=bool`, **When** parsed, **Then** the spec records type as "bool".
3. **Given** a field with `type=int`, **When** parsed, **Then** the spec records type as "int".
4. **Given** a field with `type=duration`, **When** parsed, **Then** the spec records type as "duration".
5. **Given** a field with `type=size`, **When** parsed, **Then** the spec records type as "size" (byte size).
6. **Given** a field with `type=enum,enum=debug|info|warn|error`, **When** parsed, **Then** the spec records type as "enum" with allowed values.
7. **Given** a field with `type=list`, **When** parsed, **Then** the spec records type as "list" (string list).

---

### User Story 5 - Reserve Deprecated Tag Option (Priority: P3)

As a Bosun developer, I want to reserve a `deprecated=true` tag option so that future schema migrations can mark fields as deprecated without breaking changes.

**Why this priority**: Future-proofing feature that requires minimal implementation now.

**Independent Test**: Can be tested by adding `deprecated=true` to a field tag and verifying the spec includes the deprecated flag.

**Acceptance Scenarios**:

1. **Given** a struct field with `deprecated=true` in its tag, **When** parsed, **Then** the spec marks that field as deprecated.
2. **Given** a struct field without `deprecated` in its tag, **When** parsed, **Then** the spec defaults deprecated to false.

---

### Edge Cases

- What happens when a tag has malformed syntax (e.g., missing `=` sign)?
- How does the system handle duplicate keys across different struct fields?
- What happens when enum values contain the delimiter character `|`?
- How does parsing handle empty default values vs. no default specified?
- What happens when type in tag doesn't match the Go field type (e.g., `type=int` on a string field)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a new package at `internal/config/schema`.
- **FR-002**: System MUST support custom struct tag format: `bosun:"key=...,scope=...,type=...,default=...,enum=...,required=...,doc='...'"`.
- **FR-003**: System MUST support scope values: `container`, `volume`, `network`, `global`.
- **FR-004**: System MUST support type values: `string`, `bool`, `int`, `duration`, `size`, `enum`, `list`.
- **FR-005**: System MUST provide `ParseTags[T any]() (Spec, error)` generic function that extracts tag metadata via reflection.
- **FR-006**: System MUST provide `DefaultOf[T any]() T` generic function that hydrates default values from tags.
- **FR-007**: System MUST define a v1 config struct with representative fields covering all supported types.
- **FR-008**: System MUST use dotted key paths under `bosun.*` namespace (e.g., `bosun.container.stopGracePeriod`).
- **FR-009**: System MUST reserve `deprecated=true|false` tag option for future migration support (parsed but no-op).
- **FR-010**: System MUST return descriptive errors for invalid tag syntax or values.
- **FR-011**: System MUST validate that scope values are one of the allowed constants.
- **FR-012**: System MUST validate that type values are one of the supported types.

### Key Entities

- **Spec**: A map from label key (string) to FieldSpec, representing the complete schema definition.
- **FieldSpec**: Metadata for a single configuration field including key, scope, type, default value, enum values, required flag, documentation string, and deprecated flag.
- **Scope**: Enumeration of valid scopes (container, volume, network, global) indicating where a label can be applied.
- **ConfigType**: Enumeration of supported types (string, bool, int, duration, size, enum, list).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All unit tests pass for tag parsing covering each supported type.
- **SC-002**: `ParseTags[ConfigV1]()` returns a complete Spec with all declared fields.
- **SC-003**: `DefaultOf[ConfigV1]()` returns a struct with all default values correctly hydrated.
- **SC-004**: Invalid tag syntax produces clear, actionable error messages.
- **SC-005**: The v1 config struct demonstrates at least one field of each supported type (string, bool, int, duration, size, enum, list).
- **SC-006**: Schema is export-ready for downstream JSON Schema and Markdown generation (issue #61).

## Assumptions

- Keys follow the `bosun.*` dotted namespace convention established in the project.
- Duration values use Go's `time.ParseDuration` format (e.g., "30s", "5m", "1h").
- Byte size values use docker/go-units format (e.g., "1GB", "512MB").
- List values in defaults use CSV format (e.g., "a,b,c").
- Enum delimiter is `|` character.
- Single quotes are used for doc strings to allow commas within documentation.
