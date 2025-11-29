package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// tagName is the struct tag key used for bosun configuration.
const tagName = "bosun"

// parseTagValue parses a bosun tag value into key-value pairs.
// It handles quoted strings for the doc field (e.g., doc='Description with, commas').
//
// Example input: "key=bosun.container.stop,scope=container,type=duration,doc='Grace period'"
// Example output: map[string]string{"key": "bosun.container.stop", "scope": "container", ...}
func parseTagValue(tagValue string) (map[string]string, error) {
	result := make(map[string]string)

	if tagValue == "" {
		return result, nil
	}

	// State machine to parse comma-separated key=value pairs,
	// respecting single quotes for values.
	var currentKey strings.Builder
	var currentValue strings.Builder
	inValue := false
	inQuotes := false
	escaped := false

	for i := 0; i < len(tagValue); i++ {
		c := tagValue[i]

		if escaped {
			if inValue {
				currentValue.WriteByte(c)
			} else {
				currentKey.WriteByte(c)
			}
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == '\'' {
			inQuotes = !inQuotes
			continue
		}

		if c == '=' && !inValue && !inQuotes {
			inValue = true
			continue
		}

		if c == ',' && !inQuotes {
			// End of current pair
			key := strings.TrimSpace(currentKey.String())
			value := strings.TrimSpace(currentValue.String())
			if key != "" {
				result[key] = value
			}
			currentKey.Reset()
			currentValue.Reset()
			inValue = false
			continue
		}

		if inValue {
			currentValue.WriteByte(c)
		} else {
			currentKey.WriteByte(c)
		}
	}

	// Handle the last pair
	key := strings.TrimSpace(currentKey.String())
	value := strings.TrimSpace(currentValue.String())
	if key != "" {
		result[key] = value
	}

	return result, nil
}

// parseFieldSpec creates a FieldSpec from parsed tag parts.
// It validates required fields and enum constraints.
func parseFieldSpec(fieldName string, goType reflect.Type, tagParts map[string]string) (FieldSpec, error) {
	fs := FieldSpec{
		FieldName: fieldName,
	}

	// Required: key
	key, ok := tagParts["key"]
	if !ok || key == "" {
		return fs, fmt.Errorf("field %s: missing required tag component 'key'", fieldName)
	}
	if !strings.HasPrefix(key, "bosun.") {
		return fs, fmt.Errorf("field %s: key %q must start with 'bosun.'", fieldName, key)
	}
	fs.Key = key

	// Required: scope
	scopeStr, ok := tagParts["scope"]
	if !ok || scopeStr == "" {
		return fs, fmt.Errorf("field %s: missing required tag component 'scope'", fieldName)
	}
	if !IsValidScope(scopeStr) {
		return fs, fmt.Errorf("field %s: invalid scope %q, must be container|volume|network|global", fieldName, scopeStr)
	}
	fs.Scope = Scope(scopeStr)

	// Required: type
	typeStr, ok := tagParts["type"]
	if !ok || typeStr == "" {
		return fs, fmt.Errorf("field %s: missing required tag component 'type'", fieldName)
	}
	if !IsValidConfigType(typeStr) {
		return fs, fmt.Errorf("field %s: invalid type %q, must be string|bool|int|duration|size|enum|list", fieldName, typeStr)
	}
	fs.Type = ConfigType(typeStr)

	// Validate Go type matches config type
	if err := validateGoType(fieldName, goType, fs.Type); err != nil {
		return fs, err
	}

	// Optional: default
	if defaultVal, ok := tagParts["default"]; ok {
		fs.Default = defaultVal
	}

	// Optional: enum (required if type=enum)
	if enumStr, ok := tagParts["enum"]; ok && enumStr != "" {
		fs.Enum = strings.Split(enumStr, "|")
	}
	if fs.Type == TypeEnum && len(fs.Enum) == 0 {
		return fs, fmt.Errorf("field %s: type=enum requires enum= component with allowed values", fieldName)
	}

	// Optional: required
	if reqStr, ok := tagParts["required"]; ok {
		fs.Required = reqStr == "true"
	}

	// Optional: doc
	if doc, ok := tagParts["doc"]; ok {
		fs.Doc = doc
	}

	// Optional: deprecated
	if depStr, ok := tagParts["deprecated"]; ok {
		fs.Deprecated = depStr == "true"
	}

	return fs, nil
}

// validateGoType checks that the Go field type matches the declared config type.
func validateGoType(fieldName string, goType reflect.Type, configType ConfigType) error {
	// Handle pointer types
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	switch configType {
	case TypeString, TypeEnum:
		if goType.Kind() != reflect.String {
			return fmt.Errorf("field %s: type=%s requires Go type string, got %s", fieldName, configType, goType.Kind())
		}
	case TypeBool:
		if goType.Kind() != reflect.Bool {
			return fmt.Errorf("field %s: type=bool requires Go type bool, got %s", fieldName, goType.Kind())
		}
	case TypeInt:
		if goType.Kind() != reflect.Int && goType.Kind() != reflect.Int64 && goType.Kind() != reflect.Int32 {
			return fmt.Errorf("field %s: type=int requires Go type int/int32/int64, got %s", fieldName, goType.Kind())
		}
	case TypeDuration:
		// time.Duration is int64 under the hood, but we check the type name
		if goType.String() != "time.Duration" {
			return fmt.Errorf("field %s: type=duration requires Go type time.Duration, got %s", fieldName, goType.String())
		}
	case TypeSize:
		if goType.Kind() != reflect.Int64 {
			return fmt.Errorf("field %s: type=size requires Go type int64, got %s", fieldName, goType.Kind())
		}
	case TypeList:
		if goType.Kind() != reflect.Slice || goType.Elem().Kind() != reflect.String {
			return fmt.Errorf("field %s: type=list requires Go type []string, got %s", fieldName, goType.String())
		}
	}

	return nil
}

// ParseTags extracts the schema specification from a struct type T.
// It processes all fields with bosun: tags, including embedded structs.
//
// Returns an error if:
//   - A required tag component is missing (key, scope, type)
//   - A scope or type value is invalid
//   - Duplicate keys are found
//   - type=enum is used without enum= values
//   - The Go field type doesn't match the declared config type
func ParseTags[T any]() (Spec, error) {
	var zero T
	t := reflect.TypeOf(zero)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ParseTags requires a struct type, got %s", t.Kind())
	}

	spec := make(Spec)
	if err := parseStructFields(t, spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// parseStructFields recursively processes struct fields, handling embedded structs.
func parseStructFields(t reflect.Type, spec Spec) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := parseStructFields(field.Type, spec); err != nil {
				return err
			}
			continue
		}

		// Skip fields without bosun tag
		tagValue, ok := field.Tag.Lookup(tagName)
		if !ok || tagValue == "" {
			continue
		}

		// Parse tag value
		tagParts, err := parseTagValue(tagValue)
		if err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}

		// Create field spec
		fs, err := parseFieldSpec(field.Name, field.Type, tagParts)
		if err != nil {
			return err
		}

		// Check for duplicate keys
		if existing, ok := spec[fs.Key]; ok {
			return fmt.Errorf("duplicate key %q: defined in both %s and %s", fs.Key, existing.FieldName, fs.FieldName)
		}

		spec[fs.Key] = fs
	}

	return nil
}
