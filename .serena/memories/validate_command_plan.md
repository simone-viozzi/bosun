# Validate Command Implementation Plan

**Feature**: 003-cli-config-validate
**Status**: Tasks generated, ready for implementation

## Overview

The `bosun config validate` command validates Docker label configuration against the Bosun schema, merges config from multiple sources, and provides clear error messages.

## Key Design Decisions

1. **CLI Structure**: `bosun config validate` (new command group + subcommand)
2. **Reuse Existing Types**: Uses `loader.ValidationErrors`, `schema.ConfigV1`, `labels.Entity`
3. **Per-Entity Validation**: Iterates snapshot entities, validates each with appropriate scope
4. **Output**: Human-readable errors to stderr, JSON config to stdout (with `--print`)
5. **File Loading**: Deferred (warning only in v1)

## Files to Create

- `internal/cmd/config.go` - Config command group
- `internal/cmd/validate.go` - Validate subcommand
- `integration/validate_test.go` - Integration tests

## Files to Modify

- `internal/cmd/root.go` - Register `NewConfigCmd()`

## New Types (CLI Layer Only)

```go
type ConfigSource string // "auto", "labels", "file"

type ValidateOptions struct {
    Source      ConfigSource
    Scope       string
    PrintConfig bool
    ConfigFile  string
}

type ValidationResult struct {
    Valid        bool
    MergedConfig *schema.ConfigV1
    Errors       []EntityValidationError
    Warnings     []string
}

type EntityValidationError struct {
    Entity labels.Entity
    Errors loader.ValidationErrors
}
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--from` | `auto` | Source: auto, labels, file |
| `--scope` | (all) | Filter: container, volume, network, global |
| `--print` | false | Output merged config as JSON |
| `--stopped` | false | Include stopped containers |

## Exit Codes

- 0: Success
- 1: Validation errors
- 2: Runtime error (Docker unavailable, etc.)

## Artifacts

- `specs/003-cli-config-validate/plan.md`
- `specs/003-cli-config-validate/research.md`
- `specs/003-cli-config-validate/data-model.md`
- `specs/003-cli-config-validate/contracts/cli.md`
- `specs/003-cli-config-validate/quickstart.md`
