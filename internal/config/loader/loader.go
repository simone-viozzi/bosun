package loader

import (
	"reflect"
	"strings"
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

const bosunLabelPrefix = "bosun."

// LoadOptions configures label loading behavior.
type LoadOptions struct {
	// Strict enables strict validation mode where unknown keys are errors.
	// When false (default), unknown keys generate warnings instead of errors.
	Strict bool
}

// filterBosunLabels returns only labels that have the "bosun." prefix.
func filterBosunLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range labels {
		if strings.HasPrefix(k, bosunLabelPrefix) {
			result[k] = v
		}
	}
	return result
}

// setField sets a field in the config struct using reflection.
// It navigates through embedded structs to find the target field.
func setField(cfg *schema.ConfigV1, fieldSpec schema.FieldSpec, value any) {
	v := reflect.ValueOf(cfg).Elem()
	setFieldRecursive(v, fieldSpec.FieldName, value)
}

// setFieldRecursive recursively searches for and sets a field by name.
// Returns true if the field was found and set, false otherwise.
func setFieldRecursive(v reflect.Value, fieldName string, value any) bool {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Check if this is the target field
		if fieldType.Name == fieldName {
			setReflectValue(field, value)
			return true
		}

		// If this is an embedded struct, search within it
		if fieldType.Anonymous && field.Kind() == reflect.Struct {
			if setFieldRecursive(field, fieldName, value) {
				return true // Found and set
			}
		}
	}

	return false // Field not found
}

// setReflectValue sets a reflect.Value to the given value.
func setReflectValue(field reflect.Value, value any) {
	if !field.CanSet() {
		return // Skip unexported fields
	}

	rv := reflect.ValueOf(value)

	// Handle type conversions
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		case time.Duration:
			field.SetInt(int64(v))
		default:
			field.Set(rv)
		}
	case reflect.String:
		field.SetString(value.(string))
	case reflect.Bool:
		field.SetBool(value.(bool))
	case reflect.Slice:
		if rv.Kind() == reflect.Slice {
			field.Set(rv)
		}
	default:
		field.Set(rv)
	}
}

// FromLabels parses Docker labels into a typed ConfigV1 struct.
//
// Parameters:
//   - spec: Schema specification from ParseTags[ConfigV1]()
//   - labels: Raw Docker labels map (e.g., from container inspection)
//   - scope: The entity scope (container, volume, network)
//   - opts: Optional LoadOptions (default: lenient mode where unknown keys are warnings)
//
// Returns:
//   - cfg: Parsed configuration (partial if errors)
//   - result: ValidationErrors containing any errors and warnings
//
// Behavior:
//   - Filters labels to only those with "bosun." prefix
//   - Validates all keys are known in spec (warning in lenient mode, error in strict)
//   - Validates scope matches (global allowed anywhere)
//   - Parses values according to declared types
//   - Returns ALL errors, not just first one
//   - Callers should check result.HasErrors() to determine success
func FromLabels(spec schema.Spec, labels map[string]string, scope schema.Scope, opts ...LoadOptions) (schema.ConfigV1, ValidationErrors) {
	var cfg schema.ConfigV1
	var errs ValidationErrors

	// Merge options (default is lenient mode)
	var opt LoadOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Filter to only bosun.* labels
	bosunLabels := filterBosunLabels(labels)

	// Track which keys we've seen for required field checking
	seenKeys := make(map[string]bool)

	// Process each label
	for key, value := range bosunLabels {
		seenKeys[key] = true

		// Look up the field spec
		fieldSpec, exists := spec.Get(key)
		if !exists {
			// Unknown key: error in strict mode, warning in lenient mode
			if opt.Strict {
				errs.AddUnknownKey(key, scope)
			} else {
				errs.AddUnknownKeyWarning(key, scope)
			}
			continue
		}

		// Validate scope
		if !isScopeAllowed(fieldSpec.Scope, scope) {
			errs.AddScopeMismatch(key, fieldSpec.Scope, scope)
			continue
		}

		// Parse the value according to type
		parsedValue, parseErr := parseValue(value, fieldSpec)
		if parseErr != nil {
			errs.Add(*parseErr)
			continue
		}

		// Set the field in the config
		setField(&cfg, fieldSpec, parsedValue)
	}

	// Check for required fields
	for _, key := range spec.Keys() {
		fieldSpec, _ := spec.Get(key)
		if fieldSpec.Required && !seenKeys[key] {
			// Only check required if scope matches
			if isScopeAllowed(fieldSpec.Scope, scope) {
				errs.AddRequiredMissing(key, scope)
			}
		}
	}

	return cfg, errs
}

// isScopeAllowed checks if a field scope is allowed on an entity scope.
// Global scope is allowed on any entity.
func isScopeAllowed(fieldScope, entityScope schema.Scope) bool {
	if fieldScope == schema.ScopeGlobal {
		return true
	}
	return fieldScope == entityScope
}

// parseValue parses a string value according to the field spec's type.
func parseValue(value string, fieldSpec schema.FieldSpec) (any, *ValidationError) {
	switch fieldSpec.Type {
	case schema.TypeString:
		return value, nil

	case schema.TypeBool:
		b, err := parseBool(value)
		if err != nil {
			return nil, &ValidationError{
				Key:     fieldSpec.Key,
				Value:   value,
				Message: "invalid bool for key '" + fieldSpec.Key + "': " + err.Error(),
				Err:     err,
			}
		}
		return b, nil

	case schema.TypeInt:
		i, err := parseInt(value)
		if err != nil {
			return nil, &ValidationError{
				Key:     fieldSpec.Key,
				Value:   value,
				Message: "invalid int for key '" + fieldSpec.Key + "': " + err.Error(),
				Err:     err,
			}
		}
		return i, nil

	case schema.TypeDuration:
		d, err := parseDuration(value)
		if err != nil {
			return nil, &ValidationError{
				Key:     fieldSpec.Key,
				Value:   value,
				Message: "invalid duration for key '" + fieldSpec.Key + "': " + err.Error(),
				Err:     err,
			}
		}
		return d, nil

	case schema.TypeSize:
		s, err := parseSize(value)
		if err != nil {
			return nil, &ValidationError{
				Key:     fieldSpec.Key,
				Value:   value,
				Message: "invalid size for key '" + fieldSpec.Key + "': " + err.Error(),
				Err:     err,
			}
		}
		return s, nil

	case schema.TypeEnum:
		e, ok := parseEnum(value, fieldSpec.Enum)
		if !ok {
			return nil, &ValidationError{
				Key:     fieldSpec.Key,
				Value:   value,
				Message: "invalid enum value '" + value + "' for key '" + fieldSpec.Key + "': must be one of [" + strings.Join(fieldSpec.Enum, ", ") + "]",
			}
		}
		return e, nil

	case schema.TypeList:
		l, err := parseList(value)
		if err != nil {
			return nil, &ValidationError{
				Key:     fieldSpec.Key,
				Value:   value,
				Message: "invalid list for key '" + fieldSpec.Key + "': " + err.Error(),
				Err:     err,
			}
		}
		return l, nil

	default:
		return value, nil
	}
}
