# Implementation Plan: Label Parser and Source Merger

**Branch**: `002-label-parser-merger` | **Date**: 2025-11-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-label-parser-merger/spec.md`
**Related Issues**: #58 (loader), #59 (merger)

## Summary

Implement label parsing and config merging for Bosun's configuration system:
1. **Loader Package** (`internal/config/loader`): Parse Docker labels into typed ConfigV1 with strict validation (fail on unknown keys, scope mismatches, type errors)
2. **Merge Package** (`internal/config/merge`): Merge configs from multiple sources with deterministic precedence (defaults < file < env < labels)

Builds on existing schema package (#57) which provides `ParseTags[T]()`, `DefaultOf[T]()`, and type definitions.

## Technical Context

**Language/Version**: Go 1.24.6
**Primary Dependencies**:
- `github.com/docker/go-units` (already in go.mod) - byte size parsing
- `time` (stdlib) - duration parsing
- `encoding/json` (stdlib) - JSON array list parsing
- `reflect` (stdlib) - config merging

**Storage**: N/A (in-memory config)
**Testing**: Go standard testing (`go test`)
**Target Platform**: Linux (Docker host)
**Project Type**: Single Go module
**Performance Goals**: Parse labels in <1ms for typical configs (~20 keys)
**Constraints**: No external config libraries; use stdlib + existing code
**Scale/Scope**: ~20-30 config keys in v1, expandable

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution is a template (not filled in). Following general best practices:

- [x] **Hexagonal Architecture**: New packages fit cleanly as internal config logic, not adapters
- [x] **Test Coverage**: Plan includes comprehensive unit tests (>80% coverage target)
- [x] **No External Dependencies**: Using stdlib + existing dependencies only
- [x] **Separation of Concerns**: Loader (parsing) and Merger (combining) are separate packages

## Project Structure

### Documentation (this feature)

```text
specs/002-label-parser-merger/
├── plan.md              # This file
├── research.md          # Research decisions
├── data-model.md        # Entity definitions
├── quickstart.md        # Usage examples
├── contracts/           # API contracts
│   ├── loader.md
│   └── merge.md
└── tasks.md             # Implementation tasks (next phase)
```

### Source Code (repository root)

```text
internal/config/
├── schema/              # Existing (from #57)
│   ├── types.go         # Scope, ConfigType, FieldSpec, Spec
│   ├── tags.go          # ParseTags[T]()
│   ├── defaults.go      # DefaultOf[T](), parse helpers
│   └── config_v1.go     # ConfigV1 struct
├── loader/              # NEW: Label parsing
│   ├── doc.go           # Package documentation
│   ├── errors.go        # ValidationError, ValidationErrors
│   ├── errors_test.go   # Error type tests
│   ├── loader.go        # FromLabels()
│   └── loader_test.go   # Loader tests
└── merge/               # NEW: Config merging
    ├── doc.go           # Package documentation
    ├── merge.go         # Merge(), Options
    └── merge_test.go    # Merge tests
```

**Structure Decision**: Two new packages under `internal/config/` matching the issue split (#58=loader, #59=merger). Reuses existing `schema` package types and parsers.

## Key Implementation Details

### Loader Package

**File: `internal/config/loader/errors.go`**
- `ValidationError` struct with Key, Value, Scope, Message, Err fields
- `ValidationErrors` slice type implementing `error` interface
- `Error()` method formats all errors with newlines
- `HasErrors()` helper method

**File: `internal/config/loader/loader.go`**
- `FromLabels(spec, labels, scope) (ConfigV1, error)` main function
- Internal helpers:
  - `filterLabels(labels, prefix)` - filter to bosun.* only
  - `validateKey(spec, key)` - check key exists
  - `validateScope(fieldScope, entityScope)` - scope check
  - `parseValue(fieldSpec, value)` - type-specific parsing
  - `setField(cfg, fieldSpec, value)` - reflection-based field setting

**Parsing Functions** (reuse from schema/defaults.go or reimplement):
- `parseBool(s)` - uses strconv.ParseBool
- `parseInt(s)` - uses strconv.Atoi
- `parseDuration(s)` - uses time.ParseDuration
- `parseSize(s)` - uses units.RAMInBytes
- `parseEnum(s, allowed)` - exact match in slice
- `parseList(s)` - JSON if `[...]`, else CSV

### Merge Package

**File: `internal/config/merge/merge.go`**
- `Options` struct with `EnableEnv bool`
- `Merge(spec, defaults, file, env, labels, opts) (ConfigV1, error)` main function
- Internal helpers:
  - `mergeLayer(base, override, spec)` - merge single layer using reflection
  - `isZeroValue(v reflect.Value)` - check if value is zero
  - `shouldOverride(fieldType, value)` - determine if override applies

**Merge Logic**:
1. Start with `defaults` (required, not nil)
2. If `file != nil`, merge file layer
3. If `opts.EnableEnv && env != nil`, merge env layer
4. If `labels != nil`, merge labels layer
5. Return final merged config

## Test Strategy

### Unit Tests

**loader_test.go**:
- TestFromLabels_ValidTypes (all 7 types)
- TestFromLabels_UnknownKey
- TestFromLabels_ScopeMismatch
- TestFromLabels_InvalidDuration
- TestFromLabels_InvalidBool
- TestFromLabels_InvalidSize
- TestFromLabels_InvalidEnum
- TestFromLabels_RequiredMissing
- TestFromLabels_JSONArrayList
- TestFromLabels_CSVList
- TestFromLabels_EmptyLabels
- TestFromLabels_MultipleErrors

**errors_test.go**:
- TestValidationError_Error
- TestValidationErrors_Error
- TestValidationErrors_HasErrors

**merge_test.go**:
- TestMerge_DefaultsOnly
- TestMerge_LabelsOverrideDefaults
- TestMerge_FileOverridesDefaults
- TestMerge_LabelsOverrideFile
- TestMerge_EnvDisabled
- TestMerge_EnvEnabled
- TestMerge_NilLayers
- TestMerge_Deterministic

### Coverage Target
- `loader` package: >85%
- `merge` package: >80%
- Overall config packages: >80%

## Complexity Tracking

No constitution violations identified. Design follows existing patterns.

## Dependencies

### Upstream (this feature depends on)
- [x] #57 - Config schema package (completed)

### Downstream (depends on this feature)
- #60 - CLI `bosun config validate` command
- #62 - Integration tests for typed settings
