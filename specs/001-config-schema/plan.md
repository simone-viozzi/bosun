# Implementation Plan: Code-First Config Schema with Tags

**Branch**: `001-config-schema` | **Date**: 2025-11-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-config-schema/spec.md`
**Related Issue**: [#57](https://github.com/simone-viozzi/bosun/issues/57)

## Summary

Define a canonical config schema in Go structs with rich `bosun:` struct tags that serve as the single source of truth for (a) parsing Docker labels into strongly-typed configuration, and (b) later exporting JSON Schema/Markdown documentation. The implementation provides `ParseTags[T]()` for extracting metadata and `DefaultOf[T]()` for hydrating defaults.

## Technical Context

**Language/Version**: Go 1.24.6
**Primary Dependencies**: Standard library (`reflect`, `strings`, `strconv`, `time`), `github.com/docker/go-units` (for byte size parsing)
**Storage**: N/A (in-memory schema definition)
**Testing**: Standard Go testing (`go test`), unit tests with `_test.go` files
**Target Platform**: Linux/macOS/Windows (CLI tool)
**Project Type**: Single Go module with hexagonal architecture
**Performance Goals**: N/A (compile-time/startup-time schema parsing)
**Constraints**: Schema parsing must be deterministic and produce stable output for doc generation
**Scale/Scope**: ~7 config types, ~10-20 config fields in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution template is not fully populated for this project. Based on the project's established patterns from Serena memories:

| Principle | Status | Notes |
|-----------|--------|-------|
| Hexagonal Architecture | ✅ PASS | New `internal/config/schema` package follows domain layer pattern |
| Domain Independence | ✅ PASS | No external dependencies in schema types (only stdlib + go-units) |
| Test-First | ✅ PASS | Unit tests required for tag parsing per spec |
| Code Style | ✅ PASS | Will follow Go conventions, exported types with comments |

## Project Structure

### Documentation (this feature)

```text
specs/001-config-schema/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (N/A - no API contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── config/
│   └── schema/           # NEW: This feature
│       ├── types.go      # Scope, ConfigType, FieldSpec, Spec types
│       ├── tags.go       # Tag parsing logic, ParseTags[T]()
│       ├── defaults.go   # DefaultOf[T]() implementation
│       ├── config_v1.go  # V1 config struct definition
│       ├── types_test.go # Unit tests for types
│       ├── tags_test.go  # Unit tests for tag parsing
│       └── defaults_test.go # Unit tests for defaults hydration
├── domain/
│   └── labels/           # EXISTING: Entity types (Kind, LabeledEntity, Snapshot)
├── ports/                # EXISTING: Interface definitions
└── adapters/             # EXISTING: Docker adapter implementations
```

**Structure Decision**: Following hexagonal architecture, `internal/config/schema` sits in the configuration layer. It depends only on stdlib and go-units. Future loader (#58) and merger (#59) packages will import this schema package.

## Complexity Tracking

> No constitution violations. Design is minimal and follows established patterns.

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Package location | `internal/config/schema` | Matches FR-001, follows project structure |
| Type system | String enums + constants | Matches existing `Kind` pattern in `domain/labels` |
| Generics | `ParseTags[T any]()` | Required by spec FR-005, idiomatic Go 1.18+ |
| Dependencies | stdlib + go-units only | Minimal footprint, go-units already indirect dep |
