# Research: Label Parser and Source Merger

**Feature Branch**: `002-label-parser-merger`
**Date**: 2025-11-29
**Spec**: [spec.md](./spec.md)

## Research Tasks

### 1. Existing Schema Package Analysis

**Question**: What parsing functionality already exists in the schema package?

**Finding**: The `internal/config/schema/defaults.go` already contains parsing functions:
- `parseBool(s string) (bool, error)` - Boolean parsing
- `parseInt(s string) (int, error)` - Integer parsing
- `parseDuration(s string) (time.Duration, error)` - Duration parsing using `time.ParseDuration`
- `parseSize(s string) (int64, error)` - Byte size parsing using `units.RAMInBytes`
- `parseList(s string) []string` - List parsing (CSV only, no error return)

**Decision**: Reuse existing parsers but enhance:
- Move parsers to new `loader` package or export from `schema`
- Add JSON array support to `parseList`
- Add enum validation function
- Add proper error handling with context

**Rationale**: Don't duplicate parsing logic; extend what exists.

---

### 2. Error Handling Pattern

**Question**: How should multiple validation errors be collected and reported?

**Finding**: Go standard library patterns:
- `errors.Join` (Go 1.20+) for combining errors
- Custom error type with slice of validation errors
- Each error should include: key, value, expected type, actual error

**Decision**: Create `ValidationError` struct and `ValidationErrors` collection:
```go
type ValidationError struct {
    Key     string
    Value   string
    Scope   Scope
    Message string
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string { ... }
```

**Rationale**: Users need all errors at once, not fail-fast on first error.

---

### 3. Scope Validation Logic

**Question**: How should scope validation work with the "global" scope?

**Finding**: From spec and schema types:
- `ScopeGlobal` labels can appear on any entity type
- Other scopes (`ScopeContainer`, `ScopeVolume`, `ScopeNetwork`) must match exactly
- The caller provides the entity's scope when calling `FromLabels`

**Decision**: Scope check logic:
```go
func scopeAllowed(fieldScope, entityScope Scope) bool {
    return fieldScope == ScopeGlobal || fieldScope == entityScope
}
```

**Rationale**: Simple, explicit logic matching the spec requirements.

---

### 4. Config Merge Strategy

**Question**: How to detect "unset" fields when merging?

**Finding**: Options considered:
1. Use pointer types (`*time.Duration`) - nil means unset
2. Use zero values with explicit "set" tracking
3. Use reflection to compare against zero values

**Decision**: Use reflection-based merging that compares against zero values:
- Simple for v1 implementation
- Matches Go's natural zero-value semantics
- Can upgrade to pointer types later if needed

**Rationale**: Simpler implementation; avoids changing ConfigV1 struct to use pointers everywhere.

**Alternatives Rejected**:
- Pointer types: More complex API, harder to use
- Explicit "set" tracking: Over-engineering for v1

---

### 5. Environment Variable Naming Convention

**Question**: How should env vars map to config keys?

**Finding**: Common patterns:
- `BOSUN_CONTAINER_STOPGRACEPERIOD` (flat, uppercase, underscores)
- `BOSUN_CONTAINER_STOP_GRACE_PERIOD` (snake_case conversion)

**Decision**: Defer full env implementation (feature flag), but plan for:
- Format: `BOSUN_<SCOPE>_<FIELD>` uppercase with underscores
- Example: `bosun.container.stopGracePeriod` → `BOSUN_CONTAINER_STOPGRACEPERIOD`

**Rationale**: Feature flag means we don't need to finalize this now.

---

### 6. List Parsing: CSV vs JSON

**Question**: How to distinguish CSV from JSON array format?

**Finding**: Simple heuristic:
- If string starts with `[` and ends with `]`, try JSON parse first
- Otherwise, treat as CSV
- If JSON parse fails, don't fall back to CSV (could mask errors)

**Decision**:
```go
func parseList(s string) ([]string, error) {
    s = strings.TrimSpace(s)
    if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
        var result []string
        if err := json.Unmarshal([]byte(s), &result); err != nil {
            return nil, fmt.Errorf("invalid JSON array: %w", err)
        }
        return result, nil
    }
    // CSV fallback
    return parseCSV(s), nil
}
```

**Rationale**: Clear rules, predictable behavior, explicit errors.

---

### 7. Package Structure

**Question**: Where should loader and merger live?

**Finding**: Project structure analysis:
- `internal/config/schema/` - Schema definitions (exists)
- `internal/config/` - Config package root

**Decision**: Create two new packages:
- `internal/config/loader/` - Label parsing and validation
- `internal/config/merge/` - Multi-source config merging

**Rationale**: Clean separation, matches issue structure (#58 = loader, #59 = merger).

---

## Summary

| Decision | Choice | Why |
|----------|--------|-----|
| Parser reuse | Export/enhance existing parsers | Don't duplicate logic |
| Error handling | Custom `ValidationErrors` type | Collect all errors |
| Scope logic | Global allowed anywhere, others exact match | Matches spec |
| Merge strategy | Reflection with zero-value check | Simple for v1 |
| Env naming | Deferred (feature flag) | Not needed for MVP |
| List parsing | JSON if `[...]`, else CSV | Clear rules |
| Package structure | `loader/` + `merge/` | Matches issues |
