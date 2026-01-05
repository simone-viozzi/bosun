# Config Loader Contract

**Package**: `internal/config/loader`
**File**: `loader.go`
**Status**: Updated (lenient unknown-key handling)

## Function Signature

```go
// FromLabels parses bosun.* labels into a ConfigV1 struct.
// Unknown keys are warnings by default; use opts.Strict for errors.
func FromLabels(spec schema.Spec, labels map[string]string, scope schema.Scope, opts LoadOptions) (schema.ConfigV1, *ValidationErrors)
```

## New Types

```go
// LoadOptions configures label parsing behavior.
type LoadOptions struct {
    // Strict treats unknown bosun.* keys as errors (default: false = warnings)
    Strict bool
}

// ValidationErrors collects all validation issues.
type ValidationErrors struct {
    // Errors are blocking issues (parse failures, scope mismatches)
    Errors []ValidationError

    // Warnings are non-blocking issues (unknown keys in lenient mode)
    Warnings []ValidationError
}

// HasErrors returns true if there are blocking errors.
func (e *ValidationErrors) HasErrors() bool {
    return len(e.Errors) > 0
}

// HasWarnings returns true if there are warnings.
func (e *ValidationErrors) HasWarnings() bool {
    return len(e.Warnings) > 0
}

// All returns combined errors and warnings for display.
func (e *ValidationErrors) All() []ValidationError {
    return append(e.Errors, e.Warnings...)
}
```

## Behavioral Contract

### Unknown Key Handling

| Mode | Unknown `bosun.*` Key | Result |
|------|----------------------|--------|
| Lenient (default) | `bosun.custom.foo` | ⚠️ Warning added to `Warnings` |
| Strict | `bosun.custom.foo` | ❌ Error added to `Errors` |

### Validation Priority

1. **Scope validation** — Key must be allowed on the given scope
2. **Type parsing** — Value must parse to the expected type
3. **Enum validation** — Value must be in allowed set (if enum)
4. **Unknown key check** — Key must exist in spec (warning or error)

### Error Messages

| Condition | Message Format |
|-----------|----------------|
| Unknown key (lenient) | `unknown key: <key> (ignored)` |
| Unknown key (strict) | `unknown key: <key>` |
| Scope mismatch | `key '<key>' not allowed on scope '<scope>'` |
| Parse error | `invalid <type> for key '<key>': <error>` |
| Invalid enum | `invalid enum value '<v>' for key '<k>': must be one of [...]` |

## CLI Integration

```go
// validate.go - Add --strict flag
var strictFlag bool
cmd.Flags().BoolVar(&strictFlag, "strict", false, "Treat unknown bosun.* keys as errors")

// Usage
opts := loader.LoadOptions{Strict: strictFlag}
config, errs := loader.FromLabels(spec, labels, scope, opts)

if errs.HasErrors() {
    // Print errors, exit 2
}
if errs.HasWarnings() {
    // Print warnings to stderr, continue
}
```

## Example Output

### Lenient Mode (default)
```
$ bosun config validate
Warnings:
  container "app-1": unknown key: bosun.custom.foo (ignored)

Validation passed with 1 warning(s)
```

### Strict Mode
```
$ bosun config validate --strict
Validation errors:
  container "app-1": unknown key: bosun.custom.foo

Found 1 error(s)
```
