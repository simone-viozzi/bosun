//go:build ignore

// This program generates configuration documentation from the ConfigV1 schema.
// It is intended to be run via go generate.
//
// Usage:
//
//	go generate ./internal/config/schema/...
//
// Or directly:
//
//	go run ./internal/tools/configdoc/cmd/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/simone-viozzi/bosun/internal/config/schema"
	"github.com/simone-viozzi/bosun/internal/tools/configdoc"
)

func main() {
	// Get the project root (where docs/ should be created)
	// When run via go generate, working directory is the package directory
	// We need to find the project root by looking for go.mod
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	// Get the schema spec (V1Spec panics on error)
	spec := schema.V1Spec()

	// Create generator with output to project root's docs/
	opts := &configdoc.Options{
		OutputDir:      filepath.Join(projectRoot, "docs"),
		MarkdownFile:   "config.md",
		JSONSchemaFile: "config.schema.json",
		SchemaID:       "https://github.com/simone-viozzi/bosun/config.schema.json",
		Title:          "Bosun Configuration Reference",
	}
	gen := configdoc.New(opts)

	// Generate documentation
	if err := gen.Generate(spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating documentation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated documentation in %s/\n", opts.OutputDir)
	fmt.Printf("  - %s\n", opts.MarkdownFile)
	fmt.Printf("  - %s\n", opts.JSONSchemaFile)
}

// findProjectRoot walks up the directory tree to find the project root
// (the directory containing go.mod).
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			return "", fmt.Errorf("could not find go.mod in any parent directory")
		}
		dir = parent
	}
}
