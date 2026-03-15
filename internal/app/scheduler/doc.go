// Package scheduler implements cron-based job scheduling with config refresh.
//
// The Scheduler manages the lifecycle of scheduled jobs:
//   - Registers jobs with cron schedules via robfig/cron/v3
//   - Executes jobs through the JobExecutor port
//   - Tracks per-job status (idle, running, completed, failed)
//   - Periodically refreshes configuration from Docker labels
//   - Implements circuit-breaker (auto-disable after 3 consecutive failures)
//
// Concurrency is managed through three layers:
//   - Layer 1: Per-job overlap policies (queue or skip)
//   - Layer 2: Global semaphore for parallelism control
//   - Layer 3: Per-stack mutexes via StackLockManager
package scheduler
