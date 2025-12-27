# Job Labels Adapter

## Scope
Job discovery from Docker labels in `internal/adapters/joblabels/`.

## What

Implements `ports.JobDiscoverer` to transform labeled entities into `jobs.Job` objects.

### `Discoverer`
- `NewDiscoverer() *Discoverer` - Constructor with cron parser
- `DiscoverJobs(ctx, snapshot) ([]Job, []ValidationError, error)` - Main method

### Label Constants

Label constants are defined in `internal/config/schema/job_labels.go`:

**Container Labels** (via `schema.LabelJob*`):
- `bosun.job.enabled` - Boolean, enables job participation
- `bosun.job.name` - Unique job identifier
- `bosun.job.schedule` - Cron expression
- `bosun.job.worker.image` - Worker container image

**Volume Labels**:
- `bosun.job.attach` - Job name to attach volume to
- `bosun.job.mount.path` - Mount path in worker (default: `/mnt/<volume>`)
- `bosun.job.mount.mode` - `ro` or `rw` (default: `ro`)

**Stack Resolution** (in discoverer.go):
- `bosun.stack` - Explicit stack name (highest priority)
- `com.docker.compose.project` - Docker Compose project

### Discovery Algorithm

1. **Collect job definitions**: Scan containers with `bosun.job.enabled=true`
2. **Merge fields**: Multiple containers can define same job, detect conflicts
3. **Attach volumes**: Match volumes by `bosun.job.attach` value
4. **Validate**: Cron expressions, mount modes, orphaned references
5. **Build Jobs**: Apply defaults, construct final objects

### Validation Errors (non-fatal)
- `JobErrorMissingName` - Enabled but no name
- `JobErrorInvalidSchedule` - Bad cron expression
- `JobErrorConflictingField` - Containers disagree on field value
- `JobErrorOrphanedVolume` - Volume references non-existent job
- `JobErrorInvalidMountPath/Mode` - Bad mount config

## Why
Separates label parsing complexity from domain. Returns all errors, not just first.

## Related
- `pkg_ports` - JobDiscoverer interface
- `pkg_domain_jobs` - Job type
- `pkg_adapters_dockerlabels` - Provides Snapshot input
