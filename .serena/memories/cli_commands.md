# CLI Commands

Bosun uses the Cobra CLI framework for command-line interface. Commands are defined in `internal/cmd/`.

## Command Structure
- **Root Command** (`internal/cmd/root.go`): Main entry point for the CLI
- **Labels Command** (`internal/cmd/labels.go`): Group for label-related operations
- **Snapshot Command** (`internal/cmd/snapshot.go`): Captures and displays Docker label snapshots

## Available Commands

### `bosun config validate`
Validates Bosun configuration from Docker labels and files. Checks for typos, type errors, and scope mismatches.

**Usage**: `bosun config validate [--from <source>] [--scope <type>] [--print] [--stopped]`

**Flags**:
- `-f, --from <source>`: Config source: `auto` (default), `labels`, `file`
- `-s, --scope <type>`: Filter by entity type: `container`, `volume`, `network`, `global` (default: all)
- `-p, --print`: Print merged config as JSON instead of just validating
- `-c, --config <path>`: Path to config file (future)
- `--stopped`: Include stopped containers (default: false)

**Exit Codes**:
- `0`: Validation passed
- `1`: Validation failed (invalid config)
- `2`: Runtime error (Docker unavailable)

**Output Examples**:

Success:
```
Configuration valid
```

Failure:
```
Validation errors:

Container "myapp" (abc123def456):
  - unknown key: bosun.container.gracePeriod
  - invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration "30"

Found 2 error(s) in 1 entity(ies)
```

With `--print`:
```json
{
  "Instance": "",
  "StopGracePeriod": 30000000000,
  "HealthCheckInterval": 30000000000,
  "AutoRestart": true,
  "LogLevel": "info",
  "BackupEnabled": false,
  "MaxSize": 10737418240,
  "Priority": 100
}
```

**Implementation Details**:
- Uses `dockerlabels.NewFromEnv()` to create Docker client
- Uses `loader.FromLabels()` to validate labels per entity
- Uses `merge.Merge()` to combine defaults + file (future) + labels
- Filters entities by `--scope` flag
- Reports all validation errors, not just the first one

### `bosun labels snapshot`
Captures a snapshot of all Docker entities (containers, volumes, networks) with Bosun labels and prints as JSON.

**Usage**: `bosun labels snapshot [--stopped]`

**Flags**:
- `--stopped`: Include stopped containers in the snapshot (default: false)

**Output**: Pretty-printed JSON with structure:
```json
{
  "Entities": [
    {
      "Kind": "container|volume|network",
      "ID": "...",
      "Name": "...",
      "Labels": { "bosun.key": "value" },
      "Meta": { ... }
    }
  ],
  "TakenAt": "2025-10-05T16:48:42.364935862Z"
}
```

**Implementation Details**:
- Uses `dockerlabels.NewFromEnv()` to create Docker client
- Uses `ports.Selector` with `DefaultLabelPrefix` ("bosun.")
- Gracefully handles Docker unavailability with error message
- Thin wrapper around `LabelSource.Snapshot()` interface

**Error Handling**:
- Returns friendly error if Docker is unavailable: "failed to connect to Docker: ... Is Docker running?"
- Non-zero exit code on failure

## Main Entry Point
The main entry point (`cmd/bosun/main.go`) creates the root command with context and executes it with signal handling for graceful shutdown.
