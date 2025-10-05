# Docker Labels Adapter Documentation

## Documentation Location
Comprehensive documentation for the `dockerlabels` adapter is available in:
- **Primary**: `internal/adapters/dockerlabels/README.md` (297 lines, 8.1KB)
- **Quick Reference**: `.github/copilot-instructions.md` (Project State > Label Discovery Module section)

## README Contents
The README provides complete documentation including:

### Overview & Features
- Entity discovery for containers, volumes, and networks
- Prefix-based filtering (default: `bosun.*`)
- Metadata enrichment
- Smart exclusions (zero-label entities, stopped containers)

### Scope (v1)
- **Supported**: Containers, volumes, networks
- **Excluded**: Image labels (intentional), project filtering (field exists but unused in v1)

### Design Decisions
- Prefix-based filtering using `FilterByPrefixes`
- Metadata enrichment per entity type (containers: image, compose info; volumes: driver; networks: driver, scope)
- Stopped container exclusion by default

### Examples
1. **Docker Compose File**: Complete example with labeled services, volumes, and networks
2. **Go Code Snippet**: Working example using `NewFromEnv()`, creating selector, and calling `Snapshot()`
3. **CLI Usage**: Command-line examples with `bosun labels snapshot`
4. **JSON Output**: Sample snapshot output showing entity structure

### Gotchas / Pitfalls (Critical)
1. **Case Sensitivity**: Label keys are case-sensitive (`bosun.role` ≠ `Bosun.Role`)
2. **Image Labels Ignored**: Only container-level labels are discovered, not image LABEL directives
3. **Network Label Application**: Docker Compose may not apply network labels; manual creation may be required
4. **Empty Label Values**: Automatically filtered out

### Architecture
- File structure overview
- Key components (DockerLabelSource, NewFromEnv, Snapshot, FilterByPrefixes)
- Testing infrastructure (unit tests, integration tests, test compose files)

### Constants Reference
- `dlabels.DefaultLabelPrefix = "bosun."`
- `dlabels.LabelInstance = "bosun.instance"`

## Copilot Instructions Update
`.github/copilot-instructions.md` now includes:
- Reference to the README for comprehensive documentation
- Key conventions (case sensitivity, image labels ignored, network manual labeling)
- Links to Serena memories for implementation details

## Usage for Onboarding
The README is designed for fast onboarding:
1. Developers can start with the Overview section
2. Jump to Examples for quick-start code
3. Review Gotchas to avoid common pitfalls
4. Reference Architecture for deeper understanding

## Validation
- All code examples verified against actual implementation
- Compose examples consistent with test fixtures
- Constants match domain definitions
- CLI examples match actual command structure
