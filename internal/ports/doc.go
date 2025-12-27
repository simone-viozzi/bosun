// Package ports defines the interfaces (contracts) for external interactions
// in the hexagonal architecture.
//
// Key interfaces:
//   - LabelSource: Discovers labeled entities from external sources (e.g., Docker)
//   - Selector: Criteria for filtering discovered entities
//   - JobDiscoverer: Converts labeled entities to backup job definitions
//   - JobPlanner: Plans execution order for backup jobs
//
// Adapters in internal/adapters implement these interfaces.
package ports
