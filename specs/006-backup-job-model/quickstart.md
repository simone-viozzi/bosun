# Quickstart: Job Model & Planning

**Feature Branch**: `006-backup-job-model`
**Date**: 2025-11-30

## Overview

This guide shows you how to define jobs using Docker labels and preview execution plans. This milestone introduces job discovery and planning only—no Docker containers will be stopped or started.

---

## Prerequisites

- Docker installed and running
- Bosun CLI built (`make build`)
- A Docker Compose stack to work with

---

## Step 1: Add Job Labels to Your Containers

Add labels to your `docker-compose.yml` to define a job:

```yaml
# docker-compose.yml
services:
  web:
    image: nginx:latest
    labels:
      # Enable job discovery for this container
      bosun.job.enabled: "true"
      # Name the job (required when enabled)
      bosun.job.name: "daily-maintenance"
      # Set the schedule (optional, defaults to daily at midnight)
      bosun.job.schedule: "0 2 * * *"  # 2 AM daily
      # Specify worker image (optional)
      bosun.job.worker.image: "alpine:latest"

  api:
    image: myapi:latest
    labels:
      # This container also participates in the same job
      bosun.job.enabled: "true"
      bosun.job.name: "daily-maintenance"
      # Schedule and worker are inherited from first container

volumes:
  data:
    labels:
      # Attach this volume to the job (read-only by default)
      bosun.job.attach: "daily-maintenance"

  config:
    labels:
      # Attach with explicit read-write access
      bosun.job.attach: "daily-maintenance:rw"
```

---

## Step 2: Start Your Stack

```bash
docker compose up -d
```

---

## Step 3: List Discovered Jobs

```bash
bosun plan list
```

**Expected Output**:
```
JOBS DISCOVERED: 1

NAME                SCHEDULE        CONTAINERS  VOLUMES  STACK
daily-maintenance   0 2 * * *       2           2        myproject

Run 'bosun plan show <name>' for execution details.
```

---

## Step 4: Preview Execution Plan

```bash
bosun plan show daily-maintenance
```

**Expected Output**:
```
EXECUTION PLAN: daily-maintenance

Job:      daily-maintenance
Schedule: 0 2 * * * (Daily at 2:00 AM)
Stack:    myproject

STEPS:

1. STOP CONTAINERS
   Using: docker compose stop (all stack containers included)
   Project: myproject
   Containers:
     - myproject-web-1 (abc123)
     - myproject-api-1 (def456)

2. RUN WORKER
   Image: alpine:latest
   Volumes:
     - data → /data/data (ro)
     - config → /data/config (rw)
   Command: [to be configured]

Note: Container restart (step 3) is not included in this milestone.
```

---

## Step 5: Export Plan as JSON

For scripting and automation:

```bash
bosun plan show daily-maintenance --format json > plan.json
```

Or pipe to `jq`:

```bash
bosun plan show daily-maintenance --format json | jq '.steps'
```

---

## Label Reference

### Container Labels

| Label | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `bosun.job.enabled` | bool | No | `false` | Enable job discovery |
| `bosun.job.name` | string | Yes* | — | Job identifier (*required if enabled) |
| `bosun.job.schedule` | string | No | `0 0 * * *` | Cron expression |
| `bosun.job.worker.image` | string | No | built-in | Worker container image |

### Volume Labels

| Label | Type | Description |
|-------|------|-------------|
| `bosun.job.attach` | string | Attach to job: `<job-name>` or `<job-name>:ro`/`<job-name>:rw` |

---

## Common Patterns

### Multiple Jobs, Same Container

A container can participate in multiple jobs:

```yaml
labels:
  bosun.job.enabled: "true"
  bosun.job.name: "backup,cleanup,health-check"
```

### Stack Override

Override the automatic Compose project detection:

```yaml
labels:
  bosun.stack: "my-custom-stack"
```

### Volumes from Multiple Stacks

Volumes can attach to jobs even if they're in a different stack:

```yaml
volumes:
  shared-data:
    labels:
      bosun.job.attach: "daily-maintenance"
```

---

## Troubleshooting

### No Jobs Discovered

```bash
bosun plan list
# No jobs discovered.
```

**Check**:
1. Containers are running: `docker ps`
2. Labels are applied: `docker inspect <container> | jq '.[0].Config.Labels'`
3. Label syntax is correct (no typos in `bosun.job.enabled`)

### Validation Errors

```bash
bosun plan list
# VALIDATION ERRORS:
# Container "web": bosun.job.schedule: invalid cron expression
```

**Fix**: Correct the cron expression in your compose file and recreate:

```bash
docker compose up -d --force-recreate
```

### Include Stopped Containers

By default, stopped containers are excluded. To include them:

```bash
bosun plan list --stopped
bosun plan show daily-maintenance --stopped
```

---

## Next Steps

This milestone introduces job discovery and plan generation only. Future milestones will add:

- **Milestone 3**: Job execution (actually stopping containers, running workers)
- **Milestone 4**: Scheduled execution via cron
- **Milestone 5**: Job result verification and notifications

---

## Quick Reference

```bash
# List all jobs
bosun plan list

# List jobs as JSON
bosun plan list --format json

# Show plan for a job
bosun plan show <job-name>

# Show plan as YAML
bosun plan show <job-name> --format yaml

# Include stopped containers
bosun plan list --stopped

# Filter by stack
bosun plan list --stack myproject
```
