# Quickstart: Label Parser and Source Merger

**Feature Branch**: `002-label-parser-merger`
**Date**: 2025-11-29

## Overview

This feature adds two packages to the Bosun config system:

1. **`internal/config/loader`** - Parses Docker labels into typed ConfigV1
2. **`internal/config/merge`** - Merges configs from multiple sources with precedence

## Quick Usage

### Parse Labels into Config

```go
package main

import (
    "fmt"
    "github.com/simone-viozzi/bosun/internal/config/loader"
    "github.com/simone-viozzi/bosun/internal/config/schema"
)

func main() {
    // Get the schema spec
    spec, err := schema.V1Spec()
    if err != nil {
        panic(err)
    }

    // Docker labels from a container
    labels := map[string]string{
        "bosun.container.stopGracePeriod": "30s",
        "bosun.container.autoRestart":     "true",
        "bosun.container.logLevel":        "debug",
    }

    // Parse labels for a container entity
    cfg, err := loader.FromLabels(spec, labels, schema.ScopeContainer)
    if err != nil {
        // err contains all validation errors
        fmt.Printf("Validation errors:\n%v\n", err)
        return
    }

    fmt.Printf("StopGracePeriod: %v\n", cfg.Container.StopGracePeriod)
    fmt.Printf("AutoRestart: %v\n", cfg.Container.AutoRestart)
    fmt.Printf("LogLevel: %v\n", cfg.Container.LogLevel)
}
```

### Merge Multiple Sources

```go
package main

import (
    "fmt"
    "github.com/simone-viozzi/bosun/internal/config/loader"
    "github.com/simone-viozzi/bosun/internal/config/merge"
    "github.com/simone-viozzi/bosun/internal/config/schema"
)

func main() {
    spec, _ := schema.V1Spec()
    defaults, _ := schema.V1Defaults()

    // File config (from YAML/JSON)
    fileConfig := schema.ConfigV1{
        Container: schema.ContainerConfig{
            StopGracePeriod: 20 * time.Second,
        },
    }

    // Label config (from Docker)
    labels := map[string]string{
        "bosun.container.stopGracePeriod": "30s",
    }
    labelConfig, _ := loader.FromLabels(spec, labels, schema.ScopeContainer)

    // Merge: defaults < file < labels
    merged, err := merge.Merge(spec, defaults, &fileConfig, nil, &labelConfig, merge.Options{})
    if err != nil {
        panic(err)
    }

    // Result: 30s (labels win over file's 20s)
    fmt.Printf("Final StopGracePeriod: %v\n", merged.Container.StopGracePeriod)
}
```

## Error Handling

```go
cfg, err := loader.FromLabels(spec, labels, schema.ScopeContainer)
if err != nil {
    if verrs, ok := err.(loader.ValidationErrors); ok {
        for _, verr := range verrs {
            fmt.Printf("Error on %s: %s\n", verr.Key, verr.Message)
        }
    }
    return
}
```

## Validation Error Types

| Error Type | Example |
|------------|---------|
| Unknown key | `unknown key: bosun.container.typoedKey` |
| Scope mismatch | `key 'bosun.container.stopGracePeriod' not allowed on scope 'volume'` |
| Invalid duration | `invalid duration for key '...': time: invalid duration "invalid"` |
| Invalid bool | `invalid bool for key '...': strconv.ParseBool: parsing "maybe": invalid syntax` |
| Invalid enum | `invalid enum value 'verbose' for key '...': must be one of [debug, info, warn, error]` |
| Required missing | `required key 'bosun.global.instance' not provided` |

## Package Locations

```
internal/config/
├── schema/           # Existing: ConfigV1, Spec, types
├── loader/           # NEW: FromLabels(), ValidationErrors
│   ├── loader.go
│   ├── loader_test.go
│   ├── errors.go
│   └── errors_test.go
└── merge/            # NEW: Merge(), Options
    ├── merge.go
    └── merge_test.go
```

## Running Tests

```bash
# Unit tests for new packages
go test ./internal/config/loader/...
go test ./internal/config/merge/...

# All config tests
go test ./internal/config/...
```
