# Data Model: Label Parser and Source Merger

**Feature Branch**: `002-label-parser-merger`
**Date**: 2025-11-29
**Spec**: [spec.md](./spec.md)

## Entities

### Existing (from #57 - Schema Package)

#### Spec
Map from label key to field metadata.
- **Source**: `internal/config/schema/types.go`
- **Type**: `map[string]FieldSpec`
- **Methods**: `Keys()`, `Get()`, `Scopes()`

#### FieldSpec
Metadata for a single config field.
- **Fields**:
  - `Key` (string): Label key like `bosun.container.stopGracePeriod`
  - `Scope` (Scope): Where label applies (container/volume/network/global)
  - `Type` (ConfigType): How to parse (string/bool/int/duration/size/enum/list)
  - `Default` (string): Default value as string
  - `Enum` ([]string): Allowed values for enum type
  - `Required` (bool): Whether field must be present
  - `Doc` (string): Human description
  - `Deprecated` (bool): Whether deprecated
  - `FieldName` (string): Go struct field name

#### ConfigV1
The typed configuration struct.
- **Source**: `internal/config/schema/config_v1.go`
- **Embedded**: `GlobalConfig`, `ContainerConfig`, `VolumeConfig`, `NetworkConfig`

---

### New Entities (This Feature)

#### ValidationError
A single validation error with context.

| Field | Type | Description |
|-------|------|-------------|
| Key | string | The label key that failed validation |
| Value | string | The raw value that was provided |
| Scope | Scope | The scope where validation failed |
| Message | string | Human-readable error message |
| Err | error | Underlying error (optional) |

#### ValidationErrors
Collection of validation errors.

| Field | Type | Description |
|-------|------|-------------|
| errors | []ValidationError | List of all validation failures |

**Methods**:
- `Error() string` - Implements error interface, lists all errors
- `Add(ValidationError)` - Add an error to the collection
- `HasErrors() bool` - Check if any errors exist

#### LoaderOptions
Options for the label loader.

| Field | Type | Description |
|-------|------|-------------|
| Spec | Spec | Schema spec for validation |
| Scope | Scope | Entity scope (container/volume/network) |
| Labels | map[string]string | Raw labels to parse |
| Prefix | string | Label prefix filter (default: "bosun.") |

#### MergeOptions
Options for config merging.

| Field | Type | Description |
|-------|------|-------------|
| Spec | Spec | Schema spec for field metadata |
| EnableEnv | bool | Whether to include env layer (feature flag) |

#### ParsedConfig
Result of parsing labels, before merging.

| Field | Type | Description |
|-------|------|-------------|
| Config | ConfigV1 | Parsed config values |
| Source | ConfigSource | Where values came from |
| ParsedKeys | []string | Which keys were successfully parsed |

#### ConfigSource
Enum for config sources.

| Value | Description |
|-------|-------------|
| SourceDefaults | Built-in defaults from schema |
| SourceFile | Config file (YAML/JSON) |
| SourceEnv | Environment variables |
| SourceLabels | Docker labels |

---

## Relationships

```
┌─────────────────┐     uses      ┌──────────────┐
│     Spec        │◄──────────────│    Loader    │
│  (from schema)  │               │              │
└─────────────────┘               └──────┬───────┘
                                         │
                                         │ produces
                                         ▼
                                  ┌──────────────┐
                                  │ ParsedConfig │
                                  │  + Errors    │
                                  └──────┬───────┘
                                         │
                                         │ fed into
                                         ▼
┌─────────────────┐     uses      ┌──────────────┐
│     Spec        │◄──────────────│    Merger    │
│  (from schema)  │               │              │
└─────────────────┘               └──────┬───────┘
                                         │
                                         │ produces
                                         ▼
                                  ┌──────────────┐
                                  │   ConfigV1   │
                                  │  (merged)    │
                                  └──────────────┘
```

---

## Validation Rules

### From Spec

1. **Unknown Keys**: Any `bosun.*` key not in Spec → error
2. **Scope Mismatch**: Field scope ≠ entity scope (unless global) → error
3. **Type Mismatch**: Value doesn't parse as declared type → error
4. **Enum Invalid**: Value not in allowed enum list → error
5. **Required Missing**: Required field with no value → error

### Type Parsing Rules

| Type | Parser | Valid Examples | Invalid Examples |
|------|--------|----------------|------------------|
| string | as-is | any | (none) |
| bool | strconv.ParseBool | `true`, `false`, `1`, `0` | `yes`, `maybe` |
| int | strconv.Atoi | `42`, `-1`, `0` | `3.14`, `abc` |
| duration | time.ParseDuration | `30s`, `5m`, `1h30m` | `30`, `invalid` |
| size | units.RAMInBytes | `100MB`, `1GiB`, `500` | `big`, `-1GB` |
| enum | exact match in list | `debug`, `info`, `warn` | `verbose`, `DEBUG` |
| list | CSV or JSON array | `a,b,c`, `["a","b"]` | malformed JSON |

---

## State Transitions

### Label Parsing Flow

```
Labels (map[string]string)
    │
    ▼
┌───────────────────┐
│ Filter by prefix  │  (keep only bosun.*)
└────────┬──────────┘
         │
         ▼
┌───────────────────┐
│ Validate keys     │  (unknown? → error)
└────────┬──────────┘
         │
         ▼
┌───────────────────┐
│ Validate scopes   │  (mismatch? → error)
└────────┬──────────┘
         │
         ▼
┌───────────────────┐
│ Parse values      │  (type error? → error)
└────────┬──────────┘
         │
         ▼
┌───────────────────┐
│ Set struct fields │  (via reflection)
└────────┬──────────┘
         │
         ▼
ParsedConfig + ValidationErrors
```

### Config Merge Flow

```
┌──────────────┐
│   Defaults   │  ← Start with defaults
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Merge File   │  ← Override non-zero file values
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Merge Env    │  ← Override non-zero env values (if enabled)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Merge Labels │  ← Override non-zero label values
└──────┬───────┘
       │
       ▼
Final ConfigV1
```
