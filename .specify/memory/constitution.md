<!--
Sync Impact Report
==================
Version change: 1.0.0 → 1.1.0 (MINOR - consolidate with Serena memories)
Modified principles: None (content unchanged, references added)
Added sections: Memory References section
Removed sections: None (details moved to memories, principles remain)
Templates requiring updates: None
Follow-up TODOs: None
-->

# Bosun Constitution

This document defines the **non-negotiable principles** for Bosun development.
For implementation details, see the referenced Serena memory files.

## Core Principles

### I. Hexagonal Architecture (NON-NEGOTIABLE)

Bosun MUST follow hexagonal (ports and adapters) architecture principles:

- **Domain Independence**: Business logic in `internal/domain/` MUST NOT depend on external systems
- **Ports Define Contracts**: All external interactions MUST be defined as interfaces in `internal/ports/`
- **Adapters Implement Ports**: Concrete implementations in `internal/adapters/` MUST implement port interfaces
- **Dependency Direction**: Dependencies MUST point inward (adapters → ports → domain)

**Rationale**: Enables testability, technology swapping, and clear separation of concerns.

📚 **Details**: See `architecture` memory

### II. Label-Driven Configuration

Configuration MUST be primarily driven by Docker labels:

- **Primary Source**: `bosun.*` labels on containers, volumes, and networks define behavior
- **Schema-First**: All label keys MUST be defined in the schema package with type, scope, and documentation
- **Strict Validation**: Unknown label keys MUST cause validation failures (no silent ignoring)
- **Precedence**: defaults < file < env < labels (labels always win)

**Rationale**: Users understand system behavior by inspecting labels, enabling GitOps and declarative configuration.

📚 **Details**: See `config_schema_package`, `config_loader_package`, `config_merge_package` memories

### III. Test-First Development

All features MUST be developed with tests:

- **Unit Tests**: Standard Go tests (`*_test.go`) for individual components
- **Integration Tests**: Docker Compose-based tests with `//go:build integration` tag
- **Coverage**: Aim for meaningful coverage; avoid testing trivial code paths

**Rationale**: Tests document behavior, prevent regressions, and enable confident refactoring.

📚 **Details**: See `testing_structure` memory

### IV. CLI-First Interface

All Bosun functionality MUST be accessible via CLI:

- **Cobra Framework**: Use `github.com/spf13/cobra` for command structure
- **Output Protocol**: stdout for results, stderr for errors; support JSON output for scripting
- **Flags Over Prompts**: Prefer flags for automation; interactive prompts only when essential

**Rationale**: CLI enables automation, scripting, and integration with existing DevOps workflows.

📚 **Details**: See `cli_commands` memory

### V. Code Quality & Simplicity

Code MUST be idiomatic Go and maintainable:

- **Go Standards**: `go fmt`, `go vet`, and golangci-lint MUST pass
- **YAGNI**: Don't build features until needed; start simple
- **Pre-commit Hooks**: Run hooks before committing to catch issues early

**Rationale**: Maintainable code enables long-term velocity and onboarding of contributors.

📚 **Details**: See `code_style`, `task_completion` memories

## Technology Stack

| Category | Technology | Notes |
|----------|------------|-------|
| Language | Go 1.24+ | Module at `github.com/simone-viozzi/bosun` |
| CLI | Cobra | `github.com/spf13/cobra` |
| Docker | Docker SDK | `github.com/docker/docker` |
| Testing | testcontainers-go | Integration tests with Docker Compose |
| Build | Makefile | `make build`, `make test`, `make it` |
| Linting | golangci-lint | Pre-commit hook enforced |
| CI/CD | GitHub Actions | `.github/workflows/ci.yml` |

📚 **Commands**: See `suggested_commands` memory

## Memory Reference Index

| Memory File | Contents |
|-------------|----------|
| `project_overview` | High-level project structure and status |
| `architecture` | Hexagonal architecture implementation details |
| `config_schema_package` | Config schema types, tags, and FieldSpec |
| `config_loader_package` | Label parsing and validation errors |
| `config_merge_package` | Multi-source config merging and precedence |
| `cli_commands` | Available CLI commands and flags |
| `testing_structure` | Test infrastructure and patterns |
| `code_style` | Go conventions and formatting |
| `task_completion` | Post-task checklist |
| `suggested_commands` | Common development commands |
| `dockerlabels_adapter` | Docker label discovery adapter |
| `label_discovery_domain` | Domain models for labels |
| `config_docs_generator` | Auto-generated docs tooling |

## Governance

This constitution supersedes all other development practices for Bosun.

- **Amendments**: Changes require documentation of rationale and impact assessment
- **Versioning**: Constitution follows semantic versioning (MAJOR.MINOR.PATCH)
- **Compliance**: All PRs MUST verify compliance with these principles
- **Exceptions**: Any deviation MUST be documented with justification in the PR description
- **Guidance**: See `.github/copilot-instructions.md` for runtime AI development guidance

**Version**: 1.1.0 | **Ratified**: 2025-11-30 | **Last Amended**: 2025-11-30
