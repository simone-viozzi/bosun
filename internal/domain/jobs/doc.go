// Package jobs defines the domain entities for job discovery and execution planning.
//
// This package contains pure Go types with no external dependencies, following
// hexagonal architecture principles. Jobs are discovered from Docker container
// labels and transformed into execution plans.
//
// Key types:
//   - Job: A discovered job definition assembled from container labels
//   - ExecutionPlan: An ordered sequence of steps to execute a job
//   - PlanStep: An individual action in an execution plan
//   - VolumeAttachment: A volume to be mounted in the worker container
//   - Stack: A logical grouping of Docker entities
package jobs
