# Bosun

A Docker orchestration tool following hexagonal/clean architecture principles.

## Overview

Bosun is an early-stage project designed to provide Docker container orchestration capabilities with a focus on clean architecture and maintainability.

### Architecture

Bosun follows **hexagonal/clean architecture** principles:
- `internal/domain/` - Business logic entities and rules
- `internal/ports/` - Interface definitions for external dependencies
- `internal/adapters/` - Concrete implementations (docker, http, storage)
- `internal/app/` - Application wiring and orchestration
- `cmd/bosun/main.go` - Application entry point

## Getting Started

### Prerequisites

- Go 1.25 or later
- Docker (required for integration tests)
- Make

### Building

```bash
make build    # Builds to bin/bosun
```

### Running

```bash
make run      # Runs from source
```

### Usage

Bosun provides a CLI for inspecting Docker labels:

```bash
# View all Docker entities with bosun.* labels
bosun labels snapshot

# Include stopped containers in the snapshot
bosun labels snapshot --stopped
```

The snapshot command outputs pretty-printed JSON showing containers, volumes, and networks with their Bosun labels.

## Testing

Bosun includes comprehensive unit and integration tests. See [Testing Guide](docs/testing.md) for detailed instructions.

**Quick start:**
```bash
make test     # Run unit tests
make itv      # Run integration tests (verbose)
```

**Prerequisites for integration tests:**
- Docker must be installed and running
- Integration tests use the `integration` build tag

For detailed testing instructions, troubleshooting, and test-writing guidelines, see [docs/testing.md](docs/testing.md).

## Development

### Code Quality

```bash
make fmt      # Format code
make vet      # Run go vet
make tidy     # Tidy dependencies
```

### Project Conventions

- **Commit Messages**: Follow [Conventional Commits](https://www.conventionalcommits.org/)
- **Branching**: Semi-linear history enforced - rebase PRs, no merge commits
- **Testing**: Integration tests use build tag `//go:build integration`

## Documentation

- [Testing Guide](docs/testing.md) - How to run and write tests

## License

See [LICENSE](LICENSE) file for details.
