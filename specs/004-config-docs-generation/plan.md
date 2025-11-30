# Implementation Plan: Config Documentation Generation

**Branch**: `004-config-docs-generation` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-config-docs-generation/spec.md`

## Summary

Generate developer documentation (Markdown) and machine-readable validation schema (JSON Schema draft 2020-12) from the code-first `ConfigV1` schema. The generator will read schema metadata via `ParseTags[ConfigV1]()` and produce deterministic, sorted output suitable for version control and CI validation.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Standard library (`encoding/json`, `text/template`, `sort`), existing `internal/config/schema` package
**Storage**: File system output to `docs/` directory
**Testing**: `go test` for unit tests; verify deterministic output, JSON Schema validity
**Target Platform**: Linux/macOS/Windows (Go cross-platform)
**Project Type**: Single project - internal tool/generator
**Performance Goals**: N/A (one-time generation, not runtime critical)
**Constraints**: Output must be deterministic (byte-for-byte identical on repeated runs)
**Scale/Scope**: Single schema (~10-20 fields currently), extensible for future growth

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Hexagonal Architecture** | ✅ PASS | Generator is a build-time tool, not runtime domain logic. Lives in `internal/tools/` (separate from domain). Reads from schema package (port-like interface via `Spec` type). |
| **II. Label-Driven Configuration** | ✅ N/A | Generator documents label config, doesn't consume labels at runtime. |
| **III. Test-First Development** | ✅ PASS | Unit tests for formatters; determinism test; JSON Schema validation test. |
| **IV. CLI-First Interface** | ✅ PASS | Invocable via `make docs` and `go generate`. No interactive prompts. |
| **V. Code Quality & Simplicity** | ✅ PASS | Standard Go, follows naming conventions, documented exports. |

**Gate Result**: ✅ All principles satisfied or N/A. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/004-config-docs-generation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (generator API contract)
└── tasks.md             # Phase 2 output (not created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── config/
│   └── schema/          # Existing - provides ParseTags, Spec, FieldSpec, ConfigV1
└── tools/
    └── configdoc/       # NEW - documentation generator
        ├── doc.go       # Package documentation
        ├── generator.go # Main generator orchestration
        ├── markdown.go  # Markdown formatter
        ├── jsonschema.go # JSON Schema formatter
        └── *_test.go    # Unit tests

docs/                    # NEW - generated output directory
├── config.md            # Generated Markdown documentation
└── config.schema.json   # Generated JSON Schema

cmd/
└── bosun/
    └── main.go          # Add //go:generate directive
```

**Structure Decision**: Generator lives in `internal/tools/configdoc/` following Go conventions for internal tooling. Output goes to `docs/` at project root. This keeps generated artifacts visible and separate from source code.

## Complexity Tracking

> No constitution violations. Table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| - | - | - |
