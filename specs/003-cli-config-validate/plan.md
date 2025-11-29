# Implementation Plan: CLI Config Validate Command

**Branch**: `003-cli-config-validate` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-cli-config-validate/spec.md`
**Related Issues**: #60

## Summary

Implement `bosun config validate` CLI command that validates Docker label configuration against the schema, merges configuration from multiple sources (defaults, file, labels), and provides clear error messages for invalid or unknown keys. The command enables users to verify their configuration before deployment without affecting running services.

## Technical Context

**Language/Version**: Go 1.24.6
**Primary Dependencies**: Cobra (CLI), Docker SDK, go-units, existing config packages
**Storage**: N/A (read-only command, no persistence)
**Testing**: Go test + testcontainers-go for integration tests
**Target Platform**: Linux, macOS, Windows (Docker host)
**Project Type**: Single project (hexagonal architecture)
**Performance Goals**: < 5 seconds for typical setups (< 50 containers)
**Constraints**: Read-only operation, must not modify Docker state
**Scale/Scope**: Typical home/small server Docker setups (1-100 containers)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Hexagonal Architecture ✅
- **Compliance**: Command will use existing ports (`LabelSource`) and adapters (`dockerlabels`)
- **Domain Logic**: Validation logic already in `loader` package, merge logic in `merge` package
- **CLI Layer**: New command in `internal/cmd/` following existing `snapshot.go` pattern

### II. Label-Driven Configuration ✅
- **Compliance**: This feature IS the validation mechanism for label-driven config
- **Schema-First**: Uses existing `schema.V1Spec()` for validation rules
- **Strict Validation**: FR-003 requires all validation errors to be reported

### III. Test-First Development ✅
- **Compliance**: Unit tests for command logic, integration tests for Docker interaction
- **Build Tags**: Integration tests will use `//go:build integration`

### IV. CLI-First Interface ✅
- **Compliance**: This IS a CLI feature
- **Cobra Framework**: Will use Cobra like existing commands
- **Output Protocol**: stdout for config, stderr for errors, JSON support via `--print`

### V. Code Quality & Simplicity ✅
- **Compliance**: Follows existing patterns in `internal/cmd/`
- **YAGNI**: Minimal scope - validate only, no side effects

## Project Structure

### Documentation (this feature)

```text
specs/003-cli-config-validate/
├── plan.md              # This file
├── research.md          # Phase 0: Research findings
├── data-model.md        # Phase 1: Type definitions
├── quickstart.md        # Phase 1: Usage examples
├── contracts/           # Phase 1: CLI contract
│   └── cli.md           # Command interface specification
└── tasks.md             # Phase 2: Implementation tasks
```

### Source Code (repository root)

```text
internal/
├── cmd/
│   ├── root.go          # Add NewConfigCmd() registration
│   ├── config.go        # NEW: Config command group
│   └── validate.go      # NEW: Validate subcommand
├── config/
│   ├── schema/          # EXISTING: ConfigV1, V1Spec(), V1Defaults()
│   ├── loader/          # EXISTING: FromLabels(), ValidationErrors
│   └── merge/           # EXISTING: Merge(), Options
├── adapters/
│   └── dockerlabels/    # EXISTING: DockerLabelSource.Snapshot()
├── domain/
│   └── labels/          # EXISTING: Entity, Snapshot types
└── ports/
    └── labels.go        # EXISTING: LabelSource interface

integration/
└── validate_test.go     # NEW: Integration tests for validate command
```

**Structure Decision**: Follows existing hexagonal architecture. New files only in `internal/cmd/` for CLI and `integration/` for tests. All domain/business logic already exists in config packages.

## Complexity Tracking

> No constitution violations - feature follows all principles naturally.

| Aspect | Complexity | Justification |
|--------|------------|---------------|
| New files | Low (2-3 files) | Only CLI layer additions needed |
| Dependencies | None | Uses existing packages |
| Testing | Medium | Needs integration tests with Docker |

## Post-Design Constitution Re-Check

*GATE: Verified after Phase 1 design completion.*

### I. Hexagonal Architecture ✅ VERIFIED
- **Design Review**: `data-model.md` defines only CLI-layer types (`ValidateOptions`, `ValidationResult`)
- **No Domain Leakage**: All business logic remains in existing packages (loader, merge, schema)
- **Adapter Reuse**: Uses existing `dockerlabels.DockerLabelSource` without modification

### II. Label-Driven Configuration ✅ VERIFIED
- **Contract Review**: `contracts/cli.md` defines strict validation behavior
- **Error Messages**: All error types documented with actionable messages
- **Precedence**: Merge order (defaults → file → labels) clearly specified

### III. Test-First Development ✅ VERIFIED
- **Test Plan**: Integration tests with Docker Compose defined in project structure
- **Coverage Strategy**: Unit tests for CLI logic, integration tests for full flow

### IV. CLI-First Interface ✅ VERIFIED
- **Contract Complete**: Full CLI contract in `contracts/cli.md`
- **Exit Codes**: Defined (0=success, 1=validation error, 2=runtime error)
- **Output Streams**: stdout/stderr separation documented
- **JSON Support**: `--print` flag for machine-readable output

### V. Code Quality & Simplicity ✅ VERIFIED
- **Minimal Scope**: Only 2-3 new files, all in CLI layer
- **Pattern Compliance**: Follows existing `snapshot.go` patterns
- **No Over-engineering**: Reuses existing types, no unnecessary abstractions

## Generated Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Implementation Plan | `specs/003-cli-config-validate/plan.md` | ✅ Complete |
| Research | `specs/003-cli-config-validate/research.md` | ✅ Complete |
| Data Model | `specs/003-cli-config-validate/data-model.md` | ✅ Complete |
| CLI Contract | `specs/003-cli-config-validate/contracts/cli.md` | ✅ Complete |
| Quickstart | `specs/003-cli-config-validate/quickstart.md` | ✅ Complete |

## Next Steps

Run `/speckit.tasks` to generate implementation tasks based on this plan.
