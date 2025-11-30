# Config Documentation Generator

**Location**: `internal/tools/configdoc/`

## Overview

Auto-generates Markdown documentation and JSON Schema from the code-first `ConfigV1` schema. Reads schema metadata via `V1Spec()` and produces deterministic, version-control-friendly output.

## Generated Files

- `docs/config.md` - Human-readable Markdown reference table grouped by scope
- `docs/config.schema.json` - JSON Schema draft 2020-12 for validation/IDE support

## Key Features

1. **Dual Invocation**: `make docs` or `go generate ./internal/config/schema/...`
2. **Deterministic Output**: Sorted keys via `Spec.Keys()`, consistent scope ordering
3. **Scope Grouping**: Global → Container → Volume → Network
4. **Value Format Docs**: Includes documentation for duration, byte-size, and list formats
5. **Type Mapping**: Proper JSON Schema types with format hints

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

## Package Structure

```
internal/tools/configdoc/
├── doc.go              # Package documentation
├── errors.go           # ErrEmptySpec, ErrOutputDir
├── types.go            # Type/scope mapping helpers
├── generator.go        # Generator struct, Options, Generate()
├── markdown.go         # GenerateMarkdown(), templates
├── jsonschema.go       # GenerateJSONSchema()
├── *_test.go           # Unit tests
└── cmd/
    └── main.go         # go:generate entrypoint
```

## Usage

```bash
# Generate docs
make docs

# Verify docs are up-to-date (CI)
make docs && git diff --exit-code docs/
```

## Dependencies

- `internal/config/schema` - Provides `Spec`, `FieldSpec`, `V1Spec()`
- Standard library only: `encoding/json`, `text/template`, `sort`, `os`

## Related Memories

- `config_schema_package` - Source schema definitions
- `architecture` - Hexagonal architecture context
