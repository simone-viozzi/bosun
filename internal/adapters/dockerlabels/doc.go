// Package dockerlabels provides an adapter for discovering Docker labels from
// containers, volumes, and networks via the Docker API.
//
// It implements the ports.LabelSource interface, enabling the domain layer to
// discover labeled entities without depending on Docker-specific code.
//
// Key components:
//   - DockerLabelSource: Main adapter implementing LabelSource.Snapshot()
//   - FilterByPrefixes: Utility to filter label maps by key prefixes
//
// Usage:
//
//	source, err := dockerlabels.NewFromEnv()
//	if err != nil { ... }
//	snapshot, err := source.Snapshot(ctx, sel)
package dockerlabels
