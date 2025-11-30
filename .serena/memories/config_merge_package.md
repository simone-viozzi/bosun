# Config Merge Package

**Location**: `internal/config/merge/`

## Overview

The `merge` package combines configuration from multiple sources with defined precedence.

## Precedence Order

1. **defaults** (lowest) - Built-in defaults from schema tags
2. **file** - Config file values
3. **env** - Environment variables (optional, disabled by default)
4. **labels** (highest) - Docker label values

## Key Types

### `Options` (merge.go)
```go
type Options struct {
    EnableEnv bool  // Enable env layer (default: false)
}
```

## Key Functions

### `Merge(spec, defaults, file, env, labels, opts) (ConfigV1, error)` (merge.go)
Main entry point for config merging:
- Starts with defaults as base
- Applies file layer if non-nil
- Applies env layer if `opts.EnableEnv && env != nil`
- Applies labels layer if non-nil (highest priority)
- Non-zero values in higher layers override lower

### Helper Functions
- `mergeLayer(base, override ConfigV1) ConfigV1` - Merge two configs using reflection
- `isZeroValue(v reflect.Value) bool` - Check if value is zero/unset

## Zero Value Handling

| Type | Zero Value | Notes |
|------|------------|-------|
| string | `""` | Empty string = not set |
| bool | `false` | Limitation: can't distinguish unset from explicit false |
| int/int64 | `0` | Zero = not set |
| time.Duration | `0` | Zero = not set |
| []string | `nil` or `[]` | Empty/nil = not set |

**Known Limitation**: Boolean fields treat `false` as "not set" since there's no way to distinguish between an unset field and an explicitly set `false` without additional tracking.

## Usage

```go
spec := schema.V1Spec()
defaults := schema.V1Defaults()
labelsCfg, _ := loader.FromLabels(spec, labels, scope)

merged, err := merge.Merge(
    spec,
    defaults,
    fileConfig,    // may be nil
    nil,           // env not used
    &labelsCfg,
    merge.Options{EnableEnv: false},
)
```

## Test Coverage

Coverage: 76.3%
Tests verify:
- Defaults only
- Labels override defaults
- File overrides defaults
- Labels override file
- Nil layers handling
- Deterministic output
- Env disabled/enabled
- Env does not override labels
- Full precedence chain
