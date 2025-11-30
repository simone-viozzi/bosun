# Config Loader Package

**Location**: `internal/config/loader/`

## Overview

The `loader` package parses Docker labels into typed configuration values according to the schema specification defined in `internal/config/schema/`.

## Key Types

### `ValidationError` (errors.go)
Single validation failure:
- `Key string` - Label key that failed
- `Value string` - Raw value provided
- `Scope schema.Scope` - Entity scope
- `Message string` - Human-readable error
- `Err error` - Underlying error (for wrapping)

### `ValidationErrors` (errors.go)
Collection of errors implementing `error` interface:
- `Error() string` - Formatted multi-error output
- `HasErrors() bool` - Check if any errors
- `Add(ValidationError)` - Append error
- Helper methods: `AddUnknownKey`, `AddScopeMismatch`, `AddTypeParseFailed`, `AddInvalidEnum`, `AddRequiredMissing`

## Key Functions

### `FromLabels(spec, labels, scope) (ConfigV1, error)` (loader.go)
Main entry point for label parsing:
- Filters `bosun.*` labels only
- Validates keys against spec
- Validates scope (global allowed anywhere)
- Parses values by type (string, bool, int, duration, size, enum, list)
- Returns ALL errors (not just first)

### Parse Functions (parse.go)
- `parseBool(s)` - Via `strconv.ParseBool`
- `parseInt(s)` - Via `strconv.ParseInt`
- `parseDuration(s)` - Via `time.ParseDuration`
- `parseSize(s)` - Via `units.RAMInBytes`
- `parseEnum(s, allowed)` - Exact match against allowed values
- `parseList(s)` - JSON `["a"]` or CSV `a,b` format

## Error Messages

| Error Type | Format |
|------------|--------|
| Unknown key | `unknown key: <key>` |
| Scope mismatch | `key '<key>' not allowed on scope '<scope>'` |
| Parse error | `invalid <type> for key '<key>': <parse_error>` |
| Invalid enum | `invalid enum value '<value>' for key '<key>': must be one of [<values>]` |
| Required missing | `required key '<key>' not provided` |

## Usage

```go
spec := schema.V1Spec()
labels := map[string]string{
    "bosun.container.stopGracePeriod": "30s",
    "bosun.container.autoRestart": "true",
}

cfg, err := loader.FromLabels(spec, labels, schema.ScopeContainer)
if err != nil {
    var verrs loader.ValidationErrors
    if errors.As(err, &verrs) {
        for _, ve := range verrs {
            fmt.Printf("Error: %s\n", ve.Message)
        }
    }
}
```

## Job Label Validation (job_validation.go)

Separate validation for `bosun.job.*` labels (not part of config schema):

### `ValidateJobLabels(entities) JobValidationResult`
Validates job labels across containers and volumes:
- Container labels: `bosun.job.enabled`, `bosun.job.name`, `bosun.job.schedule`
- Volume labels: `bosun.job.attach`, `bosun.job.mount.path`, `bosun.job.mount.mode`

### Error Codes (`JobErrorCode`)
- `JobErrorInvalidEnabled`: Invalid boolean for enabled
- `JobErrorInvalidSchedule`: Invalid cron expression
- `JobErrorMissingName`: Container enabled but no job name
- `JobErrorConflictingField`: Multiple containers define conflicting field values
- `JobErrorOrphanedVolume`: Volume attached to non-existent job
- `JobErrorInvalidAttach`: Invalid attach value on volume
- `JobErrorInvalidMountPath`: Invalid mount path on volume

### Usage
```go
result := loader.ValidateJobLabels(entities)
if !result.IsValid() {
    for _, err := range result.Errors {
        fmt.Printf("[%s] %s: %s\n", err.Code, err.EntityName, err.Message)
    }
}
```

## Test Coverage

All validation paths tested with unit and integration tests.
