<!--
Sync Impact Report
==================
Version change: 1.1.0 → 1.2.0 (MINOR - align with current Serena memories, update principles)
Modified principles:
  - II: Updated validation behavior (lenient by default, --strict for errors)
  - III: Clarified integration test patterns
Added sections: None
Removed sections: None
Templates requiring updates: None
Follow-up TODOs: None
-->

# Bosun Constitution

This document defines the **non-negotiable principles** for Bosun development.
For implementation details, see the referenced Serena memory files in `.serena/memories/`.

## Core Principles

### I. Hexagonal Architecture (NON-NEGOTIABLE)

Bosun MUST follow hexagonal (ports and adapters) architecture principles:

- **Domain Independence**: Business logic in `internal/domain/` MUST NOT depend on external systems
- **Ports Define Contracts**: All external interactions MUST be defined as interfaces in `internal/ports/`
- **Adapters Implement Ports**: Concrete implementations in `internal/adapters/` MUST implement port interfaces
- **Dependency Direction**: Dependencies MUST point inward (adapters → ports → domain)
- **CLI is Thin**: CLI layer MUST only parse flags and delegate to app layer; no business logic in `internal/cmd/`

**Rationale**: Enables testability, technology swapping, and clear separation of concerns.

📚 **Details**: See `arch_overview`, `pkg_ports` memories

### II. Label-Driven Configuration

Configuration MUST be primarily driven by Docker labels:

- **Primary Source**: `bosun.*` labels on containers, volumes, and networks define behavior
- **Schema-First**: All label keys MUST be defined in the schema package with type, scope, and documentation
- **Lenient Validation**: Unknown `bosun.*` keys generate warnings by default; use `--strict` for errors
- **Precedence**: defaults < file < env < labels (labels always win)

**Rationale**: Users understand system behavior by inspecting labels, enabling GitOps and declarative configuration.

📚 **Details**: See `pkg_config_schema`, `pkg_config_loader`, `pkg_config_merge` memories

### III. Test-First Development

All features MUST be developed with tests:

- **Unit Tests**: Standard Go tests (`*_test.go`) for individual components
- **Integration Tests**: Docker Compose-based tests in `integration/` with `//go:build integration` tag
- **Test Compose Files**: Located in `internal/testutil/compose/`
- **Coverage**: Aim for meaningful coverage; avoid testing trivial code paths

**Rationale**: Tests document behavior, prevent regressions, and enable confident refactoring.

📚 **Details**: See `arch_testing` memory

### IV. CLI-First Interface

All Bosun functionality MUST be accessible via CLI:

- **Cobra Framework**: Use `github.com/spf13/cobra` for command structure
- **Output Protocol**: stdout for results, stderr for errors; support JSON/YAML output for scripting
- **Flags Over Prompts**: Prefer flags for automation; interactive prompts only when essential
- **Exit Codes**: Use semantic exit codes (0=success, 1=runtime error, 2=validation error, 10-16=job errors)

**Rationale**: CLI enables automation, scripting, and integration with existing DevOps workflows.

📚 **Details**: See `pkg_cli` memory

### V. Code Quality & Simplicity

Code MUST be idiomatic Go and maintainable:

- **Go Standards**: `go fmt`, `go vet`, and golangci-lint MUST pass
- **YAGNI**: Don't build features until needed; start simple
- **Pre-commit Hooks**: Run hooks before committing to catch issues early
- **No Magic**: Prefer explicit over implicit; avoid magic strings

**Rationale**: Maintainable code enables long-term velocity and onboarding of contributors.

📚 **Details**: See `arch_code_style`, `arch_development_lifecycle` memories

### VI. Plan-Driven Execution

Job execution MUST be plan-driven:

- **Planner Generates Plan**: `JobPlanner.Plan()` produces an `ExecutionPlan` with ordered steps
- **Executor Interprets Plan**: `JobExecutor.Execute()` iterates plan steps (no hardcoded flow)
- **DryRun Matches Execute**: What `--dry-run` shows MUST match actual execution
- **Always Restart**: Stack MUST be restarted even if worker fails (availability > job success)

**Rationale**: Single source of truth for execution flow; testable planner; predictable behavior.

📚 **Details**: See `pkg_app_planner`, `pkg_app_executor` memories

## Technology Stack

| Category | Technology | Notes |
|----------|------------|-------|
| Language | Go 1.24+ | Module at `github.com/simone-viozzi/bosun` |
| CLI | Cobra | `github.com/spf13/cobra` |
| Docker | Docker SDK | `github.com/docker/docker` |
| Testing | Go test + Docker Compose | Integration tests in `integration/` |
| Build | Makefile | `make build`, `make test`, `make it`, `make lint` |
| Linting | golangci-lint | Pre-commit hook enforced |
| CI/CD | GitHub Actions | `.github/workflows/ci.yml` |

## Serena Memory Index

All implementation details live in Serena memories (`.serena/memories/`):

### Architecture
| Memory | Contents |
|--------|----------|
| `arch_overview` | Hexagonal architecture, package structure, data flow |
| `arch_testing` | Test infrastructure and patterns |
| `arch_code_style` | Go conventions and formatting |
| `arch_development_lifecycle` | Development workflow and commands |

### Packages
| Memory | Contents |
|--------|----------|
| `pkg_ports` | Interface contracts (LabelSource, JobDiscoverer, JobPlanner, JobExecutor, etc.) |
| `pkg_domain_jobs` | Job, ExecutionPlan, PlanStep types |
| `pkg_domain_labels` | Label types and constants |
| `pkg_config_schema` | Config schema types, tags, FieldSpec |
| `pkg_config_loader` | Label parsing, ValidationErrors, LoadOptions |
| `pkg_config_merge` | Multi-source config merging |
| `pkg_cli` | CLI commands, flags, exit codes |
| `pkg_app_planner` | Execution plan generation |
| `pkg_app_executor` | Plan-driven job execution |
| `pkg_adapters_dockerlabels` | Docker label discovery adapter |
| `pkg_adapters_joblabels` | Job label parsing adapter |
| `pkg_adapters_docker_compose` | Compose stack control |
| `pkg_adapters_docker_worker` | Worker container lifecycle |
| `pkg_tools_configdoc` | Auto-generated docs tooling |

### Work-in-Progress
| Memory | Contents |
|--------|----------|
| `wip_smell_milestone3` | Code smell tracking for M3 |

## Governance

This constitution supersedes all other development practices for Bosun.

- **Amendments**: Changes require documentation of rationale and impact assessment
- **Versioning**: Constitution follows semantic versioning (MAJOR.MINOR.PATCH)
- **Compliance**: All PRs MUST verify compliance with these principles
- **Exceptions**: Any deviation MUST be documented with justification in the PR description
- **Guidance**: See `.github/copilot-instructions.md` for runtime AI development guidance

**Version**: 1.2.0 | **Ratified**: 2025-11-30 | **Last Amended**: 2026-01-05
