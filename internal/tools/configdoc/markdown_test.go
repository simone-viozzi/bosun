package configdoc

import (
	"strings"
	"testing"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

func TestGenerateMarkdown_EmptySpec(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	_, err := gen.GenerateMarkdown(schema.Spec{})
	if err != ErrEmptySpec {
		t.Errorf("expected ErrEmptySpec, got %v", err)
	}
}

func TestGenerateMarkdown_SingleField(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecSingleField()

	md, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(md)

	// Check title
	if !strings.Contains(content, "# Bosun Configuration Reference") {
		t.Error("expected title in output")
	}

	// Check auto-generated notice
	if !strings.Contains(content, "Auto-generated from ConfigV1 schema") {
		t.Error("expected auto-generated notice in output")
	}

	// Check field key
	if !strings.Contains(content, "`bosun.test.field`") {
		t.Error("expected field key in output")
	}

	// Check field type
	if !strings.Contains(content, "string") {
		t.Error("expected field type in output")
	}

	// Check default value
	if !strings.Contains(content, "default") {
		t.Error("expected default value in output")
	}

	// Check description
	if !strings.Contains(content, "A test field") {
		t.Error("expected description in output")
	}
}

func TestGenerateMarkdown_AllTypes(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecAllTypes()

	md, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(md)

	// Check all type names appear
	typeNames := []string{"string", "boolean", "integer", "duration", "byte size", "enum", "list"}
	for _, typeName := range typeNames {
		if !strings.Contains(content, typeName) {
			t.Errorf("expected type %q in output", typeName)
		}
	}

	// Check enum values appear
	if !strings.Contains(content, "a \\| b \\| c") {
		t.Error("expected enum values in output")
	}
}

func TestGenerateMarkdown_ScopeGrouping(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpec()

	md, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(md)

	// Check all scope sections appear
	scopes := []string{"Global Configuration", "Container Configuration", "Volume Configuration", "Network Configuration"}
	for _, scope := range scopes {
		if !strings.Contains(content, "## "+scope) {
			t.Errorf("expected scope section %q in output", scope)
		}
	}

	// Check scope order (Global should come before Container)
	globalIdx := strings.Index(content, "## Global Configuration")
	containerIdx := strings.Index(content, "## Container Configuration")
	volumeIdx := strings.Index(content, "## Volume Configuration")
	networkIdx := strings.Index(content, "## Network Configuration")

	if globalIdx > containerIdx {
		t.Error("Global should come before Container")
	}
	if containerIdx > volumeIdx {
		t.Error("Container should come before Volume")
	}
	if volumeIdx > networkIdx {
		t.Error("Volume should come before Network")
	}
}

func TestGenerateMarkdown_DeprecatedField(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecDeprecated()

	md, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(md)

	// Check deprecated marker
	if !strings.Contains(content, "*(deprecated)*") {
		t.Error("expected deprecated marker in output")
	}

	// Check strikethrough
	if !strings.Contains(content, "~~") {
		t.Error("expected strikethrough markup for deprecated field")
	}
}

func TestGenerateMarkdown_TableHeaders(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecSingleField()

	md, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(md)

	// Check all table headers
	headers := []string{"Key", "Type", "Default", "Required", "Allowed Values", "Description"}
	for _, header := range headers {
		if !strings.Contains(content, header) {
			t.Errorf("expected table header %q in output", header)
		}
	}
}

func TestGenerateMarkdown_FormatDocumentation(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecSingleField()

	md, err := gen.GenerateMarkdown(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(md)

	// Check format documentation section
	if !strings.Contains(content, "## Value Formats") {
		t.Error("expected Value Formats section in output")
	}

	// Check duration format
	if !strings.Contains(content, "### Duration") {
		t.Error("expected Duration format section")
	}
	if !strings.Contains(content, "30s") {
		t.Error("expected duration examples")
	}

	// Check byte size format
	if !strings.Contains(content, "### Byte Size") {
		t.Error("expected Byte Size format section")
	}
	if !strings.Contains(content, "100MB") {
		t.Error("expected byte size examples")
	}

	// Check list format
	if !strings.Contains(content, "### List") {
		t.Error("expected List format section")
	}
}

func TestFieldSpecToRow_DefaultEmpty(t *testing.T) {
	t.Parallel()
	f := schema.FieldSpec{
		Key:   "test",
		Type:  schema.TypeString,
		Scope: schema.ScopeContainer,
	}

	row := fieldSpecToRow(f)

	if row.Default != "-" {
		t.Errorf("expected '-' for empty default, got %q", row.Default)
	}
}

func TestFieldSpecToRow_EnumEmpty(t *testing.T) {
	t.Parallel()
	f := schema.FieldSpec{
		Key:   "test",
		Type:  schema.TypeString,
		Scope: schema.ScopeContainer,
	}

	row := fieldSpecToRow(f)

	if row.EnumValues != "-" {
		t.Errorf("expected '-' for empty enum, got %q", row.EnumValues)
	}
}

func TestFieldSpecToRow_Required(t *testing.T) {
	t.Parallel()

	tests := []struct {
		required bool
		want     string
	}{
		{true, "Yes"},
		{false, "No"},
	}

	for _, tt := range tests {
		f := schema.FieldSpec{
			Key:      "test",
			Type:     schema.TypeString,
			Scope:    schema.ScopeContainer,
			Required: tt.required,
		}

		row := fieldSpecToRow(f)

		if row.Required != tt.want {
			t.Errorf("Required=%v: expected %q, got %q", tt.required, tt.want, row.Required)
		}
	}
}
