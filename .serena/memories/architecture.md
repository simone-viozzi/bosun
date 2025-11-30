# Architecture Overview

Bosun follows the **Hexagonal Architecture** (also known as Ports and Adapters) pattern, which promotes separation of concerns and testability.

## Core Principles
- **Domain Independence**: Business logic doesn't depend on external systems
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Testability**: Easy to mock external dependencies

## Structure
- **Domain** (`internal/domain/`): Contains the core business logic and entities. This is the heart of the application and should not depend on external frameworks or libraries.
- **Ports** (`internal/ports/`): Defines interfaces (contracts) for external interactions. These are the \"ports\" that the domain uses to communicate with the outside world.
- **Adapters** (`internal/adapters/`): Concrete implementations of the ports. These handle external system integrations like Docker, HTTP clients, storage systems, etc.
- **Application** (`internal/app/`): Orchestrates the domain logic and wires together ports and adapters.
- **Configuration** (`internal/config/`): Handles application configuration and startup wiring.

## Benefits
- Easy to test: Mock adapters for unit testing
- Technology agnostic: Can swap implementations (e.g., different storage backends)
- Maintainable: Clear separation of concerns
- Framework independent: Domain logic doesn't depend on external libraries

## Implementation Notes
- **Domain**: `internal/domain/labels/` (label types) and `internal/domain/jobs/` (Job, ExecutionPlan, PlanStep)
- **Ports**: `internal/ports/labels.go` (LabelSource, Selector) and `internal/ports/planner.go` (JobDiscoverer, JobPlanner, ValidationError)
- **Config System**: Schema, loader, and merger packages under `internal/config/`
- **CLI**: Commands in `internal/cmd/` - config, labels, plan subcommands
- **Adapters**: `dockerlabels` (label discovery), `joblabels` (job discovery from labels)
- **Application**: `internal/app/app.go` (basic App struct), `internal/app/planner/` (JobPlanner with topological sorting)
