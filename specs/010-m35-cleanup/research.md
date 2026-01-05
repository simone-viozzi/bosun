# Research: Milestone 3.5 - Post-M3 Cleanup & Bug Fixes

**Date**: 2026-01-05
**Status**: Complete (decisions pre-made in smell analysis)

## Overview

This milestone's research was conducted during the code smell analysis phase. All technical decisions are documented in project memories:

- **`wip_smell_milestone3.md`** — 5 blocking decisions resolved
- **`wip_smell-design-m3.md`** — 15 design findings with remediation directions
- **`wip_m35.md`** — 10 GitHub issues with requirements

## Decision Summary

| ID | Decision | Rationale | Memory Reference |
|----|----------|-----------|------------------|
| #1 | **Thin CLI** — Move wiring to app layer | Hexagonal architecture best practice; CLI should only parse/present | `wip_smell_milestone3.md` Decision #1 |
| #2 | **ExecuteJob(ctx, job)** — External discovery | Single responsibility; executor executes, discoverer discovers | `wip_smell_milestone3.md` Decision #2 |
| #3 | **Rename BackupEnabled now** — Breaking change OK | Pre-1.0, clean up naming debt early | `wip_smell_milestone3.md` Decision #3 |
| #4 | **Plan-is-source-of-truth** — Executor interprets plan | Single source of truth; DryRun accuracy; extensibility | `wip_smell_milestone3.md` Decision #4 |
| #5 | **Lenient unknown keys** — Warn, don't error | Graceful degradation; rolling deployments; extension-friendly | `wip_smell_milestone3.md` Decision #5 |

## Best Practice Verification

### Decision #1: Thin CLI

**Claim**: CLI should be thin (only argument parsing + presentation)

**Verification**: ✅ Confirmed
- Hexagonal architecture: CLI is a driving adapter, should not contain business logic
- *"Controller layer shouldn't contain business logic"* (Clean Architecture)
- Enables unit testing CLI without Docker, reusable services for future entrypoints

### Decision #4: Plan-Driven Execution

**Claim**: Execution plans should be authoritative and drive runtime execution

**Verification**: ✅ Confirmed
- Single source of truth prevents divergence between preview and execution
- Used by: Terraform (plan→apply), DB query planners, CI/CD systems, Kubernetes
- Follows Command/Interpreter pattern
- Enables per-step policies (retry, timeout) in future

### Decision #5: Lenient Validation

**Claim**: Unknown label keys should warn, not error

**Verification**: ✅ Confirmed
- Docker and Kubernetes are permissive with unknown labels
- Graceful degradation improves UX
- Rolling deployments won't break on new label versions
- `--strict` flag preserves original behavior for users who need it

## Technology Choices

No new technologies required. All changes use existing stack:

| Component | Technology | Already in Use |
|-----------|------------|----------------|
| CLI Framework | Cobra | ✅ Yes |
| Docker Integration | Docker SDK | ✅ Yes |
| Testing | go test, testcontainers | ✅ Yes |
| Logging | slog | ✅ Yes |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking change to `backupEnabled` label | Medium | Low | Pre-1.0; document in release notes |
| Executor refactoring introduces bugs | Medium | Medium | Comprehensive test coverage; integration tests |
| Lenient validation masks real errors | Low | Medium | `--strict` flag; clear warning messages |

## Open Questions Resolved

All blocking questions from `wip_smell_milestone3.md` have been answered:

1. ~~Unknown label key handling~~ → **Lenient/warning**
2. ~~CLI thickness~~ → **Thin CLI**
3. ~~ExecutionPlan authority~~ → **Plan-is-source-of-truth**
4. ~~ExecuteJob API contract~~ → **Option B (external discovery)**
5. ~~BackupEnabled rename timing~~ → **Rename now (breaking OK)**

## Conclusion

No additional research needed. Proceed to Phase 1 (Design & Contracts).
