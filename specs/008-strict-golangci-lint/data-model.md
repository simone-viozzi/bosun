# Data Model: Strict golangci-lint Checks

**Feature**: 008-strict-golangci-lint
**Date**: 2025-12-27

## Overview

This feature involves code quality improvements and configuration changes. There are no domain data models introduced or modified.

## Configuration Model

The primary "data model" is the golangci-lint configuration file (`.golangci.yml`).

### Current State (Before)

```yaml
linters:
  enable:
    - bodyclose
    - nilerr
    # - prealloc    # DISABLED
    # - unparam     # DISABLED
  disable:
    - errcheck     # DISABLED

  settings:
    govet:
      disable:
        - shadow
        - fieldalignment
        - unusedwrite  # DISABLED

    staticcheck:
      checks:
        - "all"
        - "-ST1000"    # EXCLUDED
        - "-ST1016"    # EXCLUDED
        - "-ST1022"    # EXCLUDED
```

### Target State (After)

```yaml
linters:
  enable:
    - bodyclose
    - nilerr
    - prealloc     # ENABLED
    - unparam      # ENABLED
    - errcheck     # ENABLED (moved from disable)

  settings:
    govet:
      disable:
        - shadow         # Keep disabled (out of scope)
        - fieldalignment # Keep disabled (out of scope)
        # unusedwrite removed - ENABLED

    staticcheck:
      checks:
        - "all"
        # No exclusions - all ST* rules ENABLED
```

## Package Documentation Model

Each package requiring a doc comment will follow this structure:

| Package | File | Description |
|---------|------|-------------|
| `dockerlabels` | `internal/adapters/dockerlabels/doc.go` | Docker label discovery adapter |
| `app` | `internal/app/doc.go` | Application orchestration layer |
| `labels` | `internal/domain/labels/doc.go` | Label domain types |
| `ports` | `internal/ports/doc.go` | Port interfaces for hexagonal architecture |

### doc.go Template

```go
// Package {name} provides {brief description}.
//
// {Additional context about purpose, usage, or design decisions.}
package {name}
```

## No Domain Entity Changes

This feature does not introduce or modify any domain entities:
- No new structs
- No new interfaces
- No database changes
- No API changes

All changes are:
1. Code quality fixes (pre-allocation, parameter cleanup)
2. Documentation additions (package comments, constant comments)
3. Style consistency (receiver names)
4. Test improvements (unused write fixes)
5. Configuration enablement
