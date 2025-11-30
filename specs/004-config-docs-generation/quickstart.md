# Quickstart: Config Documentation Generation

**Feature**: 004-config-docs-generation
**Date**: 2025-11-30

## Overview

This feature adds automatic documentation generation from Bosun's code-first configuration schema. Running `make docs` produces:

- `docs/config.md` - Human-readable Markdown reference
- `docs/config.schema.json` - Machine-readable JSON Schema for validation and IDE support

## Quick Commands

```bash
# Generate documentation
make docs

# Or using go generate
go generate ./internal/config/schema/...

# Verify docs are up-to-date (for CI)
make docs && git diff --exit-code docs/
```

## Generated Output

### Markdown Documentation (`docs/config.md`)

A reference table showing all `bosun.*` configuration labels:

```markdown
# Bosun Configuration Reference

## Container Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.container.autoRestart` | boolean | true | No | - | Automatically restart containers |
| `bosun.container.logLevel` | enum | info | No | debug \| info \| warn \| error | Logging verbosity level |
| `bosun.container.stopGracePeriod` | duration | 30s | No | - | Grace period before force stopping |
...
```

### JSON Schema (`docs/config.schema.json`)

A JSON Schema for validation and editor autocomplete:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Bosun Configuration Reference",
  "type": "object",
  "properties": {
    "bosun.container.autoRestart": {
      "type": "boolean",
      "description": "Automatically restart containers",
      "default": true
    }
  }
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ConfigV1 (Go struct)                     │
│  internal/config/schema/config_v1.go                        │
└─────────────────────────┬───────────────────────────────────┘
                          │ ParseTags[ConfigV1]()
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Spec (metadata)                         │
│  map[string]FieldSpec with Key, Type, Default, Doc, etc.    │
└─────────────────────────┬───────────────────────────────────┘
                          │ configdoc.Generator
                          ▼
        ┌─────────────────┴─────────────────┐
        ▼                                   ▼
┌───────────────────┐             ┌───────────────────┐
│   docs/config.md  │             │ docs/config.      │
│   (Markdown)      │             │ schema.json       │
└───────────────────┘             └───────────────────┘
```

## Adding New Configuration Fields

1. Add field to appropriate config struct in `internal/config/schema/config_v1.go`:

```go
type ContainerConfig struct {
    // NewField does something useful.
    NewField string `bosun:"key=bosun.container.newField,scope=container,type=string,default=value,doc='Does something useful'"`
}
```

2. Regenerate documentation:

```bash
make docs
```

3. Commit both the code change and updated docs.

## CI Integration

Add to `.github/workflows/ci.yml`:

```yaml
- name: Verify documentation is up-to-date
  run: |
    make docs
    git diff --exit-code docs/
```

This ensures documentation stays in sync with code changes.

## File Locations

| File | Purpose |
|------|---------|
| `internal/tools/configdoc/` | Generator package |
| `internal/tools/configdoc/cmd/main.go` | Generator entrypoint |
| `internal/config/schema/config_v1.go` | `//go:generate` directive location |
| `docs/config.md` | Generated Markdown |
| `docs/config.schema.json` | Generated JSON Schema |
| `Makefile` | `docs` target |

## Troubleshooting

### "docs/ directory doesn't exist"

The generator creates it automatically. If you see permission errors, check write permissions.

### "git diff shows changes after make docs"

Documentation is out of sync. Regenerate and commit:

```bash
make docs
git add docs/
git commit -m "chore: regenerate config documentation"
```

### JSON Schema validation fails

Ensure you're using a validator that supports draft 2020-12. Test with:

```bash
# Using ajv-cli
npx ajv validate -s docs/config.schema.json --spec=draft2020 -d your-config.json
```
