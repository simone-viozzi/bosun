# Implementation Plan: Config Integration Tests

**Feature Branch**: `005-config-integration-tests`
**Created**: 2025-11-30
**Status**: ✅ Implementation Complete
**Spec**: [spec.md](./spec.md)
**Checklist**: [requirements.md](./checklists/requirements.md)

---

## Executive Summary

This feature adds end-to-end integration tests proving the typed config parsing pipeline works correctly with real Docker containers. The implementation is **~90% complete** - only one small addition is needed to complete the clarified requirements.

---

## Technical Context

### Existing Infrastructure (Reused)
- **Test Harness**: `internal/testutil/harness.go` - `StartCompose()` with unique project names
- **Docker Utilities**: `internal/testutil/docker.go` - client helpers, port discovery
- **Compose Files**: `internal/testutil/compose/*.yaml` - embedded YAML files
- **Build Tag**: `//go:build integration` for selective execution

### Dependencies (Already Implemented)
- Loader (#58): `internal/config/loader/loader.go` - parses labels to ConfigV1
- Merger (#59): `internal/config/merge/merge.go` - merges labels over defaults
- CLI Command: `internal/cmd/validate.go` - `bosun config validate` command

### Current Implementation Files
| File | Status | Purpose |
|------|--------|---------|
| `integration/validate_test.go` | ✅ Complete | 4 integration tests |
| `internal/testutil/compose/validate-valid.yaml` | ✅ Complete | Valid labels for all 6 types |
| `internal/testutil/compose/validate-invalid.yaml` | 🔶 Needs update | Missing invalid bool test |

---

## Implementation Phases

### Phase 0: Research ✅ COMPLETE
No research needed - existing infrastructure and patterns are well-documented in `testing_structure` memory.

### Phase 1: Design ✅ COMPLETE
Test architecture follows established patterns:
- Compose files define labeled Docker resources
- Tests invoke `bosun config validate` CLI
- Assertions check stdout/stderr for expected behavior

### Phase 2: Implementation 🔶 IN PROGRESS

#### Task 2.1: Add Invalid Bool Test Case
**Status**: ✅ COMPLETE
**Effort**: 5 minutes
**File**: `internal/testutil/compose/validate-invalid.yaml`

Added `bosun.container.autoRestart=maybe` to test bool parse error per US-4 Scenario 2.

Also updated volume label to use `bosun.container.stopGracePeriod` for scope mismatch test (container-scoped label on volume).

#### Task 2.2: Verify Test Coverage ✅ COMPLETE
All other tests are implemented and passing:
- `Test_Integration_ConfigValidate_ValidConfig` - happy path
- `Test_Integration_ConfigValidate_InvalidConfig` - error cases
- `Test_Integration_ConfigValidate_PrintFlag` - JSON output
- `Test_Integration_ConfigValidate_ScopeFlag` - scope filtering

### Phase 3: Testing ✅ COMPLETE
Integration tests are self-testing. Run with:
```bash
make it  # or make itv for verbose
```

### Phase 4: Documentation ✅ COMPLETE
- Checklist updated with implementation evidence
- Spec clarified with Q&A session
- Memory files document test patterns

---

## Remaining Work

| Task | Priority | Effort | Status |
|------|----------|--------|--------|
| ~~Add invalid bool test case~~ | P2 | 5 min | ✅ Done |
| Create multi-container follow-up issue | P3 | 5 min | Pending |
| ~~Run final integration test suite~~ | P1 | 2 min | ✅ Pass |

---

## Verification Steps

1. **Add invalid bool test**:
   ```bash
   # Edit validate-invalid.yaml
   # Add bosun.container.autoRestart: "maybe" to test bool parsing
   ```

2. **Run integration tests**:
   ```bash
   make it
   ```

3. **Verify all tests pass**:
   - `Test_Integration_ConfigValidate_ValidConfig` ✅
   - `Test_Integration_ConfigValidate_InvalidConfig` ✅ (now includes bool error)
   - `Test_Integration_ConfigValidate_PrintFlag` ✅
   - `Test_Integration_ConfigValidate_ScopeFlag` ✅

4. **Update checklist** to mark FR-005 and US-4 Scenario 2 complete

---

## Follow-up Issues

### Issue: Multi-container Integration Tests
**Title**: Add multi-container integration tests for per-entity config validation
**Priority**: P3 (Deferred)
**Rationale**: Per clarification session, single-container tests sufficiently prove pipeline correctness. Multi-container testing adds complexity without significant value for current use cases.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Invalid bool test conflicts with scope test | Medium | Low | Use separate service in compose file |
| Tests flaky due to Docker timing | Low | Medium | 3-min timeout provides safety margin |
| CI doesn't have Docker | Low | High | CI environment requirements documented |

---

## Definition of Done

- [x] All 6 config types tested in happy path (FR-002)
- [x] Unknown key rejection tested (FR-003)
- [x] Scope mismatch tested (FR-004)
- [x] Invalid bool parse error tested (FR-005)
- [x] Full pipeline tested with --print (FR-006)
- [x] Cleanup verified (FR-007)
- [x] Unique project names (FR-008)
- [x] Build tag integration (FR-009)
- [x] Clear error on Docker unavailable (FR-010)
- [x] Checklist fully green
- [ ] Multi-container follow-up issue created
