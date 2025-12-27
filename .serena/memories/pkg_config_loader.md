# Config Loader Package

## Scope
Label parsing and validation in `internal/config/loader/`.

## What

### Main Function
```go
FromLabels(spec, labels, scope) (ConfigV1, error)
```
- Filters `bosun.*` labels
- Validates keys against spec
- Validates scope (global allowed anywhere)
- Parses values by type
- Returns ALL errors (not just first)

### Parse Functions
- `parseBool(s)` - `strconv.ParseBool`
- `parseInt(s)` - `strconv.ParseInt`
- `parseDuration(s)` - `time.ParseDuration`
- `parseSize(s)` - `units.RAMInBytes` (e.g., "10GB")
- `parseEnum(s, allowed)` - Exact match
- `parseList(s)` - JSON `["a"]` or CSV `a,b`

### Error Types

**`ValidationError`** - Single failure
- `Key`, `Value`, `Scope`, `Message`, `Err`

**`ValidationErrors`** - Collection with helpers
- `AddUnknownKey`, `AddScopeMismatch`, `AddTypeParseFailed`, etc.

### Error Messages
| Type | Format |
|------|--------|
| Unknown key | `unknown key: <key>` |
| Scope mismatch | `key '<key>' not allowed on scope '<scope>'` |
| Parse error | `invalid <type> for key '<key>': <error>` |
| Invalid enum | `invalid enum value '<v>' for key '<k>': must be one of [...]` |

### Job Label Validation (job_validation.go)
```go
ValidateJobLabels(entities) JobValidationResult
```
- Validates `bosun.job.*` labels on containers and volumes
- Returns `JobValidationResult` with `Errors` and `Warnings`
- Uses `JobValidationError` (has `Code` field, unlike `ValidationError`)
- Error codes: `JobErrorInvalidEnabled`, `JobErrorMissingName`, `JobErrorConflictingField`, `JobErrorOrphanedVolume`, etc.

## Why
Separates parsing from schema. Returns all errors for better UX.

## Related
- `pkg_config_schema` - Provides Spec
- `pkg_config_merge` - Uses parsed config
- `pkg_cli` - Uses in `config validate`
