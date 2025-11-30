// Package configdoc generates documentation from the Bosun configuration schema.
//
// It produces two output formats:
//   - Markdown documentation (config.md) for human readers
//   - JSON Schema (config.schema.json) for tooling and validation
//
// The generator reads schema metadata from the config/schema package using
// ParseTags[ConfigV1]() and transforms it into deterministic, version-control
// friendly output files.
//
// # Usage
//
// The package can be invoked via go generate or make docs:
//
//	//go:generate go run ../../tools/configdoc/cmd/main.go
//
// Or programmatically:
//
//	spec, _ := schema.V1Spec()
//	gen := configdoc.New(nil) // uses default options
//	gen.Generate(spec)
//
// # Output Files
//
// Generated files are placed in the docs/ directory at the project root:
//   - docs/config.md: Markdown table with all config keys, grouped by scope
//   - docs/config.schema.json: JSON Schema draft 2020-12 for validation
//
// # Determinism
//
// All output is deterministic - running the generator twice with the same
// input produces byte-for-byte identical output. This is achieved by:
//   - Using sorted keys from Spec.Keys()
//   - Consistent scope ordering (Global, Container, Volume, Network)
//   - Sorted JSON object properties and required arrays
package configdoc
