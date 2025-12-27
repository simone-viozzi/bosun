# Config Merge Package

## Scope
Multi-source configuration merging in `internal/config/merge/`.

## What

### Precedence Order (lowest to highest)
1. **defaults** - Built-in from schema tags
2. **file** - Config file values
3. **env** - Environment variables (disabled by default)
4. **labels** - Docker labels (highest priority)

### Main Function
```go
Merge(spec, defaults, file, env, labels, opts) (ConfigV1, error)
```

### Options
```go
type Options struct {
    EnableEnv bool  // Enable env layer (default: false)
}
```

### Zero Value Handling
| Type | Zero Value |
|------|------------|
| string | `""` |
| bool | `false` |
| int/int64 | `0` |
| time.Duration | `0` |
| []string | `nil` or `[]` |

**Known Limitation**: Cannot distinguish unset `false` from explicit `false`.

### Usage
```go
defaults := schema.V1Defaults()
labelsCfg, _ := loader.FromLabels(spec, labels, scope)
merged, err := merge.Merge(spec, defaults, fileConfig, nil, &labelsCfg, merge.Options{})
```

## Why
Labels have highest priority per design: infrastructure-as-code via Docker labels.

## Related
- `pkg_config_schema` - Provides defaults
- `pkg_config_loader` - Parses labels
- `pkg_cli` - Uses in `config validate --print`
