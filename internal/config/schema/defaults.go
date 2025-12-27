package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
)

// DefaultOf creates a new instance of T with all default values applied
// from the bosun struct tags. Fields without defaults remain zero-valued.
// Returns an error if the type is invalid or parsing fails.
func DefaultOf[T any]() (T, error) {
	var zero T
	v := reflect.ValueOf(&zero).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return zero, fmt.Errorf("DefaultOf requires a struct type, got %s", t.Kind())
	}

	if err := applyDefaults(v, t); err != nil {
		return zero, err
	}

	return zero, nil
}

// applyDefaults recursively applies default values to a struct value.
func applyDefaults(v reflect.Value, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Handle embedded structs recursively.
		// Note: We check field.Anonymous before IsExported() because
		// embedded structs may have unexported type names but their
		// inner fields may still be settable.
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := applyDefaults(fieldVal, field.Type); err != nil {
				return err
			}
			continue
		}

		// Skip unexported fields (for non-embedded fields).
		if !field.IsExported() {
			continue
		}

		// Get the bosun tag.
		tagValue := field.Tag.Get("bosun")
		if tagValue == "" {
			continue
		}

		// Parse the tag to get the default value.
		tagParts := parseTagValue(tagValue)

		defaultStr, hasDefault := tagParts["default"]
		if !hasDefault || defaultStr == "" {
			continue
		}

		configType := ConfigType(tagParts["type"])

		// Extract enum values if present (pipe-separated).
		var enumValues []string
		if enumStr, ok := tagParts["enum"]; ok && enumStr != "" {
			parts := strings.Split(enumStr, "|")
			enumValues = make([]string, 0, len(parts))
			for _, p := range parts {
				enumValues = append(enumValues, strings.TrimSpace(p))
			}
		}

		// Parse and set the default value.
		if err := setDefaultValue(fieldVal, configType, defaultStr, enumValues); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	return nil
}

// setDefaultValue parses the default string and sets it on the field.
func setDefaultValue(fieldVal reflect.Value, configType ConfigType, defaultStr string, enumValues []string) error {
	if !fieldVal.CanSet() {
		return fmt.Errorf("cannot set field value")
	}

	switch configType {
	case TypeString:
		fieldVal.SetString(defaultStr)

	case TypeBool:
		b, err := parseBool(defaultStr)
		if err != nil {
			return fmt.Errorf("invalid bool default %q: %w", defaultStr, err)
		}
		fieldVal.SetBool(b)

	case TypeInt:
		i, err := parseInt(defaultStr)
		if err != nil {
			return fmt.Errorf("invalid int default %q: %w", defaultStr, err)
		}
		fieldVal.SetInt(i)

	case TypeDuration:
		d, err := parseDuration(defaultStr)
		if err != nil {
			return fmt.Errorf("invalid duration default %q: %w", defaultStr, err)
		}
		fieldVal.SetInt(int64(d))

	case TypeSize:
		s, err := parseSize(defaultStr)
		if err != nil {
			return fmt.Errorf("invalid size default %q: %w", defaultStr, err)
		}
		fieldVal.SetInt(s)

	case TypeEnum:
		// Validate that the default value is one of the allowed enum values.
		if len(enumValues) > 0 {
			valid := false
			for _, v := range enumValues {
				if v == defaultStr {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid enum default %q: must be one of %v", defaultStr, enumValues)
			}
		}
		fieldVal.SetString(defaultStr)

	case TypeList:
		list := parseList(defaultStr)
		fieldVal.Set(reflect.ValueOf(list))

	default:
		return fmt.Errorf("unsupported config type: %s", configType)
	}

	return nil
}

// parseBool parses a boolean string value.
// Accepts: "true", "false", "1", "0", "yes", "no".
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse %q as bool", s)
	}
}

// parseInt parses an integer string value.
func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// parseDuration parses a duration string value.
// Supports Go duration format (e.g., "30s", "5m", "1h30m").
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// parseSize parses a human-readable size string.
// Uses docker/go-units for parsing (e.g., "1GB", "500MB", "1.5TB").
func parseSize(s string) (int64, error) {
	return units.RAMInBytes(s)
}

// parseList parses a comma-separated list string.
// Returns a slice of trimmed strings.
func parseList(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
