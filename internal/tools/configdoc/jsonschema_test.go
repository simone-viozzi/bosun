package configdoc

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

func TestGenerateJSONSchema_EmptySpec(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	_, err := gen.GenerateJSONSchema(schema.Spec{})
	if err != ErrEmptySpec {
		t.Errorf("expected ErrEmptySpec, got %v", err)
	}
}

func TestGenerateJSONSchema_ValidSchema(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecSingleField()

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Verify schema version
	if doc.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("expected draft 2020-12 schema, got %s", doc.Schema)
	}

	// Verify root type
	if doc.Type != "object" {
		t.Errorf("expected type 'object', got %s", doc.Type)
	}

	// Verify title
	if doc.Title != "Bosun Configuration Reference" {
		t.Errorf("expected default title, got %s", doc.Title)
	}
}

func TestGenerateJSONSchema_AllTypes(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecAllTypes()

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Check type mappings
	tests := []struct {
		key      string
		wantType string
		wantFmt  string
	}{
		{"bosun.type.string", "string", ""},
		{"bosun.type.bool", "boolean", ""},
		{"bosun.type.int", "integer", ""},
		{"bosun.type.duration", "string", "duration"},
		{"bosun.type.size", "string", "byte-size"},
		{"bosun.type.enum", "string", ""},
		{"bosun.type.list", "array", ""},
	}

	for _, tt := range tests {
		prop, ok := doc.Properties[tt.key]
		if !ok {
			t.Errorf("missing property %s", tt.key)
			continue
		}

		if prop.Type != tt.wantType {
			t.Errorf("%s: expected type %q, got %q", tt.key, tt.wantType, prop.Type)
		}

		if prop.Format != tt.wantFmt {
			t.Errorf("%s: expected format %q, got %q", tt.key, tt.wantFmt, prop.Format)
		}
	}

	// Check list has items schema
	listProp := doc.Properties["bosun.type.list"]
	if listProp.Items == nil {
		t.Error("list type should have items schema")
	} else if listProp.Items.Type != "string" {
		t.Errorf("list items should be string, got %s", listProp.Items.Type)
	}
}

func TestGenerateJSONSchema_EnumField(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := schema.Spec{
		"bosun.test.enum": schema.FieldSpec{
			Key:       "bosun.test.enum",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeEnum,
			Enum:      []string{"a", "b", "c"},
			Default:   "a",
			Doc:       "Test enum field",
			FieldName: "EnumField",
		},
	}

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	prop := doc.Properties["bosun.test.enum"]
	if len(prop.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(prop.Enum))
	}

	expectedEnums := []string{"a", "b", "c"}
	for i, v := range expectedEnums {
		if prop.Enum[i] != v {
			t.Errorf("enum[%d]: expected %q, got %q", i, v, prop.Enum[i])
		}
	}
}

func TestGenerateJSONSchema_RequiredFields(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := schema.Spec{
		"bosun.required.field": schema.FieldSpec{
			Key:       "bosun.required.field",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			Required:  true,
			Doc:       "Required field",
			FieldName: "RequiredField",
		},
		"bosun.optional.field": schema.FieldSpec{
			Key:       "bosun.optional.field",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			Required:  false,
			Doc:       "Optional field",
			FieldName: "OptionalField",
		},
	}

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Check required array
	if len(doc.Required) != 1 {
		t.Errorf("expected 1 required field, got %d", len(doc.Required))
	}

	if len(doc.Required) > 0 && doc.Required[0] != "bosun.required.field" {
		t.Errorf("expected 'bosun.required.field' in required, got %q", doc.Required[0])
	}
}

func TestGenerateJSONSchema_DeprecatedField(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecDeprecated()

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	prop := doc.Properties["bosun.deprecated.field"]
	if !prop.Deprecated {
		t.Error("expected deprecated flag to be true")
	}
}

func TestGenerateJSONSchema_DefaultValues(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecAllTypes()

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Check bool default is actual boolean
	boolProp := doc.Properties["bosun.type.bool"]
	if boolDefault, ok := boolProp.Default.(bool); !ok || boolDefault != true {
		t.Errorf("bool default should be true, got %v (%T)", boolProp.Default, boolProp.Default)
	}

	// Check int default is actual number (JSON unmarshals to float64)
	intProp := doc.Properties["bosun.type.int"]
	if intDefault, ok := intProp.Default.(float64); !ok || intDefault != 42 {
		t.Errorf("int default should be 42, got %v (%T)", intProp.Default, intProp.Default)
	}

	// Check string default is string
	strProp := doc.Properties["bosun.type.string"]
	if strDefault, ok := strProp.Default.(string); !ok || strDefault != "hello" {
		t.Errorf("string default should be 'hello', got %v (%T)", strProp.Default, strProp.Default)
	}

	// Check list default is array
	listProp := doc.Properties["bosun.type.list"]
	if listDefault, ok := listProp.Default.([]any); !ok {
		t.Errorf("list default should be array, got %T", listProp.Default)
	} else if len(listDefault) != 3 {
		t.Errorf("list default should have 3 elements, got %d", len(listDefault))
	}
}

func TestGenerateJSONSchema_SchemaID(t *testing.T) {
	t.Parallel()
	opts := &Options{
		SchemaID: "https://example.com/schema.json",
		Title:    "Test Schema",
	}
	gen := New(opts)
	spec := testSpecSingleField()

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc JSONSchemaDoc
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if doc.ID != "https://example.com/schema.json" {
		t.Errorf("expected custom schema ID, got %s", doc.ID)
	}

	if doc.Title != "Test Schema" {
		t.Errorf("expected custom title, got %s", doc.Title)
	}
}

func TestParseDefaultValue_AllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		defaultStr string
		configType schema.ConfigType
		wantType   string
	}{
		{"true", schema.TypeBool, "bool"},
		{"false", schema.TypeBool, "bool"},
		{"42", schema.TypeInt, "int64"},
		{"hello", schema.TypeString, "string"},
		{"30s", schema.TypeDuration, "string"},
		{"1GB", schema.TypeSize, "string"},
		{"a,b,c", schema.TypeList, "[]string"},
	}

	for _, tt := range tests {
		result := parseDefaultValue(tt.defaultStr, tt.configType)

		switch tt.wantType {
		case "bool":
			if _, ok := result.(bool); !ok {
				t.Errorf("parseDefaultValue(%q, %s) = %T, want bool", tt.defaultStr, tt.configType, result)
			}
		case "int64":
			if _, ok := result.(int64); !ok {
				t.Errorf("parseDefaultValue(%q, %s) = %T, want int64", tt.defaultStr, tt.configType, result)
			}
		case "string":
			if _, ok := result.(string); !ok {
				t.Errorf("parseDefaultValue(%q, %s) = %T, want string", tt.defaultStr, tt.configType, result)
			}
		case "[]string":
			if _, ok := result.([]string); !ok {
				t.Errorf("parseDefaultValue(%q, %s) = %T, want []string", tt.defaultStr, tt.configType, result)
			}
		}
	}
}

func TestGenerateJSONSchema_Description(t *testing.T) {
	t.Parallel()
	gen := New(nil)
	spec := testSpecSingleField()

	jsonBytes, err := gen.GenerateJSONSchema(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(jsonBytes)

	// Check description is present
	if !strings.Contains(content, "Configuration schema for Bosun Docker labels") {
		t.Error("expected schema description in output")
	}

	// Check field description
	if !strings.Contains(content, "A test field") {
		t.Error("expected field description in output")
	}
}
