# Merge Package API Contract

**Package**: `internal/config/merge`
**Date**: 2025-11-29

## Public API

### Types

```go
// Options configures the merge behavior.
type Options struct {
    // EnableEnv enables the environment variable layer.
    // When false, env values are ignored in merge precedence.
    // Default: false (disabled in v1)
    EnableEnv bool
}
```

### Functions

```go
// Merge combines configuration from multiple sources with defined precedence.
//
// Precedence (lowest to highest):
//   1. defaults - Built-in defaults from schema tags
//   2. file - Config file values (if provided)
//   3. env - Environment variables (if opts.EnableEnv && env != nil)
//   4. labels - Docker label values (if provided)
//
// Parameters:
//   - spec: Schema specification for field metadata
//   - defaults: Base configuration with default values (required)
//   - file: Config loaded from file (may be nil)
//   - env: Config loaded from environment (may be nil, ignored if !opts.EnableEnv)
//   - labels: Config loaded from Docker labels (may be nil)
//   - opts: Merge options
//
// Returns:
//   - merged: The final merged configuration
//   - err: Error if merge fails (should be rare; validation happens in loader)
//
// Behavior:
//   - For each field, higher precedence non-zero values override lower
//   - Zero values in higher layers are treated as "not set" (don't override)
//   - Nil layers are skipped entirely
//   - Output is deterministic for identical inputs
//
// Example:
//   spec, _ := schema.V1Spec()
//   defaults, _ := schema.V1Defaults()
//   labelsCfg, _ := loader.FromLabels(spec, labels, scope)
//
//   merged, err := merge.Merge(spec, defaults, nil, nil, &labelsCfg, merge.Options{})
func Merge(spec schema.Spec, defaults schema.ConfigV1, file, env, labels *schema.ConfigV1, opts Options) (schema.ConfigV1, error)
```

## Merge Rules

### Precedence Table

| Priority | Source | Typical Use |
|----------|--------|-------------|
| 1 (lowest) | defaults | Safe fallbacks from schema |
| 2 | file | User's config file |
| 3 | env | Deployment overrides (optional) |
| 4 (highest) | labels | Container-specific settings |

### Zero Value Handling

For each field type, zero values are considered "not set":

| Type | Zero Value | Considered Set? |
|------|------------|-----------------|
| string | `""` | No |
| bool | `false` | **Yes** (special case) |
| int | `0` | No |
| time.Duration | `0` | No |
| int64 (size) | `0` | No |
| []string | `nil` or `[]string{}` | No |

**Special Case - Bool**: Since `false` is a meaningful value, bool fields are handled with pointer awareness in the merge:
- If a higher layer explicitly sets `false`, it overrides lower `true`
- This requires tracking which fields were actually parsed vs defaulted
- For v1, we track this via the `ParsedKeys` in `loader.ParsedConfig`

### Nil Layer Handling

```go
// Each nil layer is simply skipped:
if file != nil {
    merged = mergeLayer(merged, file)
}
if opts.EnableEnv && env != nil {
    merged = mergeLayer(merged, env)
}
if labels != nil {
    merged = mergeLayer(merged, labels)
}
```

## Invariants

1. **Deterministic**: `Merge(same inputs) → same output` always
2. **Nil-Safe**: Any combination of nil layers is valid
3. **Defaults Required**: `defaults` parameter must not be nil
4. **No Validation**: Merge assumes inputs are already validated; validation happens in loader
5. **Feature Flag**: Env layer is opt-in via `Options.EnableEnv`

## Example Scenarios

### Scenario 1: Labels Override Defaults
```
defaults: { stopGracePeriod: 10s, autoRestart: false }
labels:   { stopGracePeriod: 30s }
result:   { stopGracePeriod: 30s, autoRestart: false }
```

### Scenario 2: File Overrides Defaults, Labels Override File
```
defaults: { stopGracePeriod: 10s }
file:     { stopGracePeriod: 20s }
labels:   { stopGracePeriod: 30s }
result:   { stopGracePeriod: 30s }
```

### Scenario 3: No Labels, File Wins
```
defaults: { autoRestart: false }
file:     { autoRestart: true }
labels:   nil
result:   { autoRestart: true }
```

### Scenario 4: Env Layer Disabled
```
defaults:  { stopGracePeriod: 10s }
env:       { stopGracePeriod: 25s }  // ignored
labels:    nil
opts:      { EnableEnv: false }
result:    { stopGracePeriod: 10s }  // env ignored
```
