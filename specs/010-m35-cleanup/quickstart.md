# Quickstart: Milestone 3.5 Verification

**Date**: 2026-01-05

## Prerequisites

```bash
# Build bosun
make build

# Start test compose stack
docker compose -p joblabels -f internal/testutil/compose/joblabels-compose.yaml up -d
```

## Verification Commands

### US1: Config Validation (Lenient Mode)

```bash
# Should pass with warnings (not errors) for unknown keys
bin/bosun config validate --project joblabels

# Expected output (after fix):
# Warnings:
#   container "joblabels-no-job-labels-1": unknown key: bosun.other (ignored)
# Validation passed with 1 warning(s)

# Strict mode should fail
bin/bosun config validate --project joblabels --strict
# Expected: exit code 2, shows errors
```

### US2: Complete Execution Plan

```bash
# Should show stop → worker → start steps
bin/bosun plan show daily-backup --project joblabels --format json

# Expected output (after fix):
# {
#   "jobName": "daily-backup",
#   "steps": [
#     {"type": "stop_containers", ...},
#     {"type": "run_worker", ...},
#     {"type": "start_containers", ...}  ← NEW
#   ]
# }
```

### US3: Terminology Cleanup

```bash
# Should show "Enabled", not "BackupEnabled"
bin/bosun config validate --print | grep -i enabled

# Expected output (after fix):
# "Enabled": false

# Grep for backup in production code (should be empty)
grep -r "backup" internal/ --include="*.go" | grep -v _test.go | grep -v "// Example"
# Expected: no output (or only doc/example contexts)
```

### US4: Executor Interface

```bash
# Run executor tests
go test ./internal/app/executor/... -v

# Check for interface compliance
go build ./...  # Should compile without "not implemented" stubs
```

### US5: CLI Error Output

```bash
# Should show error only once
bin/bosun invalidcommand 2>&1 | wc -l
# Expected: reasonable line count (not duplicated)

bin/bosun invalidcommand
# Expected output (after fix):
# Error: unknown command "invalidcommand" for "bosun"
# Run 'bosun --help' for usage.
```

### US6: README Documentation

```bash
# Check README has job section
grep -A5 "Running Jobs" README.md
# Expected: section exists with examples
```

## Success Criteria Validation

```bash
# SC-001: Config validation passes
bin/bosun config validate --project joblabels
echo "Exit code: $?"  # Should be 0

# SC-002: Plan shows all steps
bin/bosun plan show daily-backup --project joblabels --format json | jq '.steps | length'
# Should be 3 (stop, worker, start)

# SC-003: No BackupEnabled in production code
grep -r "BackupEnabled" internal/ --include="*.go" | grep -v _test.go | wc -l
# Should be 0

# SC-004: CLI doesn't import adapters
grep -r "internal/adapters" internal/cmd/ | wc -l
# Should be 0

# SC-005: All tests pass
make test
make it  # integration tests

# SC-006: README has Running Jobs section
grep -c "Running Jobs" README.md
# Should be >= 1
```

## Cleanup

```bash
# Stop test stack
docker compose -p joblabels -f internal/testutil/compose/joblabels-compose.yaml down -v
```
