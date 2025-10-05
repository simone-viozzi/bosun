# Label Discovery

The label discovery system in Bosun enables querying and managing Docker entities (containers, volumes, networks) through labels. This document covers the `dockerlabels` adapter implementation.

## Overview

The `dockerlabels` adapter (located in `internal/adapters/dockerlabels/`) provides label discovery for Docker entities following hexagonal architecture principles. It implements the `ports.LabelSource` interface to enable querying Docker resources by label prefixes.

This adapter discovers and filters Docker entities based on label prefixes, enriching them with relevant metadata. It is designed for the Bosun orchestration system to identify and manage labeled Docker resources.

**Key Features:**
- **Entity Discovery**: Containers, volumes, and networks
- **Prefix Filtering**: Filter labels by prefix (default: `bosun.*`)
- **Metadata Enrichment**: Adds contextual information to entities
- **Smart Exclusions**: Drops entities with zero matching labels; excludes stopped containers by default

## Scope (v1)

### Supported Entities
- **Containers**: Running and optionally stopped containers
- **Volumes**: All volumes with matching labels
- **Networks**: All networks with matching labels

### Label Filtering
- Only labels matching specified prefixes are included (e.g., `bosun.*`)
- Empty or whitespace-only label values are excluded
- **Case-sensitive**: Label keys are case-sensitive following Docker conventions
- Entities with zero matching labels are dropped from results

### Intentional Exclusions (v1)
- **Image labels**: Labels from Docker images are **not** included in discovery
  - Only labels directly applied to containers are discovered
  - This is an intentional design decision for v1 to keep scope focused
- **Project filtering**: The `Selector.ProjectFilter` field exists but is not used in v1

## Design Decisions

### Prefix-Based Filtering
Labels are filtered using the `FilterByPrefixes` utility function:
- Inclusion-only: only labels matching prefixes are kept
- No mutation of input maps
- Efficient early-exit when no prefixes provided

### Metadata Enrichment
Each entity type is enriched with relevant metadata in the `Meta` map:

| Entity Type | Metadata Fields |
|-------------|----------------|
| **Container** | `image`, `compose.project`, `compose.service`, `instance` (if `bosun.instance` label present) |
| **Volume** | `driver`, `instance` (if `bosun.instance` label present) |
| **Network** | `driver`, `scope`, `instance` (if `bosun.instance` label present) |

### Stopped Containers
By default, stopped containers are excluded. Use `Selector.IncludeStopped = true` to include them.

## Example Usage

### Docker Compose File with Labels

Create a `docker-compose.yaml` with Bosun labels:

```yaml
services:
  web:
    image: nginx:alpine
    labels:
      bosun.role: "webserver"
      bosun.env: "production"
    volumes:
      - app-data:/data
    networks:
      - app-net

volumes:
  app-data:
    labels:
      bosun.purpose: "storage"
      bosun.backup: "daily"

networks:
  app-net:
    labels:
      bosun.type: "internal"
      bosun.zone: "private"
```

**Note on Networks**: Docker Compose may not always apply network labels correctly. You may need to manually create networks with labels:

```bash
docker network create \
  --label bosun.type=internal \
  --label bosun.zone=private \
  app-net
```

### Go Code Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Docker label source
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		panic(fmt.Errorf("failed to create Docker client: %w", err))
	}

	// Create selector with default prefix
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix}, // "bosun."
		IncludeStopped: false,
	}

	// Get snapshot
	snapshot, err := source.Snapshot(ctx, selector)
	if err != nil {
		panic(fmt.Errorf("failed to get snapshot: %w", err))
	}

	// Process entities
	fmt.Printf("Found %d entities at %s\n", len(snapshot.Entities), snapshot.TakenAt)
	for _, entity := range snapshot.Entities {
		fmt.Printf("  [%s] %s (ID: %s)\n", entity.Kind, entity.Name, entity.ID)
		fmt.Printf("    Labels: %v\n", entity.Labels)
		fmt.Printf("    Meta: %v\n", entity.Meta)
	}
}
```

### CLI Usage

Bosun provides a CLI command for quick snapshots:

```bash
# Get snapshot of all entities with bosun.* labels
bosun labels snapshot

