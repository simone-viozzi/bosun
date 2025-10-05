# Label Discovery Domain and Ports

## Overview
Issue #20 implemented the foundational domain models and ports for label discovery in Bosun. This follows hexagonal architecture principles, providing clean interfaces for querying Docker labels (and potentially other sources) while keeping domain logic decoupled from external systems.

## Domain Models (`internal/domain/labels`)
Core business entities for representing labeled entities:

### Types
- **`Kind`**: String type with constants:
  - `KindContainer = "container"`
  - `KindVolume = "volume"`
  - `KindNetwork = "network"`

- **`LabeledEntity`**: Represents a single labeled Docker entity
  - `Kind Kind`: Type of entity (container/volume/network)
  - `ID string`: Unique identifier
  - `Name string`: Human-readable name
  - `Labels map[string]string`: Docker labels (case-sensitive keys)
  - `Meta map[string]string`: Additional metadata (e.g., compose.project, compose.service, image, networks)

- **`Snapshot`**: Point-in-time collection of labeled entities
  - `Entities []LabeledEntity`: List of entities
  - `TakenAt time.Time`: Timestamp when snapshot was taken

## Ports (`internal/ports/labels`)
Clean interfaces for external label sources:

### Selector
Query parameters for filtering label snapshots:
- `Prefixes []string`: Label key prefixes to filter by (e.g., ["bosun."])
- `IncludeStopped bool`: Whether to include stopped containers
- `ProjectFilter []string`: Optional Docker Compose project names to filter by

### LabelSource Interface
Contract for label discovery implementations:
```go
type LabelSource interface {
    Snapshot(ctx context.Context, sel Selector) (dlabels.Snapshot, error)
}
```

## Usage Notes
- Labels keys are **case-sensitive** (following Docker convention)
- Domain types are vendor-agnostic (no Docker SDK dependencies)
- Ports enable easy testing with mocks and swapping implementations
- Future adapters (e.g., Docker, Kubernetes) will implement `LabelSource`

## Testing
Unit tests verify compilation and basic functionality. Integration with actual label sources will be tested in adapter implementations.</content>
<parameter name="memory_name">label_discovery_domain
