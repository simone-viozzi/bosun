package configdoc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

func TestGenerator_New_DefaultOptions(t *testing.T) {
	t.Parallel()
	gen := New(nil)

	if gen.opts.OutputDir != "docs" {
		t.Errorf("expected default OutputDir 'docs', got %q", gen.opts.OutputDir)
	}
	if gen.opts.MarkdownFile != "config.md" {
		t.Errorf("expected default MarkdownFile 'config.md', got %q", gen.opts.MarkdownFile)
	}
	if gen.opts.JSONSchemaFile != "config.schema.json" {
		t.Errorf("expected default JSONSchemaFile 'config.schema.json', got %q", gen.opts.JSONSchemaFile)
	}
}

func TestGenerator_New_CustomOptions(t *testing.T) {
	t.Parallel()
	opts := &Options{
		OutputDir:      "custom/dir",
		MarkdownFile:   "custom.md",
		JSONSchemaFile: "custom.json",
		SchemaID:       "https://example.com/schema",
		Title:          "Custom Title",
	}
	gen := New(opts)

	if gen.opts.OutputDir != "custom/dir" {
		t.Errorf("expected custom OutputDir, got %q", gen.opts.OutputDir)
	}
	if gen.opts.Title != "Custom Title" {
		t.Errorf("expected custom Title, got %q", gen.opts.Title)
	}
}

func TestGenerator_Generate_EmptySpec(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	err := gen.Generate(schema.Spec{})
	if err != ErrEmptySpec {
		t.Errorf("expected ErrEmptySpec, got %v", err)
	}
}

func TestGenerator_Generate_CreatesDirectory(t *testing.T) {
	t.Parallel()

	// Create temp directory
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "nested", "docs")

	opts := &Options{
		OutputDir:      outputDir,
		MarkdownFile:   "config.md",
		JSONSchemaFile: "config.schema.json",
		Title:          "Test",
	}
	gen := New(opts)

	err := gen.Generate(testSpecSingleField())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("expected output directory to be created")
	}

	// Verify files exist
	mdPath := filepath.Join(outputDir, "config.md")
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Error("expected Markdown file to exist")
	}

	jsonPath := filepath.Join(outputDir, "config.schema.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("expected JSON Schema file to exist")
	}
}

func TestGenerator_Generate_Deterministic_Markdown(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpec()

	// Generate twice
	md1, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("first generation failed: %v", err)
	}

	md2, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("second generation failed: %v", err)
	}

	// Compare byte-for-byte
	if !bytes.Equal(md1, md2) {
		t.Error("Markdown output is not deterministic")
		t.Logf("First:\n%s", string(md1))
		t.Logf("Second:\n%s", string(md2))
	}
}

func TestGenerator_Generate_Deterministic_JSONSchema(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpec()

	// Generate twice
	json1, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("first generation failed: %v", err)
	}

	json2, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("second generation failed: %v", err)
	}

	// Compare byte-for-byte
	if !bytes.Equal(json1, json2) {
		t.Error("JSON Schema output is not deterministic")
		t.Logf("First:\n%s", string(json1))
		t.Logf("Second:\n%s", string(json2))
	}
}

func TestGenerator_Generate_Deterministic_KeyOrdering(t *testing.T) {
	t.Parallel()
	gen := New(nil)

	// Create a spec with keys that would be in different order if not sorted
	spec := schema.Spec{
		"bosun.z.last": schema.FieldSpec{
			Key:       "bosun.z.last",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			FieldName: "Last",
		},
		"bosun.a.first": schema.FieldSpec{
			Key:       "bosun.a.first",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			FieldName: "First",
		},
		"bosun.m.middle": schema.FieldSpec{
			Key:       "bosun.m.middle",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			FieldName: "Middle",
		},
	}

	// Generate multiple times
	for i := 0; i < 5; i++ {
		md, err := gen.GenerateMarkdown(spec)
		if err != nil {
			t.Fatalf("iteration %d: markdown generation failed: %v", i, err)
		}

		json, err := gen.GenerateJSONSchema(spec)
		if err != nil {
			t.Fatalf("iteration %d: json schema generation failed: %v", i, err)
		}

		// Verify keys are in alphabetical order
		mdStr := string(md)
		jsonStr := string(json)

		// In Markdown table, 'a.first' should appear before 'm.middle' which should appear before 'z.last'
		aIdx := bytes.Index(md, []byte("bosun.a.first"))
		mIdx := bytes.Index(md, []byte("bosun.m.middle"))
		zIdx := bytes.Index(md, []byte("bosun.z.last"))

		if aIdx > mIdx || mIdx > zIdx {
			t.Errorf("iteration %d: Markdown keys not in alphabetical order", i)
			t.Logf("Positions: a=%d, m=%d, z=%d", aIdx, mIdx, zIdx)
			t.Logf("Content:\n%s", mdStr)
		}

		// Same check for JSON
		aIdxJSON := bytes.Index(json, []byte("bosun.a.first"))
		mIdxJSON := bytes.Index(json, []byte("bosun.m.middle"))
		zIdxJSON := bytes.Index(json, []byte("bosun.z.last"))

		if aIdxJSON > mIdxJSON || mIdxJSON > zIdxJSON {
			t.Errorf("iteration %d: JSON keys not in alphabetical order", i)
			t.Logf("Positions: a=%d, m=%d, z=%d", aIdxJSON, mIdxJSON, zIdxJSON)
			t.Logf("Content:\n%s", jsonStr)
		}
	}
}

func TestGenerator_Generate_WritesCorrectContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	opts := &Options{
		OutputDir:      tmpDir,
		MarkdownFile:   "config.md",
		JSONSchemaFile: "config.schema.json",
		Title:          "Test Schema",
		SchemaID:       "https://test.example/schema.json",
	}
	gen := New(opts)

	err := gen.Generate(testSpecSingleField())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read and verify Markdown
	mdContent, err := os.ReadFile(filepath.Join(tmpDir, "config.md"))
	if err != nil {
		t.Fatalf("failed to read Markdown file: %v", err)
	}
	if !bytes.Contains(mdContent, []byte("# Test Schema")) {
		t.Error("Markdown should contain custom title")
	}

	// Read and verify JSON Schema
	jsonContent, err := os.ReadFile(filepath.Join(tmpDir, "config.schema.json"))
	if err != nil {
		t.Fatalf("failed to read JSON Schema file: %v", err)
	}
	if !bytes.Contains(jsonContent, []byte("https://test.example/schema.json")) {
		t.Error("JSON Schema should contain custom schema ID")
	}
}

func TestErrOutputDir_Error(t *testing.T) {
	t.Parallel()
	err := &ErrOutputDir{
		Path: "/test/path",
		Err:  os.ErrPermission,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if !bytes.Contains([]byte(msg), []byte("/test/path")) {
		t.Error("error message should contain path")
	}
}

func TestErrOutputDir_Unwrap(t *testing.T) {
	t.Parallel()
	innerErr := os.ErrPermission
	err := &ErrOutputDir{
		Path: "/test/path",
		Err:  innerErr,
	}

	if err.Unwrap() != innerErr {
		t.Error("Unwrap should return inner error")
	}
}
