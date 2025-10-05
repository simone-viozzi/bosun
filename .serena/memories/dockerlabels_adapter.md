# Docker Labels Adapter

The `internal/adapters/dockerlabels` package provides utility functions for processing Docker label maps and implements the Docker label source adapter following hexagonal architecture.

## Current Utilities

### FilterByPrefixes
- **Function**: `FilterByPrefixes(in map[string]string, prefixes []string) map[string]string`
- **Purpose**: Filters label maps to include only keys starting with specified prefixes, while dropping empty or whitespace-only values
- **Behavior**:
  - Returns empty map if no prefixes provided (fast-path)
  - Does not mutate input map
  - Uses `strings.HasPrefix` for prefix matching
  - Uses `strings.TrimSpace` to detect empty values
- **Use Case**: Selective inclusion of labels in Docker discovery based on allowed prefixes (e.g., "bosun.*")

## Docker Label Source Implementation

### DockerLabelSource
- **Struct**: `type DockerLabelSource struct { CLI *client.Client }`
- **Constructor**: `NewFromEnv() (*DockerLabelSource, error)` - creates Docker client with environment configuration and API version negotiation
- **Interface**: Implements `ports.LabelSource` with `Snapshot()` method
- **Current State**: Container (#23), volume (#24), and network (#24) discovery fully implemented

### Container Discovery (#23 - Completed)
- **Method**: `snapshotContainers(ctx, sel)` - private method collecting containers
- **Features**:
  - Lists containers using `ContainerList` with `All: sel.IncludeStopped`
  - Filters labels by prefixes using `FilterByPrefixes`
  - Excludes containers with zero matching labels
  - Enriches entities with metadata: image, compose project/service, and instance (if `bosun.instance` label present)
  - Handles edge cases: multiple names (picks index 0), missing compose labels
- **Entity Structure**: `KindContainer`, stable ID, trimmed name, filtered labels, meta map with `image`, `compose.project`, `compose.service`, `instance`

### Volume Discovery (#24 - Completed)
- **Method**: `snapshotVolumes(ctx, sel)` - private method collecting volumes
- **Features**:
  - Lists volumes using `VolumeList` with `volume.ListOptions{}`
  - Filters labels by prefixes using `FilterByPrefixes`
  - Excludes volumes with zero matching labels
  - Enriches entities with driver metadata and instance (if `bosun.instance` label present)
  - Creates entities with `KindVolume`, `ID=Name`, `Name=Name`, filtered labels
- **Entity Structure**: `KindVolume`, ID equals Name, filtered labels, meta map with `driver`, `instance`

### Network Discovery (#24 - Completed)
- **Method**: `snapshotNetworks(ctx, sel)` - private method collecting networks
- **Features**:
  - Lists networks using `NetworkList` with `network.ListOptions{}`
  - Filters labels by prefixes using `FilterByPrefixes`
  - Excludes networks with zero matching labels
  - Enriches entities with driver and scope metadata and instance (if `bosun.instance` label present)
  - Creates entities with `KindNetwork`, `ID=n.ID`, `Name=n.Name`, filtered labels
- **Entity Structure**: `KindNetwork`, Docker network ID, network name, filtered labels, meta map with `driver`, `scope`, `instance`

### Snapshot Aggregation
- **Method**: `Snapshot(ctx, sel)` - public method implementing `LabelSource` interface
- **Features**:
  - Calls all three discovery methods (containers, volumes, networks)
  - Concatenates results using `slices.Concat()` for clarity
  - Returns unified `Snapshot` with all entity types and timestamp

## Architecture Notes
- Pure utility functions with no external dependencies
- Docker client adapter follows hexagonal architecture as adapter layer
- Designed to be used by concrete label source implementations
- Constructor creates client but defers actual daemon connection until API calls
- Uses `dlabels.LabelInstance` constant for consistent instance label access
- Replaced deprecated `types.Container` with `container.Summary` for future compatibility

## Testing
- Comprehensive unit tests for `FilterByPrefixes` with table-driven approach
- Constructor test skipped per project decision (external dependency concerns)
- Integration tests implemented for full snapshot functionality with Docker Compose stacks
- Tests validate volume/network discovery alongside containers
- Recent addition: validation of Meta enrichment (image for containers, driver for volumes, driver/scope for networks)</content>
<parameter name="memory_name">dockerlabels_adapter
