# Config Documentation Generator

**Location**: `internal/tools/configdoc/` (to be created)
**Issue**: #61
**Branch**: `004-config-docs-generation`

## Overview

Auto-generates Markdown documentation and JSON Schema from the code-first `ConfigV1` schema. Reads schema metadata via `ParseTags[ConfigV1]()` and produces deterministic, version-control-friendly output.

## Generated Files

- `docs/config.md` - Human-readable Markdown reference table
- `docs/config.schema.json` - JSON Schema draft 2020-12 for validation/IDE support

## Key Design Decisions

1. **Dual Invocation**: Both `make docs` and `go generate ./internal/config/schema/...`
2. **Deterministic Output**: Uses `Spec.Keys()` for sorted keys, `json.MarshalIndent` with sorted map keys
3. **Scope Grouping**: Markdown groups fields by scope (Global, Container, Volume, Network)
4. **Type Mapping**: Duration/Size as strings with format hints; List as array of strings

## Type Mapping (ConfigType → JSON Schema)

| ConfigType | JSON Schema | Additional |
|------------|-------------|------------|
| TypeString | "string" | - |
| TypeBool | "boolean" | - |
| TypeInt | "integer" | - |
| TypeDuration | "string" | format: "duration" |
| TypeSize | "string" | format: "byte-size" |
| TypeEnum | "string" | enum: [...] |
| TypeList | "array" | items: {type: "string"} |

## Dependencies

- `internal/config/schema` - Provides `Spec`, `FieldSpec`, `V1Spec()`, `ParseTags[T]()`
- Standard library only: `encoding/json`, `text/template`, `sort`, `os`

## Package Structure

```
internal/tools/configdoc/
├── doc.go           # Package documentation
├── generator.go     # Generator struct and Generate()
├── markdown.go      # GenerateMarkdown()
├── jsonschema.go    # GenerateJSONSchema()
├── generator_test.go
└── cmd/
    └── main.go      # Entrypoint for go:generate
```

## Usage

```bash
# Generate docs
make docs

# Verify docs are up-to-date (CI)
make docs && git diff --exit-code docs/
```

## Task Breakdown

Total: **52 tasks** across 8 phases

| Phase | Focus | Tasks |
|-------|-------|-------|
| 1 | Setup | 4 |
| 2 | Foundational | 3 |
| 3 | US1 - Markdown (P1) | 10 |
| 4 | US2 - JSON Schema (P1) | 9 |
| 5 | US3 - Determinism (P2) | 6 |
| 6 | US4 - Format Docs (P2) | 7 |
| 7 | US5 - Build Integration (P2) | 6 |
| 8 | Polish | 7 |

**MVP**: Phases 1-4 (Setup + Foundational + US1 + US2) = 26 tasks

## Related Memories

- `config_schema_package` - Source schema definitions
- `architecture` - Hexagonal architecture context
