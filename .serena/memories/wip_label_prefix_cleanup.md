# WIP: Label Prefix Documentation Cleanup

**Created**: 2025-12-29
**Status**: Documentation debt

## Issue

The M3 spec documents and decision memories incorrectly reference `bosun.backup.*` labels, but the correct prefix is `bosun.job.*`.

**Rationale**: Bosun is a **job orchestrator**, not a backup tool. While backups are a common use case, Bosun can orchestrate any job that requires stopping a stack, running a worker, and restarting. The `bosun.job.*` prefix correctly reflects this generality.

## Implementation Status

The **implementation is correct** — all code uses `bosun.job.*`:
- `bosun.job.stop-timeout`
- `bosun.job.start-timeout`
- `bosun.job.timeout`
- `bosun.job.worker.env.*`

## Files Needing Update

### Spec Documents (require manual edit)
- `specs/009-job-execution-mvp/spec.md` — FR-005, FR-011 reference `bosun.backup.*`
- `specs/009-job-execution-mvp/plan.md` — Label table uses `bosun.backup.*`
- `specs/009-job-execution-mvp/data-model.md` — Label definitions use `bosun.backup.*`
- `specs/009-job-execution-mvp/research.md` — Examples use `bosun.backup.*`
- `specs/009-job-execution-mvp/tasks.md` — T004 description uses `bosun.backup.*`

### Decision Memories (can be updated via Serena)
- `m3_failure_handling.md` — Timeout labels, label summary table
- `m3_worker_contract.md` — Worker env examples, label format

## Action

Low priority — update when touching these files for other reasons, or batch update before M3 completion.

Search pattern to find all occurrences:
```
bosun\.backup\.
```
