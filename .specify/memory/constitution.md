<!--
Sync Impact Report
==================
Version change: 0.0.0 → 1.0.0 (MAJOR - initial constitution)
Modified principles: N/A (initial version)
Added sections:
  - Core Principles (5 principles)
  - Technology Stack section
  - Development Workflow section
  - Governance section
Removed sections: N/A
Templates requiring updates:
  - .specify/templates/plan-template.md: ✅ Compatible (Constitution Check section exists)
  - .specify/templates/spec-template.md: ✅ Compatible (no constitution-specific references)
  - .specify/templates/tasks-template.md: ✅ Compatible (no constitution-specific references)
Follow-up TODOs: None
-->

# Bosun Constitution

## Core Principles

### I. Hexagonal Architecture (NON-NEGOTIABLE)

Bosun MUST follow hexagonal (ports and adapters) architecture principles:

- **Domain Independence**: Business logic in `internal/domain/` MUST NOT depend on external systems or frameworks
- **Ports Define Contracts**: All external interactions MUST be defined as interfaces in `internal/ports/`
- **Adapters Implement Ports**: Concrete implementations in `internal/adapters/` MUST implement port interfaces
- **Dependency Direction**: Dependencies MUST point inward (adapters → ports → domain)
- **Application Orchestration**: `internal/app/` wires together ports and adapters

**Rationale**: Enables testability, technology swapping, and clear separation of concerns.

### II. Label-Driven Configuration

Configuration MUST be primarily driven by Docker labels:

- **Primary Source of Truth**: `bosun.*` labels on containers, volumes, and networks define behavior
- **Schema-First Design**: All label keys MUST be defined in the schema package with type, scope, and documentation
- **Strict Validation**: Unknown label keys MUST cause validation failures (no silent ignoring)
- **Precedence Chain**: defaults < file < env < labels (labels always win)
- **Introspectable**: `bosun config validate` MUST show effective configuration from labels

**Rationale**: Users understand system behavior by inspecting labels, enabling GitOps and declarative configuration.

### III. Test-First Development

All features MUST be developed with tests:

- **Unit Tests**: Standard Go tests (`*_test.go`) for individual components
- **Integration Tests**: Docker Compose-based tests using testcontainers-go for end-to-end validation
- **Build Tags**: Integration tests MUST use `//go:build integration` tag
- **Test Utilities**: Use `internal/testutil/` harness for consistent integration test patterns
- **Coverage Targets**: Aim for meaningful coverage; avoid testing trivial code paths

**Rationale**: Tests document behavior, prevent regressions, and enable confident refactoring.

### IV. CLI-First Interface

All Bosun functionality MUST be accessible via CLI:

- **Cobra Framework**: Use `github.com/spf13/cobra` for command structure
- **Output Protocol**: stdout for results, stderr for errors; support JSON output for scripting
- **Subcommand Hierarchy**: Logical grouping (e.g., `bosun labels snapshot`, `bosun config validate`)
- **Flags Over Prompts**: Prefer flags for automation; interactive prompts only when essential
- **Graceful Errors**: Return friendly error messages with actionable guidance

**Rationale**: CLI enables automation, scripting, and integration with existing DevOps workflows.

### V. Code Quality & Simplicity

Code MUST be idiomatic Go and maintainable:

- **Go Standards**: `go fmt`, `go vet`, and golangci-lint MUST pass
- **Naming Conventions**: Exported identifiers start with capital letters; use camelCase
- **YAGNI**: Don't build features until needed; start simple
- **Documentation**: Exported functions and types MUST have doc comments
- **Pre-commit Hooks**: Run hooks before committing to catch issues early

**Rationale**: Maintainable code enables long-term velocity and onboarding of contributors.

## Technology Stack

Bosun development MUST use these technologies and conventions:

| Category | Technology | Notes |
|----------|------------|-------|
| Language | Go 1.24+ | Module at `github.com/simone-viozzi/bosun` |
| CLI | Cobra | `github.com/spf13/cobra` |
| Docker | Docker SDK | `github.com/docker/docker` |
| Testing | testcontainers-go | Integration tests with Docker Compose |
| Build | Makefile | `make build`, `make test`, `make it` |
| Linting | golangci-lint | Pre-commit hook enforced |
| CI/CD | GitHub Actions | `.github/workflows/ci.yml` |

## Development Workflow

All contributions MUST follow this workflow:

1. **Spec-Driven**: Features start with specs in `specs/###-feature-name/`
2. **Branch Naming**: Use `###-feature-name` pattern matching spec directory
3. **Pre-commit Hooks**: MUST pass before committing
4. **Test Execution**: Run `make test` for unit tests, `make it` for integration tests
5. **Code Review**: PRs require review and CI passing
6. **Serena-First**: Use Serena tools for code navigation and editing when available

## Governance

This constitution supersedes all other development practices for Bosun.

- **Amendments**: Changes to this constitution require documentation of rationale and impact assessment
- **Versioning**: Constitution follows semantic versioning (MAJOR.MINOR.PATCH)
- **Compliance**: All PRs MUST verify compliance with these principles
- **Exceptions**: Any deviation MUST be documented with justification in the PR description
- **Guidance**: See `.github/copilot-instructions.md` for runtime AI development guidance

**Version**: 1.0.0 | **Ratified**: 2025-11-30 | **Last Amended**: 2025-11-30
