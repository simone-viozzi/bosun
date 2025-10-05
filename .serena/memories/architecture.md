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

## Current State
The domain and ports have been implemented for label discovery functionality. The adapters include placeholder directories for docker, http, and storage integrations, plus a dockerlabels adapter with utility functions. The main application in `internal/app/app.go` has a TODO to wire everything together.</content>
<parameter name="memory_name">architecture
