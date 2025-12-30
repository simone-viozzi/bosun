# CLI Commands

## Scope
Command-line interface in `internal/cmd/` using Cobra.

## What

### Command Structure
```
bosun
├── config
│   └── validate    # Validate config labels
├── labels
│   └── snapshot    # Show Docker entities with bosun.* labels
├── plan
│   ├── list        # List discovered jobs
│   └── show <job>  # Show execution plan for a job
└── job
    └── run <job>   # Execute a backup job
```

### `bosun config validate`
Validates configuration from Docker labels.

**Flags**:
- `-f, --from` - Source: `auto`, `labels`, `file`
- `-s, --scope` - Validate only: `container`, `volume`, `network`, `global`
- `-p, --print` - Print merged config as JSON
- `-c, --config` - Path to config file
- `--stopped` - Include stopped containers

**Exit Codes** (`internal/cmd/exitcodes.go`):
- `0` - Success
- `1` - Runtime error (Docker unavailable)
- `2` - Validation error

### `bosun labels snapshot`
Captures snapshot of all Docker entities with `bosun.*` labels.

**Flags**:
- `--stopped` - Include stopped containers

**Output**: JSON with `Entities[]` and `TakenAt`

### `bosun plan list`
Lists jobs discovered from labels.

**Flags**:
- `-f, --format` - `text`, `json`, `yaml`
- `--stopped`, `--stack`, `--project` - Filters

### `bosun plan show <job>`
Shows execution plan for a specific job.

**Flags**: Same as `plan list`

**Output**: Ordered steps (stop, run_worker, start)

### `bosun job run <job>`
Executes a backup job.

**Flags**:
- `--dry-run` - Validate and show plan without executing
- `--timeout` - Worker execution timeout (default: 1h)
- `--stop-timeout` - Stack stop timeout (default: 30s)
- `--start-timeout` - Stack start timeout (default: 30s)
- `--keep-worker` - Keep worker container after execution
- `--project` - Filter by Compose project name

**Exit Codes** (M3 additions, 10-16 range):
- `10` - Worker failed (non-zero exit)
- `11` - Stop failed
- `12` - Start failed
- `13` - Timeout
- `14` - Image not found
- `15` - Job not found
- `16` - Interrupted

**Behavior**:
- Pre-validates worker image before stopping stack
- Always restarts stack, even if worker fails
- Captures worker logs to stdout
- Graceful shutdown on SIGINT/SIGTERM

### Main Entry Point
`cmd/bosun/main.go` - Context with signal handling for graceful shutdown.

## Why
Standardized exit codes enable scripting. JSON output enables automation.

## Related
- `pkg_adapters_dockerlabels` - Used by all commands
- `pkg_adapters_joblabels` - Used by plan commands
- `pkg_app_planner` - Used by plan show
