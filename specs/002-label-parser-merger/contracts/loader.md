# Loader Package API Contract

**Package**: `internal/config/loader`
**Date**: 2025-11-29

## Public API

### Types

```go
// ValidationError represents a single validation failure.
type ValidationError struct {
    Key     string // Label key that failed (e.g., "bosun.container.unknownKey")
    Value   string // Raw value that was provided
    Scope   schema.Scope // Scope where validation failed
    Message string // Human-readable error message
    Err     error  // Underlying error (optional, for wrapping)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool
```

### Functions

```go
// FromLabels parses Docker labels into a typed ConfigV1 struct.
//
// Parameters:
//   - spec: Schema specification from ParseTags[ConfigV1]()
//   - labels: Raw Docker labels map (e.g., from container inspection)
//   - scope: The entity scope (container, volume, network)
//
// Returns:
//   - cfg: Parsed configuration (partial if errors)
//   - err: ValidationErrors if any validation failed, nil on success
//
// Behavior:
//   - Filters labels to only those with "bosun." prefix
//   - Validates all keys are known in spec
//   - Validates scope matches (global allowed anywhere)
//   - Parses values according to declared types
//   - Returns ALL errors, not just first one
//
// Example:
//   spec, _ := schema.V1Spec()
//   labels := map[string]string{
//       "bosun.container.stopGracePeriod": "30s",
//       "bosun.container.autoRestart": "true",
//   }
//   cfg, err := loader.FromLabels(spec, labels, schema.ScopeContainer)
//   if err != nil {
//       // err is ValidationErrors with all failures
//   }
func FromLabels(spec schema.Spec, labels map[string]string, scope schema.Scope) (schema.ConfigV1, error)
```

## Error Contract

### Unknown Key Error
- **When**: Label key not found in spec
- **Message Format**: `unknown key: <key>`
- **Example**: `unknown key: bosun.container.typoedKey`

### Scope Mismatch Error
- **When**: Field scope doesn't match entity scope (and not global)
- **Message Format**: `key '<key>' not allowed on scope '<scope>'`
- **Example**: `key 'bosun.container.stopGracePeriod' not allowed on scope 'volume'`

### Type Parse Error
- **When**: Value doesn't parse as declared type
- **Message Format**: `invalid <type> for key '<key>': <parse_error>`
- **Examples**:
  - `invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration "invalid"`
  - `invalid bool for key 'bosun.container.autoRestart': strconv.ParseBool: parsing "maybe": invalid syntax`
  - `invalid size for key 'bosun.volume.maxSize': invalid size: 'notasize'`

### Enum Validation Error
- **When**: Value not in allowed enum list
- **Message Format**: `invalid enum value '<value>' for key '<key>': must be one of [<allowed>]`
- **Example**: `invalid enum value 'verbose' for key 'bosun.container.logLevel': must be one of [debug, info, warn, error]`

### Required Field Error
- **When**: Required field has no value
- **Message Format**: `required key '<key>' not provided`
- **Example**: `required key 'bosun.global.instance' not provided`

## Invariants

1. **All Errors Reported**: `FromLabels` always returns all validation errors, never stops at first
2. **No Partial State on Error**: If validation fails, returned ConfigV1 may be partially populated but should not be used
3. **Prefix Filtering**: Only labels starting with `bosun.` are processed; others are silently ignored
4. **Case Sensitivity**: Keys are case-sensitive; `bosun.Container.X` ≠ `bosun.container.x`
5. **Empty Values**: Empty string values for optional fields use defaults; for required fields, they error
