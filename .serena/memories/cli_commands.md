# CLI Commands

Bosun uses the Cobra CLI framework for command-line interface. Commands are defined in `internal/cmd/`.

## Command Structure
- **Root Command** (`internal/cmd/root.go`): Main entry point for the CLI
- **Labels Command** (`internal/cmd/labels.go`): Group for label-related operations
- **Snapshot Command** (`internal/cmd/snapshot.go`): Captures and displays Docker label snapshots

## Available Commands

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
