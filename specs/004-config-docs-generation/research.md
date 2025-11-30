# Research: Config Documentation Generation

**Feature**: 004-config-docs-generation
**Date**: 2025-11-30

## Research Questions

### 1. JSON Schema Draft 2020-12 Requirements

**Question**: What are the key elements needed to produce valid JSON Schema draft 2020-12?

**Decision**: Use `$schema: "https://json-schema.org/draft/2020-12/schema"` with standard keywords.

**Rationale**: Draft 2020-12 is the latest stable JSON Schema specification, widely supported by editors and validators. Key keywords needed:
- `$schema` - Schema version identifier
- `type` - Data type (`string`, `boolean`, `integer`, `number`, `array`, `object`)
- `properties` - Object property definitions
- `required` - Required property list
- `enum` - Enumerated values
- `default` - Default values
- `description` - Human-readable descriptions
- `deprecated` - Mark deprecated fields

**Alternatives Considered**:
- Draft 7: More widely supported but older, lacks `deprecated` keyword
- Draft 2019-09: Intermediate version, no significant advantage over 2020-12

### 2. Deterministic JSON Output in Go

**Question**: How to ensure JSON output is deterministic (same input → identical bytes)?

**Decision**: Use `json.Marshal` with sorted map keys (Go guarantees this since 1.12) and consistent struct field ordering.

**Rationale**: Go's `encoding/json` package sorts map keys alphabetically by default since Go 1.12. For struct types, fields are marshaled in definition order. Combined with consistent `Spec.Keys()` which returns sorted keys, output will be deterministic.

**Implementation Notes**:
- Use `json.MarshalIndent` for readable output with 2-space indentation
- Build schema using ordered slice of properties, then convert to map for JSON
- Verify with test that runs generator twice and compares bytes

**Alternatives Considered**:
- Custom JSON encoder: Unnecessary complexity, standard library sufficient
- External library (jsoniter): No benefit for this use case

### 3. Markdown Table Generation

**Question**: Best approach for generating Markdown tables in Go?

**Decision**: Use `text/template` with custom template for table generation.

**Rationale**: `text/template` is standard library, provides clean separation of format from logic, and handles escaping. Tables will use GitHub-flavored Markdown (GFM) pipe syntax.

**Template Structure**:
```markdown
| Key | Scope | Type | Default | Required | Description |
|-----|-------|------|---------|----------|-------------|
{{range .Fields}}| `{{.Key}}` | {{.Scope}} | {{.Type}} | {{.Default}} | {{.Required}} | {{.Doc}} |
{{end}}
```

**Alternatives Considered**:
- String concatenation: Error-prone, harder to maintain
- External library (goldmark): Overkill for simple table generation

### 4. Type Mapping: Go → JSON Schema

**Question**: How to map ConfigType values to JSON Schema types?

**Decision**: Direct mapping table:

| ConfigType | JSON Schema Type | Additional Keywords |
|------------|------------------|---------------------|
| TypeString | `"string"` | - |
| TypeBool | `"boolean"` | - |
| TypeInt | `"integer"` | - |
| TypeDuration | `"string"` | `format: "duration"` (custom), pattern for Go syntax |
| TypeSize | `"string"` | `format: "byte-size"` (custom) |
| TypeEnum | `"string"` | `enum: [...]` |
| TypeList | `"array"` | `items: { type: "string" }` |

**Rationale**: JSON Schema has limited built-in types. Duration and size are represented as strings with format hints for tooling. Custom formats are allowed by JSON Schema spec.

**Alternatives Considered**:
- Integer for duration (milliseconds): Loses human readability
- Separate pattern for each duration unit: Over-complicated

### 5. Generator Invocation Method

**Question**: How should the generator be invoked (`go generate` vs standalone)?

**Decision**: Dual approach - `go generate` directive AND `make docs` target.

**Rationale**:
- `go generate` integrates with Go toolchain, standard for code generation
- `make docs` provides explicit, discoverable target for users
- Both invoke the same underlying generator

**Implementation**:
```go
// In internal/config/schema/config_v1.go or cmd/bosun/main.go
//go:generate go run ../tools/configdoc/cmd/generate.go
```

```makefile
# In Makefile
docs:
	go generate ./internal/config/schema/...
```

**Alternatives Considered**:
- Standalone binary: Extra build step, not needed for internal tool
- Only `make docs`: Misses `go generate ./...` convention

### 6. Output Directory Structure

**Question**: Where should generated files be placed?

**Decision**: `docs/` directory at project root.

**Rationale**:
- Standard location for documentation
- Visible in repository root (easy to find)
- Separate from source code
- GitHub renders `docs/` specially in some contexts

**Files**:
- `docs/config.md` - Human-readable Markdown
- `docs/config.schema.json` - Machine-readable JSON Schema

**Alternatives Considered**:
- `api/` directory: Conflicts with existing API directory purpose
- Inline in `internal/`: Would be hidden, not user-facing

### 7. Handling Embedded Structs

**Question**: How to document fields from embedded structs (GlobalConfig, ContainerConfig, etc.)?

**Decision**: Flatten into single table, grouped by scope.

**Rationale**: `ParseTags[ConfigV1]()` already flattens embedded structs into a single `Spec` map. Documentation should mirror this flattened view since that's how users interact with labels. Group by scope for organization.

**Output Structure**:
```markdown
## Global Configuration
| Key | Type | ... |

## Container Configuration
| Key | Type | ... |

## Volume Configuration
| Key | Type | ... |

## Network Configuration
| Key | Type | ... |
```

**Alternatives Considered**:
- Nested sections per struct: Exposes internal implementation detail
- Single flat table: Harder to navigate for users

### 8. Format Documentation for Special Types

**Question**: How to document duration and byte size formats?

**Decision**: Include format explanation section in generated Markdown.

**Rationale**: Users need to know valid syntax. Document Go's duration format and docker/go-units byte format.

**Content**:
```markdown
## Value Formats

### Duration
Go duration syntax: `30s`, `5m`, `1h30m`, `100ms`
Units: `ns`, `us`/`µs`, `ms`, `s`, `m`, `h`

### Byte Size
Docker/go-units syntax: `100MB`, `1GiB`, `500KB`
Units: `B`, `KB`, `MB`, `GB`, `TB` (base-10), `KiB`, `MiB`, `GiB`, `TiB` (base-2)

### List
CSV syntax: `value1,value2,value3`
JSON array: `["value1", "value2", "value3"]`
```

## Dependencies Verification

| Dependency | Status | Notes |
|------------|--------|-------|
| `internal/config/schema.ParseTags` | ✅ Exists | Returns `Spec` with all field metadata |
| `internal/config/schema.V1Spec` | ✅ Exists | Convenience function for ConfigV1 |
| `internal/config/schema.FieldSpec` | ✅ Exists | Has Key, Scope, Type, Default, Enum, Required, Doc, Deprecated |
| `internal/config/schema.Spec.Keys` | ✅ Exists | Returns sorted keys for determinism |
| `internal/config/schema.Spec.Scopes` | ✅ Exists | Groups fields by scope for organization |

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| JSON Schema validation failures | Low | Medium | Use online validator during development; add CI check |
| Non-deterministic output | Low | High | Test by running twice, comparing bytes |
| Future schema changes break generator | Medium | Low | Generator reads metadata dynamically; only format changes need updates |
| Markdown rendering issues | Low | Low | Test with GitHub preview; use standard GFM |

## Open Questions (Resolved)

All questions resolved. No NEEDS CLARIFICATION remaining.
