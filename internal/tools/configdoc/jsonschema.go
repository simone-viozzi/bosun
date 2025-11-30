package configdoc

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// JSONSchemaDoc represents a JSON Schema document.
type JSONSchemaDoc struct {
	Schema      string                    `json:"$schema"`
	ID          string                    `json:"$id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Type        string                    `json:"type"`
	Properties  map[string]PropertySchema `json:"properties"`
	Required    []string                  `json:"required,omitempty"`
}

// PropertySchema represents a JSON Schema property definition.
type PropertySchema struct {
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Default     any             `json:"default,omitempty"`
	Enum        []string        `json:"enum,omitempty"`
	Deprecated  bool            `json:"deprecated,omitempty"`
	Format      string          `json:"format,omitempty"`
	Items       *PropertySchema `json:"items,omitempty"`
}

// fieldSpecToProperty converts a FieldSpec to a PropertySchema for JSON Schema.
func fieldSpecToProperty(f schema.FieldSpec) PropertySchema {
	prop := PropertySchema{
		Type:        typeToJSONSchemaType(f.Type),
		Description: f.Doc,
		Deprecated:  f.Deprecated,
	}

	// Add format for special types
	if format := typeToJSONSchemaFormat(f.Type); format != "" {
		prop.Format = format
	}

	// Add enum values
	if f.Type == schema.TypeEnum && len(f.Enum) > 0 {
		prop.Enum = f.Enum
	}

	// Add items schema for list type
	if f.Type == schema.TypeList {
		prop.Items = &PropertySchema{Type: "string"}
	}

	// Parse and set default value with correct type
	if f.Default != "" {
		prop.Default = parseDefaultValue(f.Default, f.Type)
	}

	return prop
}

// parseDefaultValue parses a default value string into the appropriate Go type.
func parseDefaultValue(defaultStr string, configType schema.ConfigType) any {
	switch configType {
	case schema.TypeBool:
		if defaultStr == "true" {
			return true
		}
		return false
	case schema.TypeInt:
		if v, err := strconv.ParseInt(defaultStr, 10, 64); err == nil {
			return v
		}
		return defaultStr
	case schema.TypeDuration:
		// Keep as string for JSON Schema (human-readable)
		return defaultStr
	case schema.TypeSize:
		// Keep as string for JSON Schema (human-readable)
		return defaultStr
	case schema.TypeList:
		// Parse CSV into array
		if defaultStr != "" {
			return strings.Split(defaultStr, ",")
		}
		return []string{}
	default:
		return defaultStr
	}
}

// generateJSONSchema produces JSON Schema from the spec.
func generateJSONSchema(spec schema.Spec, title, schemaID string) ([]byte, error) {
	// Get sorted keys for deterministic output
	keys := spec.Keys()

	// Build properties map
	properties := make(map[string]PropertySchema)
	var required []string

	for _, key := range keys {
		field, _ := spec.Get(key)
		properties[key] = fieldSpecToProperty(field)

		if field.Required {
			required = append(required, key)
		}
	}

	// Sort required array for determinism
	sort.Strings(required)

	doc := JSONSchemaDoc{
		Schema:      "https://json-schema.org/draft/2020-12/schema",
		ID:          schemaID,
		Title:       title,
		Description: "Configuration schema for Bosun Docker labels",
		Type:        "object",
		Properties:  properties,
		Required:    required,
	}

	// Marshal with indentation for readability
	return json.MarshalIndent(doc, "", "  ")
}

// Ensure time package is used (for potential future use)
var _ = time.Now
