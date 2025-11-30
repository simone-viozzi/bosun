package configdoc

import (
	"os"
	"path/filepath"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

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

// DefaultOptions provides sensible defaults for the generator.
var DefaultOptions = Options{
	OutputDir:      "docs",
	MarkdownFile:   "config.md",
	JSONSchemaFile: "config.schema.json",
	SchemaID:       "https://github.com/simone-viozzi/bosun/config.schema.json",
	Title:          "Bosun Configuration Reference",
}

// Generator produces documentation from a schema Spec.
type Generator struct {
	opts Options
}

// New creates a new Generator with the given options.
// If opts is nil, default options are used.
func New(opts *Options) *Generator {
	if opts == nil {
		opts = &DefaultOptions
	}
	return &Generator{opts: *opts}
}

// Generate produces documentation files from the given spec.
// It writes to OutputDir/MarkdownFile and OutputDir/JSONSchemaFile.
// Returns an error if generation or file writing fails.
func (g *Generator) Generate(spec schema.Spec) error {
	if len(spec) == 0 {
		return ErrEmptySpec
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.opts.OutputDir, 0o755); err != nil {
		return &ErrOutputDir{Path: g.opts.OutputDir, Err: err}
	}

	// Generate Markdown
	mdContent, err := g.GenerateMarkdown(spec)
	if err != nil {
		return err
	}

	mdPath := filepath.Join(g.opts.OutputDir, g.opts.MarkdownFile)
	if err := os.WriteFile(mdPath, mdContent, 0o644); err != nil {
		return &ErrOutputDir{Path: mdPath, Err: err}
	}

	// Generate JSON Schema
	jsonContent, err := g.GenerateJSONSchema(spec)
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(g.opts.OutputDir, g.opts.JSONSchemaFile)
	if err := os.WriteFile(jsonPath, jsonContent, 0o644); err != nil {
		return &ErrOutputDir{Path: jsonPath, Err: err}
	}

	return nil
}

// GenerateMarkdown produces Markdown documentation from the spec.
// Returns the Markdown content as a byte slice.
func (g *Generator) GenerateMarkdown(spec schema.Spec) ([]byte, error) {
	if len(spec) == 0 {
		return nil, ErrEmptySpec
	}
	return generateMarkdown(spec, g.opts.Title)
}

// GenerateJSONSchema produces JSON Schema from the spec.
// Returns the JSON Schema content as a byte slice.
func (g *Generator) GenerateJSONSchema(spec schema.Spec) ([]byte, error) {
	if len(spec) == 0 {
		return nil, ErrEmptySpec
	}
	return generateJSONSchema(spec, g.opts.Title, g.opts.SchemaID)
}
