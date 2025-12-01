# CLI Contract: Milestone 2.5 Updates

**Date**: 2025-11-30

## Exit Codes

All commands MUST use these exit codes:

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Command completed successfully |
| 1 | Runtime Error | Docker unavailable, I/O failure, network error |
| 2 | Validation Error | Invalid labels, missing required fields, config errors |

## Command Updates

### `bosun` (root)

**Before**:
```
Bosun - Docker label management tool
```

**After**:
```
Bosun - Docker-label-driven backup job orchestrator

Bosun discovers backup jobs from Docker labels, generates execution plans,
and orchestrates safe backup workflows (stop containers, run worker, restart).
```

---

### `bosun plan list`

**New Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project` | string | "" | Filter by Docker Compose project name |
| `--stack` | string | "" | Filter by bosun.stack label value |

**Behavior**:
- No filter: Show all jobs (backward compatible)
- `--project X`: Show only jobs from containers with `com.docker.compose.project=X`
- `--stack Y`: Show only jobs from containers with `bosun.stack=Y`
- Both flags: Show jobs matching both (AND)
- No matches: Exit 0, empty JSON array, stderr "No jobs found"

**Example**:
```bash
# All jobs
bosun plan list

# Only jobs from compose project "myapp"
bosun plan list --project myapp

# Only jobs with stack "production"
bosun plan list --stack production

# Jobs in myapp project AND production stack
bosun plan list --project myapp --stack production
```

---

### `bosun plan show <job-name>`

**New Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project` | string | "" | Limit scope to Docker Compose project |
| `--stack` | string | "" | Limit scope to bosun.stack label value |

**Behavior**:
- Filters apply to container discovery for the job
- If job not found (after filtering): Exit 1, stderr suggests `bosun plan list`

**Example**:
```bash
# Show plan for job in specific project
bosun plan show daily-backup --project myapp
```

---

### `bosun labels snapshot`

**New Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project` | string | "" | Filter by Docker Compose project name |
| `--stack` | string | "" | Filter by bosun.stack label value |

**Behavior**:
- Same filtering semantics as `plan list`
- Empty result: Exit 0, empty JSON object

---

### `bosun config validate`

**No Changes** to flags.

**Exit Code Update**:
- Validation errors now return exit code 2 (was 1)

---

## Error Message Format

All error messages MUST follow this format:

```
Error: <context>: <message>

<optional: suggestion for next steps>
```

**Examples**:

```
Error: plan show: job "nonexistent" not found

Run 'bosun plan list' to see available jobs.
```

```
Error: Docker connection failed: Cannot connect to the Docker daemon

Is Docker running? Check with 'docker info'.
```

```
Error: validation: container "app" has bosun.job.enabled=true but missing bosun.job.name

Add a 'bosun.job.name' label to the container.
```
