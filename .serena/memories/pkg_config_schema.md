# Config Schema Package

## Scope
Code-first configuration schema in `internal/config/schema/`.

## What

### Core Types

**`Scope`** - Where a label can be applied
- `ScopeContainer`, `ScopeVolume`, `ScopeNetwork`, `ScopeGlobal`

**`ConfigType`** - Value types
- `TypeString`, `TypeBool`, `TypeInt`
- `TypeDuration` (time.Duration)
- `TypeSize` (int64 bytes, via go-units)
- `TypeEnum` (string with allowed values)
- `TypeList` ([]string)

**`FieldSpec`** - Single field metadata
- `Key`, `Scope`, `Type`, `Default`, `Enum`, `Required`, `Doc`, `Deprecated`, `FieldName`

**`Spec`** - Map of key→FieldSpec with helpers
- `Keys() []string` - Sorted keys
- `Get(key) (FieldSpec, bool)` - Lookup
- `Scopes() map[Scope][]FieldSpec` - Group by scope

### Tag Format
```go
`bosun:"key=bosun.x.y,scope=container,type=duration,default=30s,doc='Description'"`
```

### Functions
- `ParseTags[T]() (Spec, error)` - Parse struct tags into Spec
- `DefaultOf[T]() (T, error)` - Create T with defaults applied
- `V1Spec()` / `V1Defaults()` - Convenience for ConfigV1

### ConfigV1 Structure
- `GlobalConfig` - Instance field
- `ContainerConfig` - StopGracePeriod, HealthCheckInterval, AutoRestart, LogLevel
- `VolumeConfig` - BackupEnabled, MaxSize
- `NetworkConfig` - Priority

### Job Labels (separate)
- `JobLabelConfig` - Container job labels
- `JobVolumeConfig` - Volume job labels
- `JobSpec()` - Combined job label spec

## Why
Code-first: schema is source of truth for validation, docs, and JSON Schema generation.

## Related
- `pkg_config_loader` - Uses Spec for validation
- `pkg_config_merge` - Uses Spec for merging
- `pkg_tools_configdoc` - Generates docs from Spec