# Include stopped containers
bosun labels snapshot --stopped

# Pretty-printed JSON output with all entity details
```

Example output:
```json
{
  "Entities": [
    {
      "Kind": "container",
      "ID": "abc123...",
      "Name": "myapp-web-1",
      "Labels": {
        "bosun.role": "webserver",
        "bosun.env": "production"
      },
      "Meta": {
        "image": "nginx:alpine",
        "compose.project": "myapp",
        "compose.service": "web"
      }
    },
    {
      "Kind": "volume",
      "ID": "myapp_app-data",
      "Name": "myapp_app-data",
      "Labels": {
        "bosun.purpose": "storage",
        "bosun.backup": "daily"
      },
      "Meta": {
        "driver": "local"
      }
    }
  ],
  "TakenAt": "2025-01-16T10:30:00Z"
}
```

## Gotchas / Pitfalls

### Case Sensitivity
Label keys are **case-sensitive**. `bosun.role` and `Bosun.Role` are different labels. Docker follows this convention strictly.

```yaml
# ✓ Correct - will be discovered
labels:
  bosun.role: "webserver"

# ✗ Wrong - will NOT match "bosun." prefix
labels:
  Bosun.role: "webserver"
```

### Image Labels Are Ignored (v1)
Only labels directly applied to containers are discovered. Labels from Docker images are **not** included:

```dockerfile
# These labels in the Dockerfile/image are IGNORED
LABEL bosun.app=myapp
LABEL bosun.version=1.0
```

To label containers, add labels in your Docker Compose file or `docker run` command:

```yaml
services:
  app:
    image: myimage:latest
    labels:  # ✓ These labels are discovered
      bosun.app: "myapp"
      bosun.version: "1.0"
```

### Network Label Application
Docker Compose may not properly apply labels to networks in all versions. If network labels are not discovered:

1. **Option 1**: Create the network manually before starting compose:
   ```bash
   docker network create --label bosun.type=internal myapp-net
   ```

2. **Option 2**: Create the network programmatically:
   ```go
   import "github.com/docker/docker/api/types/network"

   cli.NetworkCreate(ctx, "myapp-net", network.CreateOptions{
       Labels: map[string]string{
           "bosun.type": "internal",
       },
   })
   ```

### Empty Label Values
Labels with empty or whitespace-only values are automatically filtered out:

```yaml
# These will be excluded from results
labels:
  bosun.empty: ""
  bosun.whitespace: "   "
  bosun.valid: "ok"  # ✓ Only this will be included
```

## Architecture

The adapter follows hexagonal architecture principles:

```
internal/adapters/dockerlabels/
├── filters.go         # FilterByPrefixes utility
├── filters_test.go    # Unit tests for filtering
├── source.go          # DockerLabelSource implementation
└── source_test.go     # Unit tests for source
```

**Key Components:**
- `DockerLabelSource`: Implements `ports.LabelSource` interface
- `NewFromEnv()`: Constructor using Docker environment variables
- `Snapshot()`: Main discovery method returning all entities
- `FilterByPrefixes()`: Pure utility function for label filtering

**Testing:**
- Unit tests for filtering logic in `internal/adapters/dockerlabels/`
- Integration tests in `integration/dockerlabels_test.go`
- Test compose files in `internal/testutil/compose/`

## Constants

```go
import dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"

dlabels.DefaultLabelPrefix  // "bosun."
dlabels.LabelInstance       // "bosun.instance"
```

Use these constants instead of hardcoding strings.

## Future Enhancements

Potential additions for future versions:
- Image label discovery (requires design for label inheritance)
- Project filtering implementation (using `Selector.ProjectFilter`)
- Kubernetes label source adapter
- Custom metadata extractors
- Incremental updates vs. full snapshots
