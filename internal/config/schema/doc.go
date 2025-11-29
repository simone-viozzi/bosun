// Package schema defines Bosun's configuration schema as Go structs with rich
// `bosun:` struct tags. It serves as the single source of truth for:
//
//   - Parsing Docker labels into strongly-typed configuration
//   - Exporting JSON Schema and Markdown documentation
//   - Validating configuration values
//
// # Tag Format
//
// The package uses a custom struct tag format:
//
//	bosun:"key=<label_key>,scope=<scope>,type=<type>[,default=<value>][,enum=<a|b|c>][,required=true][,doc='<description>'][,deprecated=true]"
//
// Required components:
//   - key: The Docker label key (must start with "bosun.")
//   - scope: One of "container", "volume", "network", "global"
//   - type: One of "string", "bool", "int", "duration", "size", "enum", "list"
//
// Optional components:
//   - default: Default value as string
//   - enum: Pipe-separated allowed values (required if type=enum)
//   - required: Mark field as required (default: false)
//   - doc: Human-readable description (single quotes allow commas)
//   - deprecated: Mark field as deprecated (default: false)
//
// # Usage
//
// Parse schema from a struct type:
//
//	spec, err := schema.ParseTags[schema.ConfigV1]()
//	if err != nil {
//	    log.Fatalf("invalid schema: %v", err)
//	}
//
// Create config with defaults:
//
//	cfg, err := schema.DefaultOf[schema.ConfigV1]()
//	if err != nil {
//	    log.Fatalf("failed to hydrate defaults: %v", err)
//	}
//
// # Related Issues
//
//   - #57: This package (code-first schema)
//   - #58: Label parser (uses this schema)
//   - #59: Config merger (uses this schema)
//   - #61: Doc generator (uses this schema)
package schema
