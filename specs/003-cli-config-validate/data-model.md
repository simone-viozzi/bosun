# Data Model: CLI Config Validate Command

**Date**: 2025-11-30
**Feature**: 003-cli-config-validate

## Overview

This feature primarily uses existing types from the `loader`, `merge`, and `schema` packages. Only minimal new types are needed for the CLI layer.

## Existing Types (Reused)

### From `internal/config/schema/`

```go
// ConfigV1 - The merged configuration structure
type ConfigV1 struct {
    GlobalConfig
    ContainerConfig
    VolumeConfig
    NetworkConfig
}

// Scope - Where a label can be applied
type Scope string // "container", "volume", "network", "global"
```

### From `internal/config/loader/`

```go
// ValidationError - Single validation failure
type ValidationError struct {
    Key     string       // Label key (e.g., "bosun.container.unknownKey")
    Value   string       // Raw value provided
    Scope   schema.Scope // Entity scope
    Message string       // Human-readable error
    Err     error        // Underlying error (optional)
}

// ValidationErrors - Collection of errors (implements error interface)
type ValidationErrors []ValidationError
```

### From `internal/domain/labels/`

```go
// Entity - A Docker entity with labels
type Entity struct {
    Kind   EntityKind        // container, volume, network
    ID     string            // Docker ID
    Name   string            // Human-readable name
    Labels map[string]string // bosun.* labels
    Meta   map[string]string // Additional metadata
}

// Snapshot - Collection of entities at a point in time
type Snapshot struct {
    Entities []Entity
    TakenAt  time.Time
}
```

## New Types (CLI Layer)

### ConfigSource Enum

```go
// ConfigSource indicates where config values come from
type ConfigSource string

const (
    SourceAuto   ConfigSource = "auto"   // Merge all sources (default)
    SourceLabels ConfigSource = "labels" // Docker labels only
    SourceFile   ConfigSource = "file"   // Config file only (future)
)
```

### ValidateOptions

```go
// ValidateOptions holds CLI flag values
type ValidateOptions struct {
    Source      ConfigSource // --from flag
    Scope       string       // --scope flag (empty = all)
    PrintConfig bool         // --print flag
    ConfigFile  string       // --config flag (future)
}
```

### ValidationResult

```go
// ValidationResult holds the outcome of validation
type ValidationResult struct {
    Valid        bool                    // Overall success
    MergedConfig *schema.ConfigV1        // Merged config (if valid or for --print)
    Errors       []EntityValidationError // Per-entity errors
    Warnings     []string                // Non-fatal warnings
}

// EntityValidationError wraps validation errors with entity context
type EntityValidationError struct {
    Entity labels.Entity            // The entity that failed
    Errors loader.ValidationErrors  // Validation errors for this entity
}
```

## Type Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                            │
├─────────────────────────────────────────────────────────────┤
│  ValidateOptions ──────────► ValidationResult               │
│       │                           │                         │
│       │                           ├── MergedConfig          │
│       │                           │      (schema.ConfigV1)  │
│       │                           │                         │
│       │                           └── Errors[]              │
│       │                                  │                  │
│       ▼                                  ▼                  │
│  ConfigSource                   EntityValidationError       │
│  (auto/labels/file)                    │                    │
│                                        ├── Entity           │
│                                        │   (labels.Entity)  │
│                                        │                    │
│                                        └── Errors           │
│                                            (loader.         │
│                                             ValidationErrors)│
└─────────────────────────────────────────────────────────────┘
```

## State Transitions

The validate command has no persistent state. It performs a single read-only operation:

```
[Start]
    │
    ▼
[Load Snapshot] ─── Docker unavailable ───► [Error: Docker not running]
    │
    ▼
[Filter by Scope] (if --scope specified)
    │
    ▼
[Validate Each Entity] ─── Per-entity errors ───► [Collect in result.Errors]
    │
    ▼
[Merge Configs] (defaults + file + labels)
    │
    ▼
[Check result.Errors]
    │
    ├── Has errors ───► [Print errors to stderr, exit 1]
    │
    └── No errors
           │
           ├── --print ───► [Print JSON config to stdout, exit 0]
           │
           └── No --print ───► [Print success message, exit 0]
```

## Validation Rules

Validation rules are defined in `schema.V1Spec()` via struct tags. The validate command enforces:

| Rule | Package | Description |
|------|---------|-------------|
| Unknown key | loader | Key not in spec = error |
| Scope mismatch | loader | Label scope doesn't match entity type |
| Type parse error | loader | Value doesn't parse to declared type |
| Invalid enum | loader | Value not in allowed enum values |
| Required missing | loader | Required field not provided |

## JSON Output Schema (--print)

When `--print` is used, output follows `schema.ConfigV1` structure:

```json
{
  "instance": "",
  "stopGracePeriod": "30s",
  "healthCheckInterval": "10s",
  "autoRestart": false,
  "logLevel": "info",
  "backupEnabled": false,
  "maxSize": "0",
  "priority": 0
}
```

## Error Output Format

Validation errors are printed to stderr in human-readable format:

```
Validation errors:

Container "myapp" (abc123):
  - unknown key: bosun.container.typoKey
  - invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration "notaduration"

Volume "data" (data):
  - key 'bosun.container.stopGracePeriod' not allowed on scope 'volume'

Found 3 errors in 2 entities
```
