# Docker Labels Adapter

## Scope
Docker label discovery adapter in `internal/adapters/dockerlabels/`.

## What

Implements `ports.LabelSource` to discover labeled Docker entities (containers, volumes, networks).

### `DockerLabelSource`
- `NewFromEnv() (*DockerLabelSource, error)` - Create with env config
- `Snapshot(ctx, sel) (Snapshot, error)` - Discover all entity types

### Discovery Methods (private)
- `snapshotContainers(ctx, sel)` - List containers with `bosun.*` labels
- `snapshotVolumes(ctx, sel)` - List volumes with `bosun.*` labels
- `snapshotNetworks(ctx, sel)` - List networks with `bosun.*` labels
- `buildLabelFilters(sel)` - Convert Selector to Docker API filters

### Utility Functions
- `FilterByPrefixes(labels, prefixes)` - Filter label map by key prefixes
- Empty/whitespace values are dropped

### Filtering
- `ProjectFilter` → `com.docker.compose.project` label (Docker Compose)
- `StackFilter` → `bosun.stack` label
- Filters applied server-side via Docker API for efficiency

### Metadata Enrichment
- **Containers**: `image`, `compose.project`, `compose.service`, `instance`
- **Volumes**: `driver`, `instance`
- **Networks**: `driver`, `scope`, `instance`

## Why
Adapter isolates Docker SDK dependency. Uses internal `dockerClient` interface for testability.

## Related
- `pkg_ports` - LabelSource interface
- `pkg_domain_labels` - LabeledEntity, Snapshot types
- `arch_testing` - Integration tests

---

## Gotchas

### Label Keys are Case-Sensitive
Docker convention: `bosun.role` ≠ `Bosun.Role`

### Image Labels Ignored
Only runtime container labels are discovered, not Dockerfile `LABEL` instructions.

### Network Label Quirk
Docker Compose may not apply labels to networks. Workaround: create networks manually with `docker network create --label`.
