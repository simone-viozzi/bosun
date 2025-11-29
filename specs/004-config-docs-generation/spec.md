# Feature Specification: Config Documentation Generation

**Feature Branch**: `004-config-docs-generation`
**Created**: 2025-11-29
**Status**: Draft
**Input**: User description: "Config Documentation Generation - Auto-generate Markdown documentation and JSON Schema from the code-first config schema for developer-friendly docs and machine-readable validation"
**Related Issues**: #61

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Markdown Documentation (Priority: P1)

As a Bosun user, I want to read documentation that shows all available `bosun.*` configuration keys, their types, defaults, and descriptions, so that I know exactly what I can configure and how.

**Why this priority**: Documentation is essential for adoption. Users cannot use what they don't know exists. This is the primary deliverable.

**Independent Test**: Can be tested by running the generator and verifying the output contains a table with all config keys from the schema, properly formatted.

**Acceptance Scenarios**:

1. **Given** the ConfigV1 schema, **When** I run `make docs`, **Then** `docs/config.md` is generated with a table of all config keys
2. **Given** a config field with `doc='Description'` tag, **When** docs are generated, **Then** the description appears in the table
3. **Given** a field with enum values `debug|info|warn|error`, **When** docs are generated, **Then** allowed values are listed
4. **Given** a field with default value `30s`, **When** docs are generated, **Then** the default is shown in the table
5. **Given** multiple scopes (container, volume, network, global), **When** docs are generated, **Then** scope is clearly indicated for each key

---

### User Story 2 - Generate JSON Schema (Priority: P1)

As a developer or tool author, I want a JSON Schema that describes the Bosun configuration format, so that I can validate config files and get editor autocomplete.

**Why this priority**: JSON Schema enables tooling ecosystem - IDE autocomplete, CI validation, external integrations. Equal importance to human docs.

**Independent Test**: Can be tested by running the generator and validating the output against JSON Schema draft 2020-12 spec.

**Acceptance Scenarios**:

1. **Given** the ConfigV1 schema, **When** I run `make docs`, **Then** `docs/config.schema.json` is generated
2. **Given** the generated schema, **When** validated against JSON Schema draft 2020-12, **Then** it passes validation
3. **Given** a string field, **When** schema is generated, **Then** field type is "string"
4. **Given** an enum field with values `a|b|c`, **When** schema is generated, **Then** field has "enum": ["a", "b", "c"]
5. **Given** a required field, **When** schema is generated, **Then** field appears in "required" array

---

### User Story 3 - Deterministic Output (Priority: P2)

As a developer, I want the generated documentation to be deterministic (same input = same output), so that I can track changes in version control without spurious diffs.

**Why this priority**: Essential for maintainability. Non-deterministic output causes confusion in PRs and makes it hard to verify changes.

**Independent Test**: Can be tested by running the generator twice and comparing outputs - they must be identical.

**Acceptance Scenarios**:

1. **Given** identical schema input, **When** generator is run twice, **Then** outputs are byte-for-byte identical
2. **Given** multiple config fields, **When** docs are generated, **Then** fields are in stable sorted order (alphabetical by key)
3. **Given** JSON Schema output, **When** object properties are compared, **Then** they are in deterministic order

---

### User Story 4 - Document Special Formats (Priority: P2)

As a Bosun user, I want the documentation to explain special value formats like durations and byte sizes, so that I know how to write valid configuration values.

**Why this priority**: Users need to know the syntax for complex types. Without this, they'll make mistakes.

**Independent Test**: Can be tested by verifying the generated docs include format examples for duration and size types.

**Acceptance Scenarios**:

1. **Given** a duration type field, **When** docs are generated, **Then** format explanation shows examples like `30s`, `5m`, `1h`
2. **Given** a byte size type field, **When** docs are generated, **Then** format explanation shows examples like `100MB`, `1GiB`
3. **Given** a list type field, **When** docs are generated, **Then** format explanation shows both CSV and JSON array syntax

---

### User Story 5 - Integration with Build System (Priority: P2)

As a developer, I want to regenerate documentation via `make docs` or `go generate`, so that docs stay in sync with code changes through CI.

**Why this priority**: Automation ensures docs don't drift from code. Manual doc updates are error-prone.

**Independent Test**: Can be tested by running `make docs` and verifying the command succeeds and produces expected files.

**Acceptance Scenarios**:

1. **Given** a Makefile target `docs`, **When** I run `make docs`, **Then** documentation files are generated
2. **Given** `go generate` directives in code, **When** I run `go generate ./...`, **Then** documentation files are generated
3. **Given** generated docs committed to repo, **When** CI runs `make docs && git diff`, **Then** no changes if docs are up-to-date

---

### Edge Cases

- What happens if schema has no fields? (Should generate empty table with headers, not fail)
- How are deprecated fields handled? (Should be marked as deprecated in both Markdown and JSON Schema)
- What if output directory doesn't exist? (Should create it automatically)
- How are nested/embedded structs documented? (Should flatten into key namespace like `bosun.container.x`)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST generate `docs/config.md` with a table of all config keys
- **FR-002**: System MUST generate `docs/config.schema.json` conforming to JSON Schema draft 2020-12
- **FR-003**: Generated docs MUST include: key, scope, type, default, required status, enum values, description
- **FR-004**: Generated output MUST be deterministic (stable ordering)
- **FR-005**: System MUST document list encoding formats (CSV and JSON array)
- **FR-006**: System MUST document duration format (Go duration syntax)
- **FR-007**: System MUST document byte size format (docker/go-units syntax)
- **FR-008**: System MUST be invocable via `make docs` or `go generate`
- **FR-009**: System MUST mark deprecated fields in both outputs
- **FR-010**: Generator MUST read schema from existing ParseTags[ConfigV1]() function

### Key Entities

- **DocGenerator**: The tool that reads schema and produces documentation files
- **MarkdownFormatter**: Formats schema spec as Markdown table
- **JSONSchemaFormatter**: Formats schema spec as JSON Schema document

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Generated Markdown is valid and renders correctly in GitHub
- **SC-002**: Generated JSON Schema validates against JSON Schema draft 2020-12
- **SC-003**: Running generator twice produces identical output
- **SC-004**: All config fields from schema appear in generated documentation
- **SC-005**: CI can verify documentation is up-to-date (no uncommitted changes after `make docs`)

## Assumptions

- Schema package from #57 provides `ParseTags[T]()` function that returns complete metadata
- Output files go in `docs/` directory at project root
- JSON Schema draft 2020-12 is the target version
- Generator lives in `internal/tools/configdoc/` or similar
