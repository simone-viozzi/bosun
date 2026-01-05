package loader

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// getTestSpec returns the V1Spec for testing.
func getTestSpec(t *testing.T) schema.Spec {
	t.Helper()
	return schema.V1Spec()
}

// TestFromLabels_StringType tests parsing of string type labels.
func TestFromLabels_StringType(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.instance": "my-instance",
	}

	cfg, err := FromLabels(spec, labels, schema.ScopeGlobal)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}

	if cfg.Instance != "my-instance" {
		t.Errorf("Instance = %q, want %q", cfg.Instance, "my-instance")
	}
}

// TestFromLabels_BoolType tests parsing of boolean type labels.
func TestFromLabels_BoolType(t *testing.T) {
	spec := getTestSpec(t)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"True", "True", true},
		{"False", "False", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := map[string]string{
				"bosun.container.autoRestart": tt.value,
			}

			cfg, err := FromLabels(spec, labels, schema.ScopeContainer)
			if err != nil {
				t.Fatalf("FromLabels() error = %v", err)
			}

			if cfg.AutoRestart != tt.want {
				t.Errorf("AutoRestart = %v, want %v", cfg.AutoRestart, tt.want)
			}
		})
	}
}

// TestFromLabels_IntType tests parsing of integer type labels.
func TestFromLabels_IntType(t *testing.T) {
	spec := getTestSpec(t)

	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"zero", "0", 0},
		{"positive", "42", 42},
		{"negative", "-10", -10},
		{"large", "9999", 9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := map[string]string{
				"bosun.network.priority": tt.value,
			}

			cfg, err := FromLabels(spec, labels, schema.ScopeNetwork)
			if err != nil {
				t.Fatalf("FromLabels() error = %v", err)
			}

			if cfg.Priority != tt.want {
				t.Errorf("Priority = %v, want %v", cfg.Priority, tt.want)
			}
		})
	}
}

// TestFromLabels_DurationType tests parsing of duration type labels.
func TestFromLabels_DurationType(t *testing.T) {
	spec := getTestSpec(t)

	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"hours", "1h", time.Hour},
		{"combined", "1h30m", time.Hour + 30*time.Minute},
		{"milliseconds", "500ms", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := map[string]string{
				"bosun.container.stopGracePeriod": tt.value,
			}

			cfg, err := FromLabels(spec, labels, schema.ScopeContainer)
			if err != nil {
				t.Fatalf("FromLabels() error = %v", err)
			}

			if cfg.StopGracePeriod != tt.want {
				t.Errorf("StopGracePeriod = %v, want %v", cfg.StopGracePeriod, tt.want)
			}
		})
	}
}

// TestFromLabels_SizeType tests parsing of size type labels.
func TestFromLabels_SizeType(t *testing.T) {
	spec := getTestSpec(t)

	tests := []struct {
		name  string
		value string
		want  int64
	}{
		{"bytes", "1024", 1024},
		{"kilobytes", "1K", 1024},
		{"megabytes", "1M", 1024 * 1024},
		{"gigabytes", "1G", 1024 * 1024 * 1024},
		{"10GB", "10GB", 10 * 1024 * 1024 * 1024},
		{"500MB", "500MB", 500 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := map[string]string{
				"bosun.volume.maxSize": tt.value,
			}

			cfg, err := FromLabels(spec, labels, schema.ScopeVolume)
			if err != nil {
				t.Fatalf("FromLabels() error = %v", err)
			}

			if cfg.MaxSize != tt.want {
				t.Errorf("MaxSize = %v, want %v", cfg.MaxSize, tt.want)
			}
		})
	}
}

// TestFromLabels_EnumType tests parsing of enum type labels.
func TestFromLabels_EnumType(t *testing.T) {
	spec := getTestSpec(t)

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"debug", "debug", "debug"},
		{"info", "info", "info"},
		{"warn", "warn", "warn"},
		{"error", "error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := map[string]string{
				"bosun.container.logLevel": tt.value,
			}

			cfg, err := FromLabels(spec, labels, schema.ScopeContainer)
			if err != nil {
				t.Fatalf("FromLabels() error = %v", err)
			}

			if cfg.LogLevel != tt.want {
				t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, tt.want)
			}
		})
	}
}

