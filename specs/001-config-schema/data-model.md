# Data Model: Code-First Config Schema

**Feature**: 001-config-schema
**Date**: 2025-11-29

## Entity Definitions

### Scope

Enumeration of valid scopes indicating where a Docker label can be applied.

| Value | Description |
|-------|-------------|
| `container` | Label applies to containers |
| `volume` | Label applies to volumes |
| `network` | Label applies to networks |
| `global` | Label applies to any entity type |

**Validation**: Must be one of the four defined values.

---

### ConfigType

Enumeration of supported configuration value types.

| Value | Go Type | Parsing |
|-------|---------|---------|
| `string` | `string` | Direct assignment |
| `bool` | `bool` | `strconv.ParseBool` |
| `int` | `int` | `strconv.Atoi` |
| `duration` | `time.Duration` | `time.ParseDuration` |
| `size` | `int64` (bytes) | `go-units.RAMInBytes` |
| `enum` | `string` | Must match one of `Enum` values |
| `list` | `[]string` | CSV split or JSON array |

---

### FieldSpec

Metadata for a single configuration field extracted from struct tags.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `Key` | `string` | Yes | Label key (e.g., `bosun.container.stopGracePeriod`) |
| `Scope` | `Scope` | Yes | Where label can be applied |
| `Type` | `ConfigType` | Yes | Value type for parsing |
| `Default` | `string` | No | Default value (raw string, parsed by loader) |
| `Enum` | `[]string` | No | Allowed values if Type is `enum` |
| `Required` | `bool` | No | Whether field must be present (default: false) |
| `Doc` | `string` | No | Human-readable description |
| `Deprecated` | `bool` | No | Whether field is deprecated (default: false) |
| `FieldName` | `string` | Yes | Go struct field name (for reflection) |

**Invariants**:
- `Key` must start with `bosun.`
- If `Type` is `enum`, `Enum` must be non-empty
- `Key` must be unique within a Spec

---

### Spec

A map from label key to FieldSpec, representing the complete schema definition.

| Attribute | Type | Description |
|-----------|------|-------------|
| (map key) | `string` | Label key (e.g., `bosun.container.stopGracePeriod`) |
| (map value) | `FieldSpec` | Field metadata |

**Operations**:
- `Keys() []string` - Returns all keys in deterministic order
- `Get(key string) (FieldSpec, bool)` - Returns field spec by key
- `Scopes() map[Scope][]FieldSpec` - Groups fields by scope

---

### ConfigV1

The v1 configuration struct demonstrating all supported types.

```
ConfigV1
├── Global (embedded)
│   └── Instance: string (bosun.instance)
├── Container (embedded)
│   ├── StopGracePeriod: time.Duration (bosun.container.stopGracePeriod)
│   ├── HealthCheckInterval: time.Duration (bosun.container.healthCheckInterval)
│   ├── AutoRestart: bool (bosun.container.autoRestart)
│   └── LogLevel: string/enum (bosun.container.logLevel)
├── Volume (embedded)
│   ├── BackupEnabled: bool (bosun.volume.backupEnabled)
│   └── MaxSize: int64 (bosun.volume.maxSize)
└── Network (embedded)
    └── Priority: int (bosun.network.priority)
```

**Notes**:
- Embedded structs organize fields by scope
- Each field has exactly one `bosun:` tag
- Fields without tags are allowed but not included in Spec

## Relationships

```
┌─────────────┐
│  ConfigV1   │ (concrete struct with bosun: tags)
└──────┬──────┘
       │ ParseTags[ConfigV1]()
       ▼
┌─────────────┐
│    Spec     │ (map[string]FieldSpec)
└──────┬──────┘
       │ used by
       ▼
┌─────────────────────────────────────────┐
│  Loader (#58) │ Merger (#59) │ Docs (#61) │
└─────────────────────────────────────────┘
```

## Tag Format Specification

```
bosun:"key=<label_key>,scope=<scope>,type=<type>[,default=<value>][,enum=<a|b|c>][,required=true][,doc='<description>'][,deprecated=true]"
```

### Required Components
- `key=<label_key>` - The Docker label key (must start with `bosun.`)
- `scope=<scope>` - One of: `container`, `volume`, `network`, `global`
- `type=<type>` - One of: `string`, `bool`, `int`, `duration`, `size`, `enum`, `list`

### Optional Components
- `default=<value>` - Default value as string (parsed according to type)
- `enum=<a|b|c>` - Pipe-separated allowed values (required if type=enum)
- `required=true` - Mark field as required (default: false)
- `doc='<description>'` - Human-readable description (single quotes allow commas)
- `deprecated=true` - Mark field as deprecated (default: false)

### Examples

```go
// Duration with default
StopGracePeriod time.Duration `bosun:"key=bosun.container.stopGracePeriod,scope=container,type=duration,default=30s,doc='Grace period before force stopping'"`

// Enum with allowed values
LogLevel string `bosun:"key=bosun.container.logLevel,scope=container,type=enum,enum=debug|info|warn|error,default=info"`

// Required string
Instance string `bosun:"key=bosun.instance,scope=global,type=string,required=true,doc='Unique instance identifier'"`

// Byte size
MaxSize int64 `bosun:"key=bosun.volume.maxSize,scope=volume,type=size,default=1GB"`

// Boolean with default true
AutoRestart bool `bosun:"key=bosun.container.autoRestart,scope=container,type=bool,default=true"`
```
