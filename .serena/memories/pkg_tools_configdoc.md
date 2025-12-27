# Config Documentation Generator

## Scope
Auto-generated documentation in `internal/tools/configdoc/`.

## What

### Generated Files
- `docs/config.md` - Markdown reference grouped by scope
- `docs/config.schema.json` - JSON Schema draft 2020-12

### Invocation
- `make docs`
- `go generate ./internal/config/schema/...`

### Type Mapping (ConfigType → JSON Schema)
| ConfigType | JSON Schema | Additional |
|------------|-------------|------------|
| TypeString | "string" | - |
| TypeBool | "boolean" | - |
| TypeInt | "integer" | - |
| TypeDuration | "string" | format: "duration" |
| TypeSize | "string" | format: "byte-size" |
| TypeEnum | "string" | enum: [...] |
| TypeList | "array" | items: {type: "string"} |

### Features
- Deterministic output (sorted keys)
- Scope grouping: Global → Container → Volume → Network
- Value format documentation
- Deprecation markers

### Package Structure
```
internal/tools/configdoc/
├── generator.go     # Main Generator, Options
├── markdown.go      # GenerateMarkdown()
├── jsonschema.go    # GenerateJSONSchema()
└── cmd/main.go      # go:generate entrypoint
```

## Why
Single source of truth: code defines schema, docs are generated. No drift.

## Related
- `pkg_config_schema` - Source schema
- `arch_development_lifecycle` - `make docs` command
