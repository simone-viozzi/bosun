# Quickstart: Config Schema Package

**Feature**: 001-config-schema
**Package**: `internal/config/schema`

## Overview

The `schema` package defines Bosun's configuration schema as Go structs with rich `bosun:` tags. It provides:

1. **Type definitions** - `Scope`, `ConfigType`, `FieldSpec`, `Spec`
2. **Tag parsing** - `ParseTags[T]()` extracts metadata from struct tags
3. **Default hydration** - `DefaultOf[T]()` creates structs with default values
4. **V1 config** - `ConfigV1` struct with representative fields

## Usage

### Import

```go
import "github.com/simone-viozzi/bosun/internal/config/schema"
```

### Parse Schema from Struct

```go
// Get the complete spec for ConfigV1
spec, err := schema.ParseTags[schema.ConfigV1]()
if err != nil {
    log.Fatalf("invalid schema: %v", err)
}

// Iterate all fields
for key, field := range spec {
    fmt.Printf("Key: %s, Scope: %s, Type: %s\n", key, field.Scope, field.Type)
}

// Get specific field
if field, ok := spec["bosun.container.stopGracePeriod"]; ok {
    fmt.Printf("Default: %s, Doc: %s\n", field.Default, field.Doc)
}
```

### Create Config with Defaults

```go
// Get a ConfigV1 populated with default values
cfg := schema.DefaultOf[schema.ConfigV1]()

fmt.Printf("StopGracePeriod: %v\n", cfg.Container.StopGracePeriod) // 30s
fmt.Printf("LogLevel: %s\n", cfg.Container.LogLevel)               // info
fmt.Printf("AutoRestart: %v\n", cfg.Container.AutoRestart)         // true
```

### Access Scope Constants

```go
// Use scope constants for type safety
if field.Scope == schema.ScopeContainer {
    // Container-specific logic
}

// Available scopes
schema.ScopeContainer // "container"
schema.ScopeVolume    // "volume"
schema.ScopeNetwork   // "network"
schema.ScopeGlobal    // "global"
```

### Access Type Constants

```go
// Use type constants
switch field.Type {
case schema.TypeString:
    // Handle string
case schema.TypeDuration:
    // Handle duration
case schema.TypeEnum:
    // Check field.Enum for allowed values
}

// Available types
schema.TypeString   // "string"
schema.TypeBool     // "bool"
schema.TypeInt      // "int"
schema.TypeDuration // "duration"
schema.TypeSize     // "size"
schema.TypeEnum     // "enum"
schema.TypeList     // "list"
```

## Defining Custom Config Structs

### Basic Field

```go
type MyConfig struct {
    Name string `bosun:"key=bosun.myapp.name,scope=global,type=string,doc='Application name'"`
}
```

### With Default Value

```go
type MyConfig struct {
    Timeout time.Duration `bosun:"key=bosun.myapp.timeout,scope=container,type=duration,default=30s"`
}
```

### Enum Field

```go
type MyConfig struct {
    Mode string `bosun:"key=bosun.myapp.mode,scope=container,type=enum,enum=dev|staging|prod,default=dev"`
}
```

### Required Field

```go
type MyConfig struct {
    APIKey string `bosun:"key=bosun.myapp.apiKey,scope=global,type=string,required=true"`
}
```

### Embedded Structs

```go
type ContainerSettings struct {
    MaxRetries int `bosun:"key=bosun.container.maxRetries,scope=container,type=int,default=3"`
}

type MyConfig struct {
    ContainerSettings // Embedded - fields are flattened into spec
}
```

## Error Handling

```go
spec, err := schema.ParseTags[MyConfig]()
if err != nil {
    // Error types:
    // - "field X: missing required tag component 'key'"
    // - "field X: invalid scope 'foo', must be container|volume|network|global"
    // - "field X: invalid type 'bar', must be string|bool|int|duration|size|enum|list"
    // - "field X: type=enum requires enum= component"
    // - "duplicate key 'bosun.x' in field Y (already defined in field Z)"
    log.Fatal(err)
}
```

## Integration with Downstream Packages

This package is the foundation for:

- **Loader** (`internal/config/loader`, #58) - Parses Docker labels using the spec
- **Merger** (`internal/config/merge`, #59) - Merges config sources using the spec
- **Doc Generator** (`internal/tools/configdoc`, #61) - Exports Markdown/JSON Schema

```go
// Example: Loader uses spec to validate labels
spec, _ := schema.ParseTags[schema.ConfigV1]()

for labelKey, labelValue := range dockerLabels {
    if field, ok := spec[labelKey]; ok {
        // Parse labelValue according to field.Type
    } else {
        // Unknown key - fail hard
    }
}
```
