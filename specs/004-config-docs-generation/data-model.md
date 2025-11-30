# Data Model: Config Documentation Generation

**Feature**: 004-config-docs-generation
**Date**: 2025-11-30

## Overview

The documentation generator reads schema metadata and transforms it into output formats. This document defines the internal data structures and transformations.

## Input Entities (from existing schema package)

### FieldSpec (existing)
Source: `internal/config/schema/types.go`

| Field | Type | Description |
|-------|------|-------------|
| Key | `string` | Docker label key (e.g., `bosun.container.stopGracePeriod`) |
| Scope | `Scope` | Where label applies (container, volume, network, global) |
| Type | `ConfigType` | Value type (string, bool, int, duration, size, enum, list) |
| Default | `string` | Default value as unparsed string |
| Enum | `[]string` | Allowed values for enum type |
| Required | `bool` | Whether field must be present |
| Doc | `string` | Human-readable description |
| Deprecated | `bool` | Whether field is deprecated |
| FieldName | `string` | Go struct field name |

### Spec (existing)
Source: `internal/config/schema/types.go`

```go
type Spec map[string]FieldSpec
```

Methods:
- `Keys() []string` - Returns sorted keys
- `Get(key) (FieldSpec, bool)` - Lookup by key
- `Scopes() map[Scope][]FieldSpec` - Group by scope

## Output Entities

### MarkdownDoc
Internal representation for Markdown generation.

| Field | Type | Description |
|-------|------|-------------|
| Title | `string` | Document title |
| GeneratedAt | `string` | Generation timestamp (for comments, not output) |
| Sections | `[]ScopeSection` | Sections grouped by scope |
| FormatDocs | `[]FormatDoc` | Special format documentation |

### ScopeSection
A section of the Markdown document for one scope.

| Field | Type | Description |
|-------|------|-------------|
| Scope | `string` | Scope name (Global, Container, Volume, Network) |
| Description | `string` | Section description |
| Fields | `[]FieldRow` | Table rows for fields in this scope |

### FieldRow
A single row in the configuration table.

| Field | Type | Description |
|-------|------|-------------|
| Key | `string` | Label key |
| Type | `string` | Human-readable type name |
| Default | `string` | Default value (or "-" if none) |
| Required | `string` | "Yes" or "No" |
| EnumValues | `string` | Pipe-separated values (or "-") |
| Description | `string` | From Doc field |
| Deprecated | `bool` | Whether to show deprecated marker |

### FormatDoc
Documentation for special value formats.

| Field | Type | Description |
|-------|------|-------------|
| Name | `string` | Format name (Duration, Byte Size, List) |
| Description | `string` | Format explanation |
| Examples | `[]string` | Example values |

### JSONSchemaDoc
JSON Schema document structure.

| Field | Type | Description |
|-------|------|-------------|
| Schema | `string` | Schema URI (`https://json-schema.org/draft/2020-12/schema`) |
| ID | `string` | Schema identifier |
| Title | `string` | Schema title |
| Description | `string` | Schema description |
| Type | `string` | Root type (`object`) |
| Properties | `map[string]PropertySchema` | Property definitions |
| Required | `[]string` | Required property keys |

### PropertySchema
JSON Schema property definition.

| Field | Type | Description |
|-------|------|-------------|
| Type | `string` | JSON Schema type |
| Description | `string` | From Doc field |
| Default | `any` | Default value (typed) |
| Enum | `[]string` | Enumerated values (if applicable) |
| Deprecated | `bool` | Deprecation flag |
| Format | `string` | Format hint (for duration, size) |
| Items | `*PropertySchema` | Array item schema (for list type) |

## Transformations

### FieldSpec → FieldRow

```
Input: FieldSpec
Output: FieldRow

Key        → Key
Type       → TypeToHumanReadable(Type)
Default    → Default or "-"
Required   → "Yes" if Required else "No"
Enum       → strings.Join(Enum, " | ") or "-"
Doc        → Description
Deprecated → Deprecated
```

Type mapping:
| ConfigType | Human-Readable |
|------------|----------------|
| TypeString | string |
| TypeBool | boolean |
| TypeInt | integer |
| TypeDuration | duration |
| TypeSize | byte size |
| TypeEnum | enum |
| TypeList | list |

### FieldSpec → PropertySchema

```
Input: FieldSpec
Output: PropertySchema

Type       → ConfigTypeToJSONType(Type)
Doc        → Description
Default    → ParseDefault(Default, Type)
Enum       → Enum (if Type == TypeEnum)
Deprecated → Deprecated
Format     → "duration" if TypeDuration, "byte-size" if TypeSize
Items      → {type: "string"} if TypeList
```

JSON type mapping:
| ConfigType | JSON Type | Additional |
|------------|-----------|------------|
| TypeString | "string" | - |
| TypeBool | "boolean" | - |
| TypeInt | "integer" | - |
| TypeDuration | "string" | format: "duration" |
| TypeSize | "string" | format: "byte-size" |
| TypeEnum | "string" | enum: [...] |
| TypeList | "array" | items: {type: "string"} |

### Spec → MarkdownDoc

```
Input: Spec
Output: MarkdownDoc

1. Get scopes map via Spec.Scopes()
2. For each scope in order [Global, Container, Volume, Network]:
   a. Create ScopeSection
   b. Sort fields by Key
   c. Transform each FieldSpec to FieldRow
3. Add FormatDocs for Duration, Byte Size, List
```

### Spec → JSONSchemaDoc

```
Input: Spec
Output: JSONSchemaDoc

1. Set schema metadata ($schema, $id, title, description)
2. For each key in Spec.Keys() (sorted):
   a. Transform FieldSpec to PropertySchema
   b. Add to Properties map
   c. If Required, add key to Required array
3. Sort Required array
```

## Validation Rules

### Input Validation
- Spec must not be empty (at least one field)
- All FieldSpec.Key must be non-empty
- All FieldSpec.Type must be valid ConfigType

### Output Validation
- Markdown must be valid GFM
- JSON Schema must validate against draft 2020-12 meta-schema
- Output must be deterministic (byte-identical on repeated runs)

## State Transitions

N/A - Generator is stateless transformation.

## Relationships

```
┌─────────────────┐
│    ConfigV1     │ (existing Go struct)
└────────┬────────┘
         │ ParseTags[ConfigV1]()
         ▼
┌─────────────────┐
│      Spec       │ (existing map[string]FieldSpec)
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐  ┌──────────────┐
│  MD   │  │  JSON Schema │
│ Doc   │  │     Doc      │
└───┬───┘  └──────┬───────┘
    │             │
    ▼             ▼
┌───────┐  ┌──────────────┐
│config │  │config.schema │
│ .md   │  │    .json     │
└───────┘  └──────────────┘
```
