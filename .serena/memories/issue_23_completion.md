# Issue #23 Implementation - Container Discovery

## Summary
Issue #23 "Container Discovery (running by default) for Label Discovery" has been fully implemented and committed.

## What Was Implemented
- **Container Collection**: `DockerLabelSource.snapshotContainers()` method that lists Docker containers
- **Filtering Logic**: Uses `FilterByPrefixes()` to filter labels by specified prefixes
- **Entity Construction**: Creates `LabeledEntity` with KindContainer, stable ID, trimmed name, filtered labels, and metadata
- **Metadata Enrichment**: Includes compose.project, compose.service, and image information
- **Edge Case Handling**: Handles multiple container names (picks primary), missing compose labels (empty strings)
- **Running by Default**: Uses `All: sel.IncludeStopped` to include stopped containers only when requested

## Files Modified
- `internal/adapters/dockerlabels/source.go`: Added container discovery logic

## Testing Status
- **Unit Tests**: Existing tests pass, no new unit tests added (external dependency)
- **Integration Tests**: Deferred to Issue #26 (full snapshot functionality)
- **Code Quality**: Passes `make fmt`, `make vet`, pre-commit hooks

## Architecture Compliance
- Follows hexagonal architecture (adapter implements ports.LabelSource)
- Uses domain types (LabeledEntity, Snapshot)
- Reuses existing utilities (FilterByPrefixes)
- Clean separation of concerns

## Next Steps
- Issue #24: Volume discovery
- Issue #25: Network discovery + aggregation
- Issue #26: Integration tests

## Commit Details
- **Hash**: 17469c5
- **Message**: feat(adapters/dockerlabels): implement container discovery for label discovery
- **Closes**: #23
