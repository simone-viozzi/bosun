# Quickstart: Config Validate Command

**Date**: 2025-11-30
**Feature**: 003-cli-config-validate

## Overview

The `bosun config validate` command validates your Docker label configuration against the Bosun schema. It catches typos, type errors, and scope mismatches before they cause runtime issues.

## Prerequisites

- Docker running
- Containers/volumes/networks with `bosun.*` labels

## Basic Usage

### Validate All Configuration

```bash
bosun config validate
```

**Success output:**
```
Configuration valid (5 entities checked)
```

**Failure output:**
```
Validation errors:

Container "myapp" (abc123def456):
  - unknown key: bosun.container.gracePeriod
  - invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration "30"

Found 2 errors in 1 entity
```

### View Merged Configuration

```bash
bosun config validate --print
```

Shows the effective configuration after merging defaults, file, and labels:

```json
{
  "instance": "homelab",
  "stopGracePeriod": "30s",
  "healthCheckInterval": "10s",
  "autoRestart": false,
  "logLevel": "info",
  "backupEnabled": true,
  "maxSize": "10737418240",
  "priority": 0
}
```

## Common Scenarios

### Scenario 1: Pre-deployment Check

Before deploying a new Docker Compose stack:

```bash
# Deploy the stack
docker compose up -d

# Validate Bosun labels
bosun config validate

# If validation fails, fix labels and redeploy
docker compose down
# ... fix labels in docker-compose.yaml ...
docker compose up -d
bosun config validate
```

### Scenario 2: Debug Label Issues

When Bosun isn't behaving as expected:

```bash
# See what config Bosun is using
bosun config validate --print

# Check only container labels
bosun config validate --scope container --print

# Check only labels (ignore config file)
bosun config validate --from labels --print
```

### Scenario 3: Validate Specific Entity Types

```bash
# Only validate containers
bosun config validate --scope container

# Only validate volumes
bosun config validate --scope volume

# Only validate networks
bosun config validate --scope network
```

### Scenario 4: Include Stopped Containers

By default, only running containers are checked. To include stopped ones:

```bash
bosun config validate --stopped
```

## Common Errors and Fixes

### Unknown Key

```
unknown key: bosun.container.gracePeriod
```

**Problem**: Typo or unsupported label key.

**Fix**: Check the valid keys. Did you mean `bosun.container.stopGracePeriod`?

---

### Invalid Duration

```
invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration "30"
```

**Problem**: Duration values need a unit suffix.

**Fix**: Use `30s`, `5m`, `1h`, etc. Not just `30`.

---

### Scope Mismatch

```
key 'bosun.container.stopGracePeriod' not allowed on scope 'volume'
```

**Problem**: Container-specific label applied to a volume.

**Fix**: Move the label to a container, or use a volume-appropriate label.

---

### Invalid Enum

```
invalid enum value 'debug' for key 'bosun.container.logLevel': must be one of [trace, info, warn, error]
```

**Problem**: Value not in allowed list.

**Fix**: Use one of the allowed values: `trace`, `info`, `warn`, or `error`.

## CI/CD Integration

### GitHub Actions

```yaml
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Start Docker Compose
        run: docker compose up -d

      - name: Install Bosun
        run: go install github.com/simone-viozzi/bosun@latest

      - name: Validate Configuration
        run: bosun config validate
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Start containers if compose file changed
if git diff --cached --name-only | grep -q 'docker-compose'; then
  docker compose up -d
  bosun config validate || exit 1
fi
```

## Tips

1. **Run early, run often**: Validate after any label change
2. **Use `--print` for debugging**: See exactly what Bosun sees
3. **Scope your checks**: Use `--scope` to focus on one entity type
4. **Check stopped containers**: Use `--stopped` if you have containers that aren't always running
5. **Automate in CI**: Catch config errors before deployment
