// Package compose implements Compose stack control via Docker API.
//
// This package provides an adapter that implements the ports.ComposeController
// interface using the Docker SDK for Go. It manages Compose stack lifecycle
// operations: discovering containers, stopping stacks, and starting stacks.
//
// # Stack Discovery
//
// Containers are identified as belonging to a Compose stack using labels:
//   - com.docker.compose.project: Stack name
//   - com.docker.compose.service: Service name within stack
//   - com.docker.compose.depends_on: Comma-separated service dependencies
//
// # Dependency Ordering
//
// Container dependencies are discovered from com.docker.compose.depends_on labels
// and sorted topologically using depth-first search (DFS). Stop order is the
// reverse of start order:
//   - Stop: Dependents first (reverse topological)
//   - Start: Dependencies first (forward topological)
//
// # M3 Limitations
//
// This implementation does NOT support:
//   - Health check waiting (containers are "started" but may not be healthy)
//   - Condition-based dependencies (depends_on: condition: service_healthy)
//   - Orphan container cleanup
//   - Volume management
//
// These limitations are acceptable for M3 MVP and documented in user-facing docs.
//
// # References
//
//   - GitHub Issue: #118
//   - Port Interface: internal/ports/compose.go (#115)
//   - Research: .serena/memories/m3_compose_control_decision.md
package compose