// TestFromLabels_ListType tests parsing of list type labels (CSV and JSON).
// Note: Currently ConfigV1 doesn't have a list type field, so this test
// verifies the parse function directly. When a list field is added to the
// schema, this test should be updated to test it through FromLabels.
func TestFromLabels_ListType(t *testing.T) {
	// Test parseList directly since ConfigV1 doesn't have list fields yet
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "csv",
			input: "a,b,c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "json",
			input: `["a", "b", "c"]`,
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "csv with spaces",
			input: "a, b, c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "json with special chars",
			input: `["hello,world", "foo"]`,
			want:  []string{"hello,world", "foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("parseList() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFromLabels_UnknownKey tests that unknown keys are rejected.
func TestFromLabels_UnknownKey(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.unknownKey": "value",
	}

	_, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err == nil {
		t.Fatal("FromLabels() should return error for unknown key")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if len(verrs.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(verrs.Errors))
	}

	if !strings.Contains(verrs.Errors[0].Message, "unknown key") {
		t.Errorf("Error message should contain 'unknown key', got %q", verrs.Errors[0].Message)
	}
}

// TestFromLabels_MultipleUnknownKeys tests that multiple unknown keys are all reported.
func TestFromLabels_MultipleUnknownKeys(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.unknownKey1": "value1",
		"bosun.container.unknownKey2": "value2",
		"bosun.container.unknownKey3": "value3",
	}

	_, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err == nil {
		t.Fatal("FromLabels() should return error for unknown keys")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if len(verrs.Errors) != 3 {
		t.Fatalf("Expected 3 errors, got %d", len(verrs.Errors))
	}
}

// TestFromLabels_AllErrorsReported tests that all errors are reported, not just the first.
func TestFromLabels_AllErrorsReported(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.unknownKey":      "value",
		"bosun.container.stopGracePeriod": "invalid-duration",
		"bosun.container.autoRestart":     "not-a-bool",
	}

	_, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err == nil {
		t.Fatal("FromLabels() should return error")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	// Should have at least 3 errors
	if len(verrs.Errors) < 3 {
		t.Fatalf("Expected at least 3 errors, got %d: %v", len(verrs.Errors), verrs)
	}
}

// TestFromLabels_ScopeMismatch tests that labels applied to wrong scopes are rejected.
func TestFromLabels_ScopeMismatch(t *testing.T) {
	spec := getTestSpec(t)

	// Container label on volume scope
	labels := map[string]string{
		"bosun.container.stopGracePeriod": "30s",
	}

	_, err := FromLabels(spec, labels, schema.ScopeVolume)
	if err == nil {
		t.Fatal("FromLabels() should return error for scope mismatch")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if len(verrs.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(verrs.Errors))
	}

	if !strings.Contains(verrs.Errors[0].Message, "not allowed on scope") {
		t.Errorf("Error message should contain 'not allowed on scope', got %q", verrs.Errors[0].Message)
	}
}

// TestFromLabels_GlobalScopeAllowed tests that global scope labels work on any entity.
func TestFromLabels_GlobalScopeAllowed(t *testing.T) {
	spec := getTestSpec(t)

	// Global label (bosun.instance) should work on container scope
	labels := map[string]string{
		"bosun.instance": "my-instance",
	}

	// Test on container scope
	cfg, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}
	if cfg.Instance != "my-instance" {
		t.Errorf("Instance = %q, want %q", cfg.Instance, "my-instance")
	}

	// Test on volume scope
	cfg, err = FromLabels(spec, labels, schema.ScopeVolume)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}
	if cfg.Instance != "my-instance" {
		t.Errorf("Instance = %q, want %q", cfg.Instance, "my-instance")
	}

	// Test on network scope
	cfg, err = FromLabels(spec, labels, schema.ScopeNetwork)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}
	if cfg.Instance != "my-instance" {
		t.Errorf("Instance = %q, want %q", cfg.Instance, "my-instance")
	}
}

// TestFromLabels_MatchingScope tests that labels work correctly on their intended scope.
func TestFromLabels_MatchingScope(t *testing.T) {
	spec := getTestSpec(t)

	tests := []struct {
		name   string
		labels map[string]string
		scope  schema.Scope
	}{
		{
			name:   "container label on container",
			labels: map[string]string{"bosun.container.stopGracePeriod": "30s"},
			scope:  schema.ScopeContainer,
		},
		{
			name:   "volume label on volume",
			labels: map[string]string{"bosun.volume.maxSize": "10GB"},
			scope:  schema.ScopeVolume,
		},
		{
			name:   "network label on network",
			labels: map[string]string{"bosun.network.priority": "50"},
			scope:  schema.ScopeNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromLabels(spec, tt.labels, tt.scope)
			if err != nil {
				t.Fatalf("FromLabels() error = %v", err)
			}
		})
	}
}

// TestFromLabels_InvalidDuration tests error handling for invalid duration values.
func TestFromLabels_InvalidDuration(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.stopGracePeriod": "not-a-duration",
	}

	_, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err == nil {
		t.Fatal("FromLabels() should return error for invalid duration")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if !strings.Contains(verrs.Errors[0].Message, "invalid duration") {
		t.Errorf("Error message should contain 'invalid duration', got %q", verrs.Errors[0].Message)
	}
}

