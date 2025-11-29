# Research: Code-First Config Schema with Tags

**Feature**: 001-config-schema
**Date**: 2025-11-29

## Research Tasks

### 1. Go Struct Tag Parsing Best Practices

**Decision**: Use standard `reflect` package with custom tag key `bosun`

**Rationale**:
- Go's `reflect.StructTag.Get()` returns the entire tag value for a key
- Custom parsing required for comma-separated key=value pairs within the tag
- Pattern is well-established (see `encoding/json`, `encoding/xml` tags)

**Alternatives Considered**:
- Third-party tag parsing libraries: Rejected - adds dependency for simple task
- Multiple tags per field (e.g., `bosun_key`, `bosun_scope`): Rejected - verbose, harder to read

**Implementation Notes**:
- Parse tag value by splitting on `,` (respecting quoted strings for `doc='...'`)
- Handle edge cases: empty values, missing keys, malformed syntax
- Return structured errors with field name and issue description

### 2. Duration and Size Parsing in Go

**Decision**: Use `time.ParseDuration` for durations, `github.com/docker/go-units` for byte sizes

**Rationale**:
- `time.ParseDuration` is stdlib, handles "30s", "5m", "1h30m" etc.
- `go-units` is already an indirect dependency (via Docker SDK) and handles "1GB", "512MB", "1KiB" etc.
- Matches Docker ecosystem conventions

**Alternatives Considered**:
- Custom size parser: Rejected - reinventing the wheel, go-units is battle-tested
- `dustin/go-humanize`: Rejected - not already in dependency tree

### 3. Generic Function Design in Go

**Decision**: Use type parameter `[T any]` with reflection to iterate struct fields

**Rationale**:
- `ParseTags[T any]() (Spec, error)` - caller specifies config struct type
- `DefaultOf[T any]() T` - returns zero-initialized struct with defaults populated
- Pattern allows multiple config versions (ConfigV1, ConfigV2) without code duplication

**Implementation Pattern**:
```go
func ParseTags[T any]() (Spec, error) {
    var zero T
    t := reflect.TypeOf(zero)
    // ... iterate fields, parse tags
}
```

### 4. Enum Handling Strategy

**Decision**: String-based enums with validation at parse time

**Rationale**:
- Go doesn't have native enums; string constants are idiomatic
- Enum values stored in FieldSpec as `[]string`
- Validation happens in loader (#58), not in schema package

**Pattern**:
```go
type Scope string
const (
    ScopeContainer Scope = "container"
    ScopeVolume    Scope = "volume"
    ScopeNetwork   Scope = "network"
    ScopeGlobal    Scope = "global"
)
```

### 5. Handling Quoted Strings in Tags

**Decision**: Use single quotes for doc strings to allow commas within documentation

**Rationale**:
- Tag format: `bosun:"key=x,scope=y,doc='Description with, commas'"`
- Single quotes are valid in Go struct tags
- Simple state machine parser to handle quoted sections

**Implementation**:
- Split on `,` only when not inside single quotes
- Strip quotes from doc value after extraction

### 6. Embedded Struct Handling

**Decision**: Recursively process embedded structs to flatten field spec

**Rationale**:
- Allows grouping related config fields: `ContainerConfig`, `VolumeConfig` embedded in `ConfigV1`
- Each embedded field's tags are processed the same way
- Keys must be unique across entire struct tree (error on duplicates)

## Key Design Decisions Summary

| Decision | Choice | Why |
|----------|--------|-----|
| Tag format | `bosun:"key=...,scope=..."` | Single tag, clear key-value pairs |
| Duration parsing | `time.ParseDuration` | Stdlib, proven |
| Size parsing | `docker/go-units` | Already in dep tree, Docker ecosystem |
| Scope type | String enum | Matches existing `Kind` pattern |
| ConfigType | String enum | Simple, extensible |
| Defaults hydration | Reflection-based | Type-safe, compile-time struct |
| Error handling | Descriptive errors with field context | Developer experience |

## Open Questions (Resolved)

1. **Q**: Should we validate Go field type matches tag type?
   **A**: Yes, `ParseTags` will error if `type=int` on a `string` field.

2. **Q**: How to handle fields without `bosun:` tag?
   **A**: Skip silently - not all struct fields need to be config keys.

3. **Q**: Should `DefaultOf` error on unparseable defaults?
   **A**: Yes, return error not panic - bad defaults are schema bugs.
