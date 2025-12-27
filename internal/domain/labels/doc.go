// Package labels defines domain types for labeled entities discovered from
// Docker containers, volumes, and networks.
//
// Key types:
//   - LabeledEntity: Represents a Docker resource with bosun.* labels
//   - Snapshot: A point-in-time collection of labeled entities
//   - EntityKind: Enum for container, volume, or network
//
// This package is part of the domain layer and has no external dependencies.
package labels
