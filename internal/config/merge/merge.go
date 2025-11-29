package merge

import (
	"reflect"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// Options configures the merge behavior.
type Options struct {
	// EnableEnv enables the environment variable layer.
	// When false, env values are ignored in merge precedence.
	// Default: false (disabled in v1)
	EnableEnv bool
}

// Merge combines configuration from multiple sources with defined precedence.
//
// Precedence (lowest to highest):
//  1. defaults - Built-in defaults from schema tags
//  2. file - Config file values (if provided)
//  3. env - Environment variables (if opts.EnableEnv && env != nil)
//  4. labels - Docker label values (if provided)
//
// Parameters:
//   - spec: Schema specification for field metadata (unused in v1, reserved for future)
//   - defaults: Base configuration with default values (required)
//   - file: Config loaded from file (may be nil)
//   - env: Config loaded from environment (may be nil, ignored if !opts.EnableEnv)
//   - labels: Config loaded from Docker labels (may be nil)
//   - opts: Merge options
//
// Returns:
//   - merged: The final merged configuration
//   - err: Error if merge fails (should be rare; validation happens in loader)
//
// Behavior:
//   - For each field, higher precedence non-zero values override lower
//   - Zero values in higher layers are treated as "not set" (don't override)
//   - Nil layers are skipped entirely
//   - Output is deterministic for identical inputs
func Merge(_ schema.Spec, defaults schema.ConfigV1, file, env, labels *schema.ConfigV1, opts Options) (schema.ConfigV1, error) {
	// Start with defaults as the base
	merged := defaults

	// Apply file layer if provided
	if file != nil {
		merged = mergeLayer(merged, *file)
	}

	// Apply env layer if enabled and provided
	if opts.EnableEnv && env != nil {
		merged = mergeLayer(merged, *env)
	}

	// Apply labels layer if provided (highest precedence)
	if labels != nil {
		merged = mergeLayer(merged, *labels)
	}

	return merged, nil
}

// mergeLayer merges an override config into a base config.
// Non-zero values in override replace values in base.
func mergeLayer(base, override schema.ConfigV1) schema.ConfigV1 {
	result := base

	baseVal := reflect.ValueOf(&result).Elem()
	overrideVal := reflect.ValueOf(override)

	mergeStructRecursive(baseVal, overrideVal)

	return result
}

// mergeStructRecursive recursively merges struct fields.
func mergeStructRecursive(base, override reflect.Value) {
	t := base.Type()

	for i := 0; i < base.NumField(); i++ {
		baseField := base.Field(i)
		overrideField := override.Field(i)
		fieldType := t.Field(i)

		// Handle embedded structs recursively
		if fieldType.Anonymous && baseField.Kind() == reflect.Struct {
			mergeStructRecursive(baseField, overrideField)
			continue
		}

		// Skip if base field is not settable
		if !baseField.CanSet() {
			continue
		}

		// Only override if the override field is not a zero value
		if !isZeroValue(overrideField) {
			baseField.Set(overrideField)
		}
	}
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
// Special handling for bool: false is considered a valid value, not zero.
// We use a different approach - we consider bool zero only if it came from
// a struct that wasn't explicitly set. Since we're merging parsed configs,
// any bool value that differs from zero is intentional.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		// For bools, we treat false as "not set" since we can't distinguish
		// between an unset bool and an explicitly set false without additional tracking.
		// This is a known limitation - see the contract documentation.
		// In practice, users who want to override to false should rely on
		// the default being true and not setting the label at all.
		return !v.Bool() // Always treat false as zero for now
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// For structs, check if all fields are zero
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
