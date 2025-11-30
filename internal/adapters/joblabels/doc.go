// Package joblabels provides an adapter for discovering jobs from Docker label snapshots.
//
// This adapter implements the ports.JobDiscoverer interface, extracting job definitions
// from containers with bosun.job.* labels and volumes with bosun.job.attach labels.
//
// The discoverer:
//   - Filters containers with bosun.job.enabled=true
//   - Merges multiple containers into a single job by bosun.job.name
//   - Validates cron expressions using robfig/cron/v3
//   - Discovers volume attachments from bosun.job.attach labels
//   - Resolves stack names from bosun.stack or com.docker.compose.project
package joblabels