// TestFromLabels_InvalidBool tests error handling for invalid boolean values.
func TestFromLabels_InvalidBool(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.autoRestart": "maybe",
	}

	_, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err == nil {
		t.Fatal("FromLabels() should return error for invalid bool")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if !strings.Contains(verrs.Errors[0].Message, "invalid bool") {
		t.Errorf("Error message should contain 'invalid bool', got %q", verrs.Errors[0].Message)
	}
}

// TestFromLabels_InvalidSize tests error handling for invalid size values.
func TestFromLabels_InvalidSize(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.volume.maxSize": "not-a-size",
	}

	_, err := FromLabels(spec, labels, schema.ScopeVolume)
	if err == nil {
		t.Fatal("FromLabels() should return error for invalid size")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if !strings.Contains(verrs.Errors[0].Message, "invalid size") {
		t.Errorf("Error message should contain 'invalid size', got %q", verrs.Errors[0].Message)
	}
}

// TestFromLabels_InvalidEnum tests error handling for invalid enum values.
func TestFromLabels_InvalidEnum(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.logLevel": "verbose", // Not a valid log level
	}

	_, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err == nil {
		t.Fatal("FromLabels() should return error for invalid enum")
	}

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("Error should be ValidationErrors, got %T", err)
	}

	if !strings.Contains(verrs.Errors[0].Message, "invalid enum") {
		t.Errorf("Error message should contain 'invalid enum', got %q", verrs.Errors[0].Message)
	}

	// Should list valid values
	if !strings.Contains(verrs.Errors[0].Message, "debug") {
		t.Errorf("Error message should list valid values, got %q", verrs.Errors[0].Message)
	}
}

// TestFromLabels_RequiredMissing tests error handling when required fields are missing.
// Note: Currently no fields in ConfigV1 are marked as required, so this test
// verifies the logic is in place for when required fields are added.
func TestFromLabels_RequiredMissing(t *testing.T) {
	// This test is a placeholder for when required fields are added to the schema.
	// The required field checking logic is already implemented in FromLabels().
	t.Skip("No required fields in current schema - test will be enabled when required fields are added")
}

// TestFromLabels_NonBosunLabelsIgnored tests that non-bosun labels are silently ignored.
func TestFromLabels_NonBosunLabelsIgnored(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"com.example.label":               "ignored",
		"org.label-schema.version":        "1.0",
		"bosun.container.stopGracePeriod": "30s",
	}

	cfg, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}

	if cfg.StopGracePeriod != 30*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v", cfg.StopGracePeriod, 30*time.Second)
	}
}

// TestFromLabels_EmptyLabels tests handling of empty labels map.
func TestFromLabels_EmptyLabels(t *testing.T) {
	spec := getTestSpec(t)

	cfg, err := FromLabels(spec, map[string]string{}, schema.ScopeContainer)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}

	// Should return zero-value config
	if cfg.StopGracePeriod != 0 {
		t.Errorf("StopGracePeriod should be zero, got %v", cfg.StopGracePeriod)
	}
}

// TestFromLabels_MultipleValidLabels tests parsing multiple valid labels together.
func TestFromLabels_MultipleValidLabels(t *testing.T) {
	spec := getTestSpec(t)

	labels := map[string]string{
		"bosun.container.stopGracePeriod":     "45s",
		"bosun.container.healthCheckInterval": "1m",
		"bosun.container.autoRestart":         "false",
		"bosun.container.logLevel":            "debug",
	}

	cfg, err := FromLabels(spec, labels, schema.ScopeContainer)
	if err != nil {
		t.Fatalf("FromLabels() error = %v", err)
	}

	if cfg.StopGracePeriod != 45*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v", cfg.StopGracePeriod, 45*time.Second)
	}
	if cfg.HealthCheckInterval != time.Minute {
		t.Errorf("HealthCheckInterval = %v, want %v", cfg.HealthCheckInterval, time.Minute)
	}
	if cfg.AutoRestart != false {
		t.Errorf("AutoRestart = %v, want false", cfg.AutoRestart)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
}

// TestFilterBosunLabels tests the label filtering function.
func TestFilterBosunLabels(t *testing.T) {
	labels := map[string]string{
		"bosun.container.stopGracePeriod":  "30s",
		"bosun.instance":                   "test",
		"com.example.label":                "ignored",
		"org.opencontainers.image.version": "1.0",
	}

	result := filterBosunLabels(labels)

	if len(result) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(result))
	}

	if _, ok := result["bosun.container.stopGracePeriod"]; !ok {
		t.Error("Expected bosun.container.stopGracePeriod in result")
	}
	if _, ok := result["bosun.instance"]; !ok {
		t.Error("Expected bosun.instance in result")
	}
	if _, ok := result["com.example.label"]; ok {
		t.Error("com.example.label should not be in result")
	}
}
