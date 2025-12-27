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
└── plan
    ├── list        # List discovered jobs
    └── show <job>  # Show execution plan for a job
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

### Main Entry Point
`cmd/bosun/main.go` - Context with signal handling for graceful shutdown.

## Why
Standardized exit codes enable scripting. JSON output enables automation.

## Related
- `pkg_adapters_dockerlabels` - Used by all commands
- `pkg_adapters_joblabels` - Used by plan commands
- `pkg_app_planner` - Used by plan show
