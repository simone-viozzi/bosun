// Package planner provides the pure-function job planner implementation.
//
// The planner transforms discovered jobs into execution plans, generating
// ordered steps for stopping containers and running worker containers.
//
// Key characteristics:
//   - Deterministic: same inputs always produce identical outputs
//   - Pure: no Docker API calls or side effects
//   - Validates Compose dependency graph to prevent orphaned dependents
//
// This package implements the ports.JobPlanner interface.
package planner
