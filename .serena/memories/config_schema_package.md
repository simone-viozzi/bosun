# Config Schema Package

**Location**: `internal/config/schema/`

## Overview

The `schema` package implements a code-first configuration schema using Go struct tags. It provides the foundation for Bosun's configuration system.

## Key Types

### `Scope` (types.go)
Enum for where a Docker label can be applied:
- `ScopeContainer` - Container labels
- `ScopeVolume` - Volume labels
- `ScopeNetwork` - Network labels
- `ScopeGlobal` - Any entity type

### `ConfigType` (types.go)
Enum for configuration value types:
- `TypeString`, `TypeBool`, `TypeInt`
- `TypeDuration` (time.Duration)
- `TypeSize` (int64 bytes, uses go-units)
- `TypeEnum` (string with allowed values)
- `TypeList` ([]string)

### `FieldSpec` (types.go)
Metadata for a single config field:
```go
type FieldSpec struct {
    Key        string     // Label key (e.g., "bosun.container.stopGracePeriod")
    Scope      Scope      // Where label applies
    Type       ConfigType // Value type
    Default    string     // Default value (unparsed)
    Enum       []string   // Allowed values for enum type
    Required   bool       // Whether field is required
    Doc        string     // Human-readable description
    Deprecated bool       // Whether field is deprecated
    FieldName  string     // Go struct field name
}
```

### `Spec` (types.go)
Map from label key to FieldSpec with helper methods:
- `Keys() []string` - Sorted keys
- `Get(key) (FieldSpec, bool)` - Lookup by key
- `Scopes() map[Scope][]FieldSpec` - Group by scope

## Key Functions

### `ParseTags[T any]() (Spec, error)` (tags.go)
Parses bosun struct tags from type T into a Spec. Handles:
- Embedded structs (recursive)
- All 7 config types
- Duplicate key detection
- Validation of key prefix, scope, type

### `DefaultOf[T any]() (T, error)` (defaults.go)
Creates a new T with default values from tags applied. Uses:
- `time.ParseDuration` for duration
- `units.RAMInBytes` for size
- CSV split for list
- Handles embedded structs

## Tag Format

```go
`bosun:"key=bosun.x.y,scope=container,type=duration,default=30s,doc='Description'"`
```

Required: `key`, `scope`, `type`
Optional: `default`, `enum` (pipe-separated), `required`, `doc` (quoted), `deprecated`

## ConfigV1 (config_v1.go)

The v1 config struct with embedded groups:
- `GlobalConfig` - Instance field
- `ContainerConfig` - StopGracePeriod, HealthCheckInterval, AutoRestart, LogLevel
- `VolumeConfig` - BackupEnabled, MaxSize
- `NetworkConfig` - Priority

Convenience functions: `V1Spec()`, `V1Defaults()`

## Dependencies

- `github.com/docker/go-units` - Size parsing (RAMInBytes)
- Standard library: `reflect`, `time`, `strconv`, `strings`, `sort`

## Testing

All tests in `*_test.go` files alongside source. Key test files:
- `types_test.go` - Scope/ConfigType validation
- `tags_test.go` - Tag parsing, ParseTags tests
- `defaults_test.go` - Value parsers, DefaultOf tests
- `config_v1_test.go` - Integration tests

## Usage by Downstream Packages

- `config/loader` - Uses Spec for label→config mapping
- `config/merge` - Uses Spec for multi-source merging
- `tools/configdoc` - Uses Spec for documentation generation
