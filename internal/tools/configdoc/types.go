package configdoc

import "github.com/simone-viozzi/bosun/internal/config/schema"

// scopeOrder defines the canonical order for scope sections in documentation.
// This ensures deterministic output.
var scopeOrder = []schema.Scope{
	schema.ScopeGlobal,
	schema.ScopeContainer,
	schema.ScopeVolume,
	schema.ScopeNetwork,
}

// scopeDisplayName returns the human-readable name for a scope.
func scopeDisplayName(s schema.Scope) string {
	switch s {
	case schema.ScopeGlobal:
		return "Global"
	case schema.ScopeContainer:
		return "Container"
	case schema.ScopeVolume:
		return "Volume"
	case schema.ScopeNetwork:
		return "Network"
	default:
		return string(s)
	}
}

// typeDisplayName returns the human-readable name for a config type.
func typeDisplayName(t schema.ConfigType) string {
	switch t {
	case schema.TypeString:
		return "string"
	case schema.TypeBool:
		return "boolean"
	case schema.TypeInt:
		return "integer"
	case schema.TypeDuration:
		return "duration"
	case schema.TypeSize:
		return "byte size"
	case schema.TypeEnum:
		return "enum"
	case schema.TypeList:
		return "list"
	default:
		return string(t)
	}
}

// typeToJSONSchemaType maps ConfigType to JSON Schema type.
func typeToJSONSchemaType(t schema.ConfigType) string {
	switch t {
	case schema.TypeString, schema.TypeDuration, schema.TypeSize, schema.TypeEnum:
		return "string"
	case schema.TypeBool:
		return "boolean"
	case schema.TypeInt:
		return "integer"
	case schema.TypeList:
		return "array"
	default:
		return "string"
	}
}

// typeToJSONSchemaFormat returns the format hint for special types.
// Returns empty string if no format applies.
func typeToJSONSchemaFormat(t schema.ConfigType) string {
	switch t {
	case schema.TypeDuration:
		return "duration"
	case schema.TypeSize:
		return "byte-size"
	default:
		return ""
	}
}
