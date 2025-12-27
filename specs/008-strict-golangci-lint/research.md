# Research: Strict golangci-lint Checks

**Feature**: 008-strict-golangci-lint
**Date**: 2025-12-27
**Purpose**: Document best practices and patterns for fixing each category of violations

## 1. prealloc Fixes

### Decision
Pre-allocate slices using `make([]T, 0, len(source))` where the capacity can be determined from the loop source.

### Rationale
- Avoids repeated slice reallocations during append operations
- Memory is allocated once upfront
- Standard Go performance best practice

### Pattern
```go
// Before
var out []T
for _, item := range source {
    out = append(out, transform(item))
}

// After
out := make([]T, 0, len(source))
for _, item := range source {
    out = append(out, transform(item))
}
```

### Alternatives Considered
- **Ignore with nolint**: Rejected - these are genuine performance improvements
- **Use exact length**: Would require knowing final size, not always possible with filtering

---

## 2. unparam Fixes

### Decision
Simplify function signatures by removing always-nil error returns or unused parameters.

### Rationale
- Unused parameters create confusion about function contracts
- Always-nil errors are misleading and add unnecessary error handling at call sites
- Go philosophy: keep interfaces minimal and honest

### Pattern A: Always-nil error return
```go
// Before
func parseTagValue(tagValue string) (map[string]string, error) {
    // ... logic that never returns error
    return result, nil
}

// After
func parseTagValue(tagValue string) map[string]string {
    // ... logic
    return result
}
```

### Pattern B: Unused parameter
```go
// Before
func setDefaultValue(fieldVal reflect.Value, fieldType reflect.Type, ...) error {
    // fieldType never used
}

// After
func setDefaultValue(fieldVal reflect.Value, ...) error {
    // removed fieldType
}
```

### Alternatives Considered
- **Keep for future use**: Rejected - YAGNI principle; add when needed
- **Use `_` parameter**: Rejected - still confusing interface

---

## 3. staticcheck ST1000 - Package Comments

### Decision
Create `doc.go` files with package-level documentation following Go conventions.

### Rationale
- Package comments are displayed in `go doc` and godoc
- Help developers understand package purpose at a glance
- Required by Go style guidelines

### Pattern
```go
// Package name provides a brief description of what the package does.
//
// Additional context about usage, design decisions, or important types
// can be included in subsequent paragraphs.
package name
```

### Location Decision
- Use separate `doc.go` files for clarity
- Keeps documentation separate from implementation
- Standard Go convention for packages with multiple files

---

## 4. staticcheck ST1016 - Receiver Names

### Decision
Use consistent single-letter receiver names throughout a type's methods.

### Rationale
- Go style guide recommends short (1-2 letter) receiver names
- Consistency within a type improves readability
- The dominant name in the type should be used everywhere

### Current Issue
`DockerLabelSource` uses:
- `s` in 3 methods
- `d` in 1 method (Snapshot)

### Fix
Change `d` to `s` in the Snapshot method.

---

## 5. staticcheck ST1022 - Comment Format

### Decision
Update exported constant comments to follow Go convention: `// Name description`.

### Rationale
- Godoc expects specific format for proper rendering
- Comments starting with the identifier name render correctly
- Improves generated documentation

### Pattern
```go
// Before
// TODO this cannot be here, we need a better way of handling this
const LabelInstance = "bosun.instance"

// After
// LabelInstance defines the label key for marking backup job instances.
const LabelInstance = "bosun.instance"
```

---

## 6. unusedwrite in Test Files

### Decision
Either use the struct fields in assertions or remove the unnecessary assignments.

### Rationale
- Unused writes may indicate incomplete test coverage
- If intentionally unused (e.g., creating struct for other purposes), restructure to avoid false positives

### Current Violations
- `types_test.go:45` - Unused write to `TakenAt` field
- `labels_test.go:54-55, 71-72, 87-88` - Unused writes to `Prefixes` and `IncludeStopped`

### Analysis
These appear to be test setup where fields are set but not relevant to the specific test assertion. Options:
1. Add assertions that use the fields
2. Use struct literal without unused fields
3. Add exclusion rule for specific test patterns (less preferred)

### Decision
Review each case and either add meaningful assertions or simplify struct initialization.

---

## 7. errcheck Status

### Current Finding
Running `golangci-lint run --enable errcheck` shows **0 issues**.

### Analysis
The issue #101 mentioned 6 errcheck violations at:
- `integration/smoke_placeholder_test.go:49` - `resp.Body.Close()`
- `internal/testutil/harness.go:54,59,83,93` - `os.RemoveAll()` and `st.Down()`

These may have been fixed in a previous PR or the linter behavior changed. Current baseline shows no violations.

### Decision
Enable errcheck in config. If violations reappear, handle with explicit `_ =` assignment pattern.

---

## Summary of Changes Required

| Category | Count | Files | Effort |
|----------|-------|-------|--------|
| prealloc | 5 | 3 files | Low |
| unparam | 3 | 3 files | Low |
| ST1000 (package comments) | 4 | 4 new doc.go files | Low |
| ST1016 (receiver name) | 1 | 1 file | Low |
| ST1022 (comment format) | 1 | 1 file | Low |
| unusedwrite | 7 | 2 files | Low |
| Config update | 1 | 1 file | Low |
| **Total** | **22** | **~12 files** | **Low** |
