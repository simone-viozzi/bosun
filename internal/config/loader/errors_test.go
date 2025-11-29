package loader

import (
	"errors"
	"strings"
	"testing"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

func TestValidationError_Error(t *testing.T) {
	ve := ValidationError{
		Key:     "bosun.container.stopGracePeriod",
		Value:   "invalid",
		Scope:   schema.ScopeContainer,
		Message: "invalid duration for key 'bosun.container.stopGracePeriod': time: invalid duration \"invalid\"",
	}

	got := ve.Error()
	if got != ve.Message {
		t.Errorf("Error() = %q, want %q", got, ve.Message)
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("parse error")
	ve := ValidationError{
		Key:     "bosun.container.stopGracePeriod",
		Message: "test error",
		Err:     underlyingErr,
	}

	got := ve.Unwrap()
	if got != underlyingErr {
		t.Errorf("Unwrap() = %v, want %v", got, underlyingErr)
	}

	// Test unwrap with nil Err
	ve2 := ValidationError{Key: "test"}
	if ve2.Unwrap() != nil {
		t.Error("Unwrap() should return nil when Err is nil")
	}
}

func TestValidationErrors_Error_Empty(t *testing.T) {
	var errs ValidationErrors
	got := errs.Error()
	want := "no validation errors"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidationErrors_Error_Single(t *testing.T) {
	errs := ValidationErrors{
		{Key: "bosun.test", Message: "test error message"},
	}
	got := errs.Error()
	want := "test error message"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidationErrors_Error_Multiple(t *testing.T) {
	errs := ValidationErrors{
		{Key: "bosun.test1", Message: "first error"},
		{Key: "bosun.test2", Message: "second error"},
	}
	got := errs.Error()

	// Should contain count
	if !strings.Contains(got, "2 validation errors") {
		t.Errorf("Error() should contain '2 validation errors', got %q", got)
	}

	// Should contain both errors
	if !strings.Contains(got, "first error") {
		t.Errorf("Error() should contain 'first error', got %q", got)
	}
	if !strings.Contains(got, "second error") {
		t.Errorf("Error() should contain 'second error', got %q", got)
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	tests := []struct {
		name string
		errs ValidationErrors
		want bool
	}{
		{
			name: "empty",
			errs: ValidationErrors{},
			want: false,
		},
		{
			name: "nil",
			errs: nil,
			want: false,
		},
		{
			name: "with errors",
			errs: ValidationErrors{{Key: "test", Message: "error"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errs.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationErrors_Add(t *testing.T) {
	var errs ValidationErrors
	ve := ValidationError{Key: "bosun.test", Message: "test error"}

	errs.Add(ve)

	if len(errs) != 1 {
		t.Errorf("Add() should add one error, got %d", len(errs))
	}
	if errs[0].Key != "bosun.test" {
		t.Errorf("Add() should preserve error key, got %q", errs[0].Key)
	}
}

func TestValidationErrors_AddUnknownKey(t *testing.T) {
	var errs ValidationErrors
	errs.AddUnknownKey("bosun.container.unknownKey", schema.ScopeContainer)

	if len(errs) != 1 {
		t.Fatalf("AddUnknownKey() should add one error, got %d", len(errs))
	}

	ve := errs[0]
	if ve.Key != "bosun.container.unknownKey" {
		t.Errorf("Key = %q, want %q", ve.Key, "bosun.container.unknownKey")
	}
	if ve.Scope != schema.ScopeContainer {
		t.Errorf("Scope = %v, want %v", ve.Scope, schema.ScopeContainer)
	}
	if !strings.Contains(ve.Message, "unknown key") {
		t.Errorf("Message should contain 'unknown key', got %q", ve.Message)
	}
}

func TestValidationErrors_AddScopeMismatch(t *testing.T) {
	var errs ValidationErrors
	errs.AddScopeMismatch("bosun.container.stopGracePeriod", schema.ScopeContainer, schema.ScopeVolume)

	if len(errs) != 1 {
		t.Fatalf("AddScopeMismatch() should add one error, got %d", len(errs))
	}

	ve := errs[0]
	if !strings.Contains(ve.Message, "not allowed on scope") {
		t.Errorf("Message should contain 'not allowed on scope', got %q", ve.Message)
	}
	if !strings.Contains(ve.Message, "volume") {
		t.Errorf("Message should contain 'volume', got %q", ve.Message)
	}
}

func TestValidationErrors_AddTypeParseFailed(t *testing.T) {
	var errs ValidationErrors
	parseErr := errors.New("time: invalid duration \"bad\"")
	errs.AddTypeParseFailed("bosun.container.stopGracePeriod", "bad", "duration", schema.ScopeContainer, parseErr)

	if len(errs) != 1 {
		t.Fatalf("AddTypeParseFailed() should add one error, got %d", len(errs))
	}

	ve := errs[0]
	if ve.Key != "bosun.container.stopGracePeriod" {
		t.Errorf("Key = %q, want %q", ve.Key, "bosun.container.stopGracePeriod")
	}
	if ve.Value != "bad" {
		t.Errorf("Value = %q, want %q", ve.Value, "bad")
	}
	if !strings.Contains(ve.Message, "invalid duration") {
		t.Errorf("Message should contain 'invalid duration', got %q", ve.Message)
	}
	if ve.Err != parseErr {
		t.Errorf("Err = %v, want %v", ve.Err, parseErr)
	}
}

func TestValidationErrors_AddInvalidEnum(t *testing.T) {
	var errs ValidationErrors
	errs.AddInvalidEnum("bosun.container.logLevel", "verbose", []string{"debug", "info", "warn", "error"}, schema.ScopeContainer)

	if len(errs) != 1 {
		t.Fatalf("AddInvalidEnum() should add one error, got %d", len(errs))
	}

	ve := errs[0]
	if !strings.Contains(ve.Message, "verbose") {
		t.Errorf("Message should contain 'verbose', got %q", ve.Message)
	}
	if !strings.Contains(ve.Message, "must be one of") {
		t.Errorf("Message should contain 'must be one of', got %q", ve.Message)
	}
	if !strings.Contains(ve.Message, "debug, info, warn, error") {
		t.Errorf("Message should list allowed values, got %q", ve.Message)
	}
}

func TestValidationErrors_AddRequiredMissing(t *testing.T) {
	var errs ValidationErrors
	errs.AddRequiredMissing("bosun.instance", schema.ScopeGlobal)

	if len(errs) != 1 {
		t.Fatalf("AddRequiredMissing() should add one error, got %d", len(errs))
	}

	ve := errs[0]
	if !strings.Contains(ve.Message, "required") {
		t.Errorf("Message should contain 'required', got %q", ve.Message)
	}
	if !strings.Contains(ve.Message, "not provided") {
		t.Errorf("Message should contain 'not provided', got %q", ve.Message)
	}
}

func TestValidationErrors_AsError(t *testing.T) {
	// Test that ValidationErrors can be used as error
	var errs ValidationErrors
	errs.Add(ValidationError{Key: "test", Message: "test error"})

	// Should work with errors.As
	var testErrs ValidationErrors
	var err error = errs
	if !errors.As(err, &testErrs) {
		t.Error("errors.As should succeed with ValidationErrors")
	}
	if len(testErrs) != 1 {
		t.Errorf("errors.As should preserve error count, got %d", len(testErrs))
	}
}
