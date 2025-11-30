# Contract: Documentation Generator

**Package**: `internal/tools/configdoc`
**Date**: 2025-11-30

## Overview

The `configdoc` package generates documentation files from the config schema. It provides a clean API for generating both Markdown and JSON Schema outputs.

## Public API

### Types

```go
// Generator produces documentation from a schema Spec.
type Generator struct {
    // contains filtered or unexported fields
}

// Options configures the generator behavior.
type Options struct {
    // OutputDir is the directory for generated files.
    // Default: "docs"
    OutputDir string

    // MarkdownFile is the name of the Markdown output file.
    // Default: "config.md"
    MarkdownFile string

    // JSONSchemaFile is the name of the JSON Schema output file.
    // Default: "config.schema.json"
    JSONSchemaFile string

    // SchemaID is the $id for the JSON Schema.
    // Default: "https://github.com/simone-viozzi/bosun/config.schema.json"
    SchemaID string

    // Title is the document title.
    // Default: "Bosun Configuration Reference"
    Title string
}
```

### Functions

```go
// New creates a new Generator with the given options.
// If opts is nil, default options are used.
func New(opts *Options) *Generator

// Generate produces documentation files from the given spec.
// It writes to OutputDir/MarkdownFile and OutputDir/JSONSchemaFile.
// Returns an error if generation or file writing fails.
func (g *Generator) Generate(spec schema.Spec) error

// GenerateMarkdown produces Markdown documentation from the spec.
// Returns the Markdown content as a byte slice.
func (g *Generator) GenerateMarkdown(spec schema.Spec) ([]byte, error)

// GenerateJSONSchema produces JSON Schema from the spec.
// Returns the JSON Schema content as a byte slice.
func (g *Generator) GenerateJSONSchema(spec schema.Spec) ([]byte, error)
```

### Default Options

```go
var DefaultOptions = Options{
    OutputDir:      "docs",
    MarkdownFile:   "config.md",
    JSONSchemaFile: "config.schema.json",
    SchemaID:       "https://github.com/simone-viozzi/bosun/config.schema.json",
    Title:          "Bosun Configuration Reference",
}
```

## Behavior Contracts

### Generate()

**Preconditions**:
- `spec` must not be empty (len > 0)
- `OutputDir` must be writable (or creatable)

**Postconditions**:
- `OutputDir/MarkdownFile` exists and contains valid Markdown
- `OutputDir/JSONSchemaFile` exists and contains valid JSON Schema
- Both files have deterministic content (identical on repeated calls with same input)

**Error Conditions**:
- Returns error if `spec` is empty
- Returns error if output directory cannot be created
- Returns error if files cannot be written

### GenerateMarkdown()

**Preconditions**:
- `spec` must not be empty

**Postconditions**:
- Returns valid GitHub-flavored Markdown
- Fields are grouped by scope (Global, Container, Volume, Network)
- Within each scope, fields are sorted alphabetically by key
- Includes format documentation for special types

**Output Format**:
```markdown
# {Title}

> Auto-generated from ConfigV1 schema. Do not edit manually.

## Global Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| ... | ... | ... | ... | ... | ... |

## Container Configuration
...

## Volume Configuration
...

## Network Configuration
...

## Value Formats

### Duration
...

### Byte Size
...

### List
...
```

### GenerateJSONSchema()

**Preconditions**:
- `spec` must not be empty

**Postconditions**:
- Returns valid JSON Schema draft 2020-12
- Schema `$schema` is `https://json-schema.org/draft/2020-12/schema`
- Properties are in deterministic order (sorted by key)
- Required array is sorted

**Output Format**:
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "{SchemaID}",
  "title": "{Title}",
  "description": "Configuration schema for Bosun Docker labels",
  "type": "object",
  "properties": {
    "bosun.container.autoRestart": {
      "type": "boolean",
      "description": "...",
      "default": true
    },
    ...
  },
  "required": ["..."]
}
```

## Error Types

```go
// ErrEmptySpec is returned when the input spec has no fields.
var ErrEmptySpec = errors.New("configdoc: spec is empty")

// ErrOutputDir is returned when the output directory cannot be created.
type ErrOutputDir struct {
    Path string
    Err  error
}

func (e *ErrOutputDir) Error() string
func (e *ErrOutputDir) Unwrap() error
```

## Usage Example

```go
package main

import (
    "log"

    "github.com/simone-viozzi/bosun/internal/config/schema"
    "github.com/simone-viozzi/bosun/internal/tools/configdoc"
)

func main() {
    // Get the schema spec
    spec, err := schema.V1Spec()
    if err != nil {
        log.Fatal(err)
    }

    // Create generator with defaults
    gen := configdoc.New(nil)

    // Generate documentation
    if err := gen.Generate(spec); err != nil {
        log.Fatal(err)
    }

    log.Println("Documentation generated successfully")
}
```

## Integration Points

### Makefile Target

```makefile
.PHONY: docs
docs:
	go generate ./internal/config/schema/...
```

### go:generate Directive

Location: `internal/config/schema/config_v1.go`

```go
//go:generate go run ../../tools/configdoc/cmd/main.go
```

### CI Verification

```yaml
# In .github/workflows/ci.yml
- name: Verify docs are up-to-date
  run: |
    make docs
    git diff --exit-code docs/
```

## Test Requirements

1. **Unit Tests** (`*_test.go`):
   - `TestGenerateMarkdown_Empty` - Error on empty spec
   - `TestGenerateMarkdown_SingleField` - Basic field rendering
   - `TestGenerateMarkdown_AllTypes` - All ConfigType values
   - `TestGenerateMarkdown_Scopes` - Grouping by scope
   - `TestGenerateMarkdown_Deprecated` - Deprecated field marking
   - `TestGenerateJSONSchema_Empty` - Error on empty spec
   - `TestGenerateJSONSchema_ValidSchema` - Schema validates
   - `TestGenerateJSONSchema_AllTypes` - Type mapping
   - `TestGenerateJSONSchema_Required` - Required fields
   - `TestGenerate_Deterministic` - Same input = identical output

2. **Golden File Tests**:
   - `testdata/golden.md` - Expected Markdown output
   - `testdata/golden.schema.json` - Expected JSON Schema output
