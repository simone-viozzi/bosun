# Feature Specification: Strict golangci-lint Checks

**Feature Branch**: `008-strict-golangci-lint`
**Created**: 2025-12-27
**Status**: Draft
**Input**: User description: "Enable stricter golangci-lint checks and fix existing violations"
**Related Issue**: [#101](https://github.com/simone-viozzi/bosun/issues/101)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Clean Lint Runs for Contributors (Priority: P1)

As a developer contributing to the Bosun project, I want `golangci-lint run` to pass with zero issues so that I can be confident my code meets quality standards without manually tracking which linters are disabled.

**Why this priority**: This is the core value of the feature - contributors need a single, reliable lint command that validates all code quality standards. Currently, disabled linters mean potential issues slip through.

**Independent Test**: Run `golangci-lint run` on the codebase and verify it returns 0 issues and exits successfully.

**Acceptance Scenarios**:

1. **Given** a clean checkout of the codebase, **When** I run `golangci-lint run`, **Then** the command exits with code 0 and reports no issues
2. **Given** code with an unchecked error return, **When** I run `golangci-lint run`, **Then** errcheck flags the violation
3. **Given** a slice that could be pre-allocated, **When** I run `golangci-lint run`, **Then** prealloc suggests the improvement

---

### User Story 2 - Consistent Code Style via Staticcheck (Priority: P2)

As a maintainer, I want all staticcheck rules enabled so that the codebase maintains consistent documentation and naming conventions without manual enforcement.

**Why this priority**: Documentation and naming consistency improves long-term maintainability, but the codebase can function without perfect comments. This is secondary to catching actual bugs.

**Independent Test**: Enable all ST* rules in `.golangci.yml` and verify no violations are reported.

**Acceptance Scenarios**:

1. **Given** a package without a doc comment, **When** staticcheck runs with ST1000 enabled, **Then** the violation is flagged
2. **Given** a type with inconsistent receiver names, **When** staticcheck runs with ST1016 enabled, **Then** the violation is flagged
3. **Given** all packages with proper comments and receiver names, **When** I run `golangci-lint run`, **Then** no ST* violations are reported

---

### User Story 3 - Test Code Quality via govet (Priority: P3)

As a developer, I want govet analyzers running on test code so that test quality issues are caught early.

**Why this priority**: Test-specific analyzers like `unusedwrite` can be noisy but help maintain test quality. This is lower priority than production code quality.

**Independent Test**: Enable `unusedwrite` analyzer and verify test files pass or have appropriate exclusions.

**Acceptance Scenarios**:

1. **Given** test code with unused struct field writes, **When** govet runs with unusedwrite enabled, **Then** violations are either fixed or explicitly excluded
2. **Given** the full test suite, **When** I run `golangci-lint run`, **Then** govet analyzers report no issues in test files

---

### Edge Cases

- What happens when new code is added with pre-allocatable slices? → prealloc will flag it, guiding the contributor
- How does system handle intentionally unused parameters (e.g., interface compliance)? → Use `nolint:unparam` directive or exclude in config
- What about error returns that are intentionally ignored in cleanup code? → Use explicit `_ =` assignment or `nolint:errcheck` with justification

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST pass `golangci-lint run` with zero issues on the entire codebase
- **FR-002**: System MUST pass `golangci-lint config verify` to ensure configuration validity
- **FR-003**: Configuration MUST enable `errcheck` linter with all existing violations fixed
- **FR-004**: Configuration MUST enable `prealloc` linter with all suggestions addressed
- **FR-005**: Configuration MUST enable `unparam` linter with unused parameters fixed or removed
- **FR-006**: Configuration MUST enable staticcheck rules ST1000, ST1016, ST1022 with violations fixed
- **FR-007**: All packages MUST have package-level documentation comments (for ST1000)
- **FR-008**: All types MUST use consistent receiver names (for ST1016)
- **FR-009**: All exported constants MUST have properly formatted comments (for ST1022)
- **FR-010**: Configuration MUST evaluate and address `unusedwrite` violations in test files

### Key Entities

- **Linter Configuration** (`.golangci.yml`): Central configuration controlling which linters are enabled, their settings, and exclusion rules
- **Package Comments**: Go doc comments at the top of package files providing package-level documentation
- **Receiver Names**: The variable name used for method receivers on types (e.g., `func (s *Server)` vs `func (srv *Server)`)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `golangci-lint run` exits with code 0 and reports 0 issues
- **SC-002**: `golangci-lint config verify` passes without errors
- **SC-003**: No linters remain in the `disable` section of `.golangci.yml`
- **SC-004**: No staticcheck rules are excluded (no `-ST*` patterns in checks)
- **SC-005**: All 6 errcheck violations from the baseline are resolved
- **SC-006**: All 5 prealloc suggestions from the baseline are addressed
- **SC-007**: All 3 unparam violations from the baseline are resolved
- **SC-008**: All 8 staticcheck violations from the baseline are resolved
- **SC-009**: CI pipeline continues to pass after changes

## Assumptions

- The baseline violation counts from issue #101 are accurate and current
- Performance impact of pre-allocating the identified slices is negligible or beneficial
- The `shadow` and `fieldalignment` govet analyzers will remain disabled as they are considered "too noisy for this project" per the existing config comments (out of scope for this feature)
- Any `nolint` directives added will include justification comments
