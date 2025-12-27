# Implementation Plan: Strict golangci-lint Checks

**Branch**: `008-strict-golangci-lint` | **Date**: 2025-12-27 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/008-strict-golangci-lint/spec.md`
**Related Issue**: [#101](https://github.com/simone-viozzi/bosun/issues/101)

## Summary

Enable all disabled golangci-lint linters and staticcheck rules by fixing existing violations. This involves:
1. Fixing 5 prealloc suggestions (slice pre-allocation)
2. Fixing 3 unparam violations (unused parameters / always-nil returns)
3. Adding package comments to 4 packages (ST1000)
4. Fixing 1 inconsistent receiver name (ST1016)
5. Fixing 1 comment format issue (ST1022)
6. Addressing 7 unusedwrite violations in test files
7. Updating `.golangci.yml` to enable all linters

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: golangci-lint (dev tool), no runtime dependencies affected
**Storage**: N/A
**Testing**: `golangci-lint run` and `golangci-lint config verify`
**Target Platform**: Development tooling (CI/local)
**Project Type**: Single CLI project
**Performance Goals**: N/A (linting is dev-time only)
**Constraints**: All violations must be genuinely fixed (not just suppressed with nolint)
**Scale/Scope**: ~23 violations across ~10 files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Hexagonal Architecture | ✅ PASS | No architectural changes - only code quality fixes |
| II. Label-Driven Configuration | ✅ PASS | No configuration changes |
| III. Test-First Development | ✅ PASS | Test files included in lint scope |
| IV. CLI-First Interface | ✅ PASS | No CLI changes |
| V. Code Quality & Simplicity | ✅ PASS | Directly improves code quality compliance |

**Gate Result**: ✅ PASS - No violations. Feature aligns with constitution principles.

### Post-Design Re-evaluation

After completing Phase 1 design:
- ✅ All changes remain within code quality scope
- ✅ No new architectural patterns introduced
- ✅ No new dependencies added
- ✅ Test files included and improved
- ✅ CLI interface unchanged

**Final Gate Status**: ✅ PASS

## Project Structure

### Documentation (this feature)

```text
specs/008-strict-golangci-lint/
├── plan.md              # This file
├── research.md          # Phase 0 output - best practices for fixes
├── data-model.md        # Configuration model documentation
├── quickstart.md        # Phase 1 output - verification commands
├── contracts/           # N/A - no API contracts
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (files to modify)

```text
# Configuration
.golangci.yml                              # Enable linters, remove exclusions

# prealloc fixes (5 violations)
internal/adapters/dockerlabels/source.go   # 3 fixes: lines 72, 113, 148
internal/cmd/validate.go                   # 1 fix: line 213
internal/tools/configdoc/markdown.go       # 1 fix: line 100

# unparam fixes (3 violations)
internal/config/loader/loader.go           # 1 fix: line 26 - setField returns always-nil
internal/config/schema/defaults.go         # 1 fix: line 93 - fieldType unused
internal/config/schema/tags.go             # 1 fix: line 17 - parseTagValue returns always-nil

# staticcheck ST1000 fixes (4 packages need doc comments)
internal/adapters/dockerlabels/doc.go      # Create: package comment
internal/app/doc.go                        # Create: package comment
internal/domain/labels/doc.go              # Create: package comment
internal/ports/doc.go                      # Create: package comment

# staticcheck ST1016 fix (1 violation)
internal/adapters/dockerlabels/source.go   # Fix receiver name: d → s on Snapshot

# staticcheck ST1022 fix (1 violation)
internal/domain/labels/types.go            # Fix comment format on LabelInstance

# unusedwrite fixes (7 violations in test files)
internal/domain/labels/types_test.go       # 1 fix: line 45
internal/ports/labels_test.go              # 6 fixes: lines 54-55, 71-72, 87-88
```

**Structure Decision**: No new directories. Adding `doc.go` files for package comments following Go convention.

## Complexity Tracking

> No complexity violations - all changes are straightforward code quality fixes.

| Item | Complexity | Rationale |
|------|------------|-----------|
| prealloc fixes | Low | Simple `make([]T, 0, len(source))` pattern |
| unparam fixes | Low | Remove unused params or simplify return types |
| Package comments | Low | Add standard doc.go files |
| Receiver name fix | Low | Single rename from `d` to `s` |
| Comment format fix | Low | Update comment to match Go convention |
| unusedwrite fixes | Low | Remove or use the assigned values |
