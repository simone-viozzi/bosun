# Bosun Configuration Reference

> Auto-generated from ConfigV1 schema. Do not edit manually.

## Global Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.instance` | string | - | No | - | Unique instance identifier |

## Container Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.container.autoRestart` | boolean | true | No | - | Automatically restart containers |
| `bosun.container.healthCheckInterval` | duration | 30s | No | - | Interval between health checks |
| `bosun.container.logLevel` | enum | info | No | debug \| info \| warn \| error | Logging verbosity level |
| `bosun.container.stopGracePeriod` | duration | 30s | No | - | Grace period before force stopping |

## Volume Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.volume.enabled` | boolean | false | No | - | Enable volume processing |
| `bosun.volume.maxSize` | byte size | 10GB | No | - | Maximum volume size |

## Network Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.network.priority` | integer | 100 | No | - | Network priority (lower = higher priority) |

## Job Configuration (M4)

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.job.enabled` | boolean | false | No | true \| false | Enable job scheduling on this container |
| `bosun.job.name` | string | - | Yes (when enabled) | - | Unique job identifier |
| `bosun.job.schedule` | cron | `0 0 * * *` | No | 5-field cron | Cron schedule (minute hour dom month dow) |
| `bosun.job.overlap-policy` | enum | queue | No | queue \| skip | Behavior when a job fires while previous run is active |
| `bosun.job.worker.image` | string | `busybox:latest` | No | - | Docker image used to run the backup worker |
| `bosun.job.attach` | string | - | No | - | Attach this volume to the named job (volume scope) |
| `bosun.job.mount-path` | string | `/mnt/<volume>` | No | - | Mount path inside the worker container |
| `bosun.job.mount-mode` | enum | ro | No | ro \| rw | Mount mode for attached volumes |
| `bosun.job.stop-timeout` | duration | 30s | No | - | Timeout for stopping containers before backup |
| `bosun.job.start-timeout` | duration | 30s | No | - | Timeout for restarting containers after backup |
| `bosun.job.timeout` | duration | 1h | No | - | Overall job execution timeout |

### Overlap Policies

- **queue** (default): If a job fires while the previous run is still active, the new run waits in a queue until the previous run completes.
- **skip**: If a job fires while the previous run is still active, the new run is skipped and a `JobSkipped` event is emitted.

## Daemon Command

The `bosun daemon` command starts the scheduling engine that monitors Docker labels and runs backup jobs on their configured schedules.

```
bosun daemon [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--parallelism` | int | 1 | Maximum number of jobs running concurrently |
| `--refresh-interval` | duration | 5m | How often to re-scan Docker labels for configuration changes |

### Features

- **Automatic discovery**: Jobs are discovered from Docker container labels on startup and on each refresh interval.
- **Graceful shutdown**: On SIGINT/SIGTERM, the daemon stops scheduling new jobs and waits for running jobs to complete (60s timeout). A second signal forces immediate exit.
- **Config refresh**: Periodically re-scans Docker labels to detect added, removed, or changed jobs.
- **Circuit breaker**: Jobs that fail 3 consecutive times are auto-disabled until manually re-enabled.
- **Per-stack serialization**: Jobs targeting the same Docker Compose stack are serialized to avoid conflicts.
- **Global concurrency control**: The `--parallelism` flag limits the total number of concurrently executing jobs.

## Value Formats

### Duration

Go duration syntax. Supported units: `ns`, `us`/`µs`, `ms`, `s`, `m`, `h`.

**Examples:** `30s`, `5m`, `1h30m`, `100ms`

### Byte Size

Docker/go-units byte size syntax. Supports base-10 (KB, MB, GB, TB) and base-2 (KiB, MiB, GiB, TiB) units.

**Examples:** `100MB`, `1GiB`, `500KB`, `10GB`

### List

List values can be specified as CSV (comma-separated) or JSON array syntax.

**Examples:** `value1,value2,value3`, `["value1", "value2"]`
