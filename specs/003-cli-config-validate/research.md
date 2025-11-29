# Research: CLI Config Validate Command

**Date**: 2025-11-30
**Feature**: 003-cli-config-validate

## Research Topics

### 1. Existing CLI Command Pattern

**Question**: How are CLI commands structured in this project?

**Findings**:
- Commands are in `internal/cmd/` using Cobra framework
- `NewRootCmd()` creates root command and registers subcommands
- `NewLabelsCmd()` creates a command group, `NewSnapshotCmd()` is a subcommand
- Pattern: `New<Name>Cmd() *cobra.Command` returns a configured Cobra command
- RunE pattern used for error handling (not Run)
- Context passed via `cmd.Context()`

**Decision**: Follow existing pattern - create `NewConfigCmd()` group and `NewValidateCmd()` subcommand.

**Alternatives Rejected**:
- Adding validate as a flag to snapshot command - violates single responsibility
- Putting validate in labels group - config validation is conceptually separate from label discovery

---

### 2. Validation Error Handling

**Question**: How should validation errors be collected and reported?

**Findings**:
- `loader.ValidationErrors` (slice of `ValidationError`) already exists
- Implements `error` interface with multi-line formatted output
- Each `ValidationError` has: Key, Value, Scope, Message, Err
- Helper methods: `AddUnknownKey`, `AddScopeMismatch`, `AddTypeParseFailed`, `AddInvalidEnum`, `AddRequiredMissing`
- Collects ALL errors, not just first (FR-003 requirement)

**Decision**: Reuse `loader.ValidationErrors` directly - no new types needed.

**Alternatives Rejected**:
- Creating new error types - duplication of existing functionality
- Failing on first error - spec requires all errors to be reported

---

### 3. Configuration Source Selection

**Question**: How to implement `--from` flag (labels, file, auto)?

**Findings**:
- `merge.Merge()` accepts optional pointers for file, env, labels layers
- Passing `nil` skips a layer
- `--from labels`: Pass `nil` for file layer
- `--from file`: Pass `nil` for labels layer (but need file loader first)
- `--from auto`: Load both and merge (default)

**Decision**: Implement `--from` as enum flag, control which layers are loaded/merged.

**Alternatives Rejected**:
- Separate commands per source - unnecessarily complex CLI

---

### 4. Per-Entity Validation

**Question**: How to validate labels per Docker entity (container, volume, network)?

**Findings**:
- `dockerlabels.Snapshot()` returns `labels.Snapshot` with `[]Entity`
- Each `Entity` has: Kind, ID, Name, Labels, Meta
- `Entity.Kind` is one of: `KindContainer`, `KindVolume`, `KindNetwork`
- `loader.FromLabels()` takes a `schema.Scope` parameter
- Need to call `FromLabels()` per entity with appropriate scope

**Decision**: Iterate over snapshot entities, call `FromLabels()` for each with matching scope.

**Alternatives Rejected**:
- Merging all labels together - loses per-entity error context

---

### 5. Output Format

**Question**: How to format validation output for different use cases?

**Findings**:
- Existing snapshot command uses `json.NewEncoder(os.Stdout)` with indentation
- `--print` flag requested for showing merged config
- Need human-readable output for errors (stderr)
- Need JSON output for merged config when `--print` used (stdout)

**Decision**:
- Success: Print "Configuration valid" to stdout (or JSON with `--print`)
- Errors: Print formatted errors to stderr, exit non-zero
- `--print`: Output merged `ConfigV1` as pretty JSON

**Alternatives Rejected**:
- YAML output - adds dependency, JSON is sufficient
- Table output - JSON is more scriptable

---

### 6. File Configuration Loading (Future)

**Question**: How will config file loading work?

**Findings**:
- No file loader exists yet (assumption in spec)
- `--from file` will need this implementation
- File format TBD (YAML or JSON)
- For now: `--from file` can warn "not implemented" or error

**Decision**: Implement `--from file` as no-op/warning for v1. Full file loading is out of scope.

**Alternatives Rejected**:
- Blocking on file loader - not required for core validation flow
- Removing the flag - spec includes it, keep interface stable

---

### 7. Scope Filtering

**Question**: How to implement `--scope` flag?

**Findings**:
- `schema.Scope` has: `ScopeContainer`, `ScopeVolume`, `ScopeNetwork`, `ScopeGlobal`
- `labels.Entity.Kind` maps to scope (container→ScopeContainer, etc.)
- Global scope applies to all entities

**Decision**: Filter entities by Kind when `--scope` specified. Global labels always included.

**Alternatives Rejected**:
- Validating global labels only once - need per-entity context for error messages

---

## Summary

All research questions resolved. Key decisions:
1. Follow existing CLI patterns (Cobra, RunE, context)
2. Reuse `loader.ValidationErrors` - no new error types
3. Per-entity validation with scope mapping
4. JSON output via `--print`, human-readable errors to stderr
5. `--from file` deferred (warning for now)
