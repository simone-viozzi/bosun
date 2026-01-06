package loader

import (
	"fmt"
	"strings"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// ValidationError represents a single validation failure during label parsing.
type ValidationError struct {
	// Key is the Docker label key that failed validation.
	// Example: "bosun.container.unknownKey"
	Key string

	// Value is the raw value that was provided for the key.
	Value string

	// Scope is the entity scope where validation failed.
	Scope schema.Scope

	// Message is a human-readable description of the error.
	Message string

	// Err is the underlying error, if any. Used for wrapping parse errors.
	Err error
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error for error unwrapping support.
func (e ValidationError) Unwrap() error {
	return e.Err
}

// ValidationErrors is a collection of validation errors and warnings.
// It implements the error interface, allowing it to be returned from functions
// that return error while still providing access to individual errors.
// Warnings are non-fatal issues (e.g., unknown keys in lenient mode).
type ValidationErrors struct {
	// Errors contains fatal validation failures that should halt processing.
	Errors []ValidationError
	// Warnings contains non-fatal issues that should be reported but don't halt processing.
	Warnings []ValidationError
}

// Error implements the error interface.
// It returns a formatted string containing all validation errors.
func (e ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Message
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d validation errors:\n", len(e.Errors)))
	for i, ve := range e.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, ve.Message))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// HasWarnings returns true if there are any validation warnings.
func (e ValidationErrors) HasWarnings() bool {
	return len(e.Warnings) > 0
}

// Add appends a validation error to the collection.
func (e *ValidationErrors) Add(err ValidationError) {
	e.Errors = append(e.Errors, err)
}

// AddWarning appends a validation warning to the collection.
func (e *ValidationErrors) AddWarning(warn ValidationError) {
	e.Warnings = append(e.Warnings, warn)
}

// AddUnknownKey adds an error for an unknown configuration key (strict mode).
func (e *ValidationErrors) AddUnknownKey(key string, scope schema.Scope) {
	e.Add(ValidationError{
		Key:     key,
		Scope:   scope,
		Message: fmt.Sprintf("unknown key: %s (rejected in strict mode)", key),
	})
}

// AddUnknownKeyWarning adds a warning for an unknown configuration key (lenient mode).
func (e *ValidationErrors) AddUnknownKeyWarning(key string, scope schema.Scope) {
	e.AddWarning(ValidationError{
		Key:     key,
		Scope:   scope,
		Message: fmt.Sprintf("unknown key: %s (ignored)", key),
	})
}

// AddScopeMismatch adds an error for a scope mismatch.
func (e *ValidationErrors) AddScopeMismatch(key string, fieldScope, entityScope schema.Scope) {
	e.Add(ValidationError{
		Key:     key,
		Scope:   entityScope,
		Message: fmt.Sprintf("key '%s' not allowed on scope '%s'", key, entityScope),
	})
}

// AddTypeParseFailed adds an error for a type parsing failure.
func (e *ValidationErrors) AddTypeParseFailed(key, value string, typeName string, scope schema.Scope, parseErr error) {
	e.Add(ValidationError{
		Key:     key,
		Value:   value,
		Scope:   scope,
		Message: fmt.Sprintf("invalid %s for key '%s': %v", typeName, key, parseErr),
		Err:     parseErr,
	})
}

// AddInvalidEnum adds an error for an invalid enum value.
func (e *ValidationErrors) AddInvalidEnum(key, value string, allowed []string, scope schema.Scope) {
	e.Add(ValidationError{
		Key:     key,
		Value:   value,
		Scope:   scope,
		Message: fmt.Sprintf("invalid enum value '%s' for key '%s': must be one of [%s]", value, key, strings.Join(allowed, ", ")),
	})
}

// AddRequiredMissing adds an error for a missing required field.
func (e *ValidationErrors) AddRequiredMissing(key string, scope schema.Scope) {
	e.Add(ValidationError{
		Key:     key,
		Scope:   scope,
		Message: fmt.Sprintf("required key '%s' not provided", key),
	})
}
