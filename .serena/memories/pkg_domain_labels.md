# Labels Domain Package

## Scope
Domain types for Docker label discovery in `internal/domain/labels/`.

## What

### Types

**`Kind`** - Entity type enum
- `KindContainer = "container"`
- `KindVolume = "volume"`
- `KindNetwork = "network"`

**`LabeledEntity`** - Single Docker entity with labels
- `Kind Kind` - Entity type
- `ID string` - Unique identifier
- `Name string` - Human-readable name
- `Labels map[string]string` - Filtered labels (case-sensitive keys)
- `Meta map[string]string` - Adapter-enriched metadata:
  - Containers: `image`, `compose.project`, `compose.service`, `instance`
  - Volumes: `driver`, `instance`
  - Networks: `driver`, `scope`, `instance`

**`Snapshot`** - Point-in-time collection
- `Entities []LabeledEntity` - Discovered entities
- `TakenAt time.Time` - Snapshot timestamp

### Constants
- `DefaultLabelPrefix = "bosun."` - Standard prefix
- `LabelInstance = "bosun.instance"` - Instance identifier

## Why
Domain types are Docker-SDK-agnostic. Adapters enrich with metadata without coupling domain to Docker types.

## Related
- `pkg_adapters_dockerlabels` - Implements discovery
- `pkg_ports` - LabelSource interface
