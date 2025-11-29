# CLI Contract: Config Validate Command

**Date**: 2025-11-30
**Feature**: 003-cli-config-validate

## Command Structure

```
bosun config validate [flags]
```

### Command Hierarchy

```
bosun
└── config              # NEW: Config command group
    └── validate        # NEW: Validate subcommand
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--from` | `-f` | string | `auto` | Config source: `auto`, `labels`, `file` |
| `--scope` | `-s` | string | (all) | Validate only: `container`, `volume`, `network`, `global` |
| `--print` | `-p` | bool | `false` | Print merged config as JSON |
| `--config` | `-c` | string | (none) | Path to config file (for `--from file` or `--from auto`) |
| `--stopped` | | bool | `false` | Include stopped containers (inherited from snapshot) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Validation passed |
| 1 | Validation failed (invalid config) |
| 2 | Runtime error (Docker unavailable, file not found, etc.) |

## Output Streams

### Success (exit 0)

**Without `--print`:**
```
stdout: Configuration valid (N entities checked)
stderr: (empty)
```

**With `--print`:**
```
stdout: {
  "instance": "",
  "stopGracePeriod": "30s",
  ...
}
stderr: (empty)
```

### Validation Failure (exit 1)

```
stdout: (empty)
stderr: Validation errors:

Container "myapp" (abc123):
  - unknown key: bosun.container.typoKey
  - invalid duration for key 'bosun.container.stopGracePeriod': ...

Found 2 errors in 1 entity
```

### Runtime Error (exit 2)

```
stdout: (empty)
stderr: Error: failed to connect to Docker: ...
Is Docker running?
```

## Flag Combinations

| `--from` | `--scope` | `--print` | Behavior |
|----------|-----------|-----------|----------|
| auto | (all) | false | Validate all entities, merge defaults+file+labels |
| auto | container | false | Validate only containers, merge all |
| labels | (all) | false | Validate labels only (ignore file) |
| labels | (all) | true | Validate labels, print merged config |
| file | (all) | false | ⚠️ Warning: file loader not implemented |
| auto | (all) | true | Validate all, print merged config as JSON |

## Detailed Behavior

### `--from auto` (default)

1. Load label snapshot from Docker
2. Load config file (if `--config` provided or default location exists)
3. For each entity, validate labels with `loader.FromLabels()`
4. Merge: defaults → file → labels
5. Report errors or success

### `--from labels`

1. Load label snapshot from Docker
2. Skip file loading
3. Validate and merge labels only
4. Report errors or success

### `--from file`

1. Skip Docker label loading
2. Load config file (required: `--config` or default location)
3. Validate file config
4. Report errors or success

**Note**: File loading not implemented in v1. Command will warn and validate defaults only.

### `--scope <type>`

Filters entities to validate:
- `container`: Only containers
- `volume`: Only volumes
- `network`: Only networks
- `global`: Only global-scoped labels (applied to any entity)

Global labels are always validated regardless of `--scope`.

### `--print`

Outputs the final merged `ConfigV1` as pretty-printed JSON to stdout.

- On success: Prints config and exits 0
- On failure: Prints errors to stderr, exits 1 (no config printed)

### `--stopped`

Include stopped containers in validation (default: running only).

## Examples

```bash
# Basic validation
bosun config validate

# Validate and show merged config
bosun config validate --print

# Validate only containers
bosun config validate --scope container

# Validate labels only (ignore config file)
bosun config validate --from labels

# Validate with specific config file
bosun config validate --config /path/to/bosun.yaml

# Include stopped containers
bosun config validate --stopped

# Combine flags
bosun config validate --from labels --scope container --print
```

## Error Message Format

### Unknown Key
```
unknown key: bosun.container.typoKey
```

### Type Parse Error
```
invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration "notaduration"
```

### Scope Mismatch
```
key 'bosun.container.stopGracePeriod' not allowed on scope 'volume'
```

### Invalid Enum
```
invalid enum value 'debug' for key 'bosun.container.logLevel': must be one of [trace, info, warn, error]
```

### Required Missing
```
required key 'bosun.global.instance' not provided
```

## Integration Points

| Component | Package | Usage |
|-----------|---------|-------|
| Label Discovery | `adapters/dockerlabels` | `DockerLabelSource.Snapshot()` |
| Validation | `config/loader` | `FromLabels()`, `ValidationErrors` |
| Merging | `config/merge` | `Merge()` |
| Schema | `config/schema` | `V1Spec()`, `V1Defaults()`, `ConfigV1` |
| Domain Types | `domain/labels` | `Entity`, `Snapshot` |
