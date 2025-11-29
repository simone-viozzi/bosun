package merge

import (
	"reflect"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// getTestSpec returns the V1Spec for testing.
func getTestSpec(t *testing.T) schema.Spec {
	t.Helper()
	return schema.V1Spec()
}

// getTestDefaults returns default config for testing.
func getTestDefaults(t *testing.T) schema.ConfigV1 {
	t.Helper()
	return schema.V1Defaults()
}

// TestMerge_DefaultsOnly tests that defaults are returned when no overrides are provided.
func TestMerge_DefaultsOnly(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	merged, err := Merge(spec, defaults, nil, nil, nil, Options{})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should match defaults
	if merged.StopGracePeriod != defaults.StopGracePeriod {
		t.Errorf("StopGracePeriod = %v, want %v", merged.StopGracePeriod, defaults.StopGracePeriod)
	}
	if merged.HealthCheckInterval != defaults.HealthCheckInterval {
		t.Errorf("HealthCheckInterval = %v, want %v", merged.HealthCheckInterval, defaults.HealthCheckInterval)
	}
	if merged.AutoRestart != defaults.AutoRestart {
		t.Errorf("AutoRestart = %v, want %v", merged.AutoRestart, defaults.AutoRestart)
	}
	if merged.LogLevel != defaults.LogLevel {
		t.Errorf("LogLevel = %v, want %v", merged.LogLevel, defaults.LogLevel)
	}
}

// TestMerge_LabelsOverrideDefaults tests that labels override defaults.
func TestMerge_LabelsOverrideDefaults(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	labels := &schema.ConfigV1{}
	labels.StopGracePeriod = 45 * time.Second
	labels.LogLevel = "debug"

	merged, err := Merge(spec, defaults, nil, nil, labels, Options{})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Labels should override
	if merged.StopGracePeriod != 45*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v", merged.StopGracePeriod, 45*time.Second)
	}
	if merged.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want %v", merged.LogLevel, "debug")
	}

	// Unchanged fields should keep defaults
	if merged.HealthCheckInterval != defaults.HealthCheckInterval {
		t.Errorf("HealthCheckInterval = %v, want %v (default)", merged.HealthCheckInterval, defaults.HealthCheckInterval)
	}
}

// TestMerge_FileOverridesDefaults tests that file config overrides defaults.
func TestMerge_FileOverridesDefaults(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	file := &schema.ConfigV1{}
	file.StopGracePeriod = 60 * time.Second
	file.Priority = 50

	merged, err := Merge(spec, defaults, file, nil, nil, Options{})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// File should override defaults
	if merged.StopGracePeriod != 60*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v", merged.StopGracePeriod, 60*time.Second)
	}
	if merged.Priority != 50 {
		t.Errorf("Priority = %v, want %v", merged.Priority, 50)
	}
}

// TestMerge_LabelsOverrideFile tests that labels override file config.
func TestMerge_LabelsOverrideFile(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	file := &schema.ConfigV1{}
	file.StopGracePeriod = 60 * time.Second
	file.LogLevel = "warn"

	labels := &schema.ConfigV1{}
	labels.StopGracePeriod = 90 * time.Second

	merged, err := Merge(spec, defaults, file, nil, labels, Options{})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Labels should override file
	if merged.StopGracePeriod != 90*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v", merged.StopGracePeriod, 90*time.Second)
	}

	// File value should remain for fields not in labels
	if merged.LogLevel != "warn" {
		t.Errorf("LogLevel = %v, want %v (from file)", merged.LogLevel, "warn")
	}
}

// TestMerge_NilLayers tests that nil layers are handled correctly.
func TestMerge_NilLayers(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	// All nil layers - should return defaults
	merged, err := Merge(spec, defaults, nil, nil, nil, Options{})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if merged.StopGracePeriod != defaults.StopGracePeriod {
		t.Errorf("StopGracePeriod = %v, want %v", merged.StopGracePeriod, defaults.StopGracePeriod)
	}
}

// TestMerge_Deterministic tests that merge produces deterministic output.
func TestMerge_Deterministic(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	file := &schema.ConfigV1{}
	file.StopGracePeriod = 60 * time.Second

	labels := &schema.ConfigV1{}
	labels.LogLevel = "debug"

	// Run merge multiple times
	var results []schema.ConfigV1
	for i := 0; i < 10; i++ {
		merged, err := Merge(spec, defaults, file, nil, labels, Options{})
		if err != nil {
			t.Fatalf("Merge() error = %v", err)
		}
		results = append(results, merged)
	}

	// All results should be identical
	first := results[0]
	for i, result := range results[1:] {
		if result.StopGracePeriod != first.StopGracePeriod {
			t.Errorf("Run %d: StopGracePeriod = %v, want %v", i+1, result.StopGracePeriod, first.StopGracePeriod)
		}
		if result.LogLevel != first.LogLevel {
			t.Errorf("Run %d: LogLevel = %v, want %v", i+1, result.LogLevel, first.LogLevel)
		}
	}
}

// TestMerge_EnvDisabled tests that env layer is ignored when disabled.
func TestMerge_EnvDisabled(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	env := &schema.ConfigV1{}
	env.StopGracePeriod = 120 * time.Second

	merged, err := Merge(spec, defaults, nil, env, nil, Options{EnableEnv: false})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Env should be ignored when disabled
	if merged.StopGracePeriod != defaults.StopGracePeriod {
		t.Errorf("StopGracePeriod = %v, want %v (env should be ignored)", merged.StopGracePeriod, defaults.StopGracePeriod)
	}
}

// TestMerge_EnvEnabled tests that env layer is applied when enabled.
func TestMerge_EnvEnabled(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	env := &schema.ConfigV1{}
	env.StopGracePeriod = 120 * time.Second

	merged, err := Merge(spec, defaults, nil, env, nil, Options{EnableEnv: true})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Env should override when enabled
	if merged.StopGracePeriod != 120*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v", merged.StopGracePeriod, 120*time.Second)
	}
}

// TestMerge_EnvDoesNotOverrideLabels tests that labels take precedence over env.
func TestMerge_EnvDoesNotOverrideLabels(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	env := &schema.ConfigV1{}
	env.StopGracePeriod = 120 * time.Second

	labels := &schema.ConfigV1{}
	labels.StopGracePeriod = 45 * time.Second

	merged, err := Merge(spec, defaults, nil, env, labels, Options{EnableEnv: true})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Labels should take precedence over env
	if merged.StopGracePeriod != 45*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v (labels should override env)", merged.StopGracePeriod, 45*time.Second)
	}
}

// TestMerge_FullPrecedenceChain tests the full precedence: defaults < file < env < labels.
func TestMerge_FullPrecedenceChain(t *testing.T) {
	spec := getTestSpec(t)
	defaults := getTestDefaults(t)

	file := &schema.ConfigV1{}
	file.StopGracePeriod = 60 * time.Second
	file.HealthCheckInterval = 2 * time.Minute
	file.LogLevel = "warn"
	file.Priority = 50

	env := &schema.ConfigV1{}
	env.StopGracePeriod = 90 * time.Second
	env.HealthCheckInterval = 3 * time.Minute
	env.LogLevel = "error"

	labels := &schema.ConfigV1{}
	labels.StopGracePeriod = 120 * time.Second
	labels.HealthCheckInterval = 5 * time.Minute

	merged, err := Merge(spec, defaults, file, env, labels, Options{EnableEnv: true})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Labels should win for fields they set
	if merged.StopGracePeriod != 120*time.Second {
		t.Errorf("StopGracePeriod = %v, want %v (from labels)", merged.StopGracePeriod, 120*time.Second)
	}
	if merged.HealthCheckInterval != 5*time.Minute {
		t.Errorf("HealthCheckInterval = %v, want %v (from labels)", merged.HealthCheckInterval, 5*time.Minute)
	}

	// Env should win for fields only set in env
	if merged.LogLevel != "error" {
		t.Errorf("LogLevel = %v, want %v (from env)", merged.LogLevel, "error")
	}

	// File should win for fields only set in file
	if merged.Priority != 50 {
		t.Errorf("Priority = %v, want %v (from file)", merged.Priority, 50)
	}

	// Defaults should remain for unset fields
	if merged.AutoRestart != defaults.AutoRestart {
		t.Errorf("AutoRestart = %v, want %v (from defaults)", merged.AutoRestart, defaults.AutoRestart)
	}
}

// TestIsZeroValue tests the isZeroValue helper function.
func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero int64", int64(0), true},
		{"non-zero int64", int64(42), false},
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"zero duration", time.Duration(0), true},
		{"non-zero duration", time.Second, false},
		{"nil slice", []string(nil), true},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"a"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isZeroValue(reflect.ValueOf(tt.value))
			if got != tt.want {
				t.Errorf("isZeroValue(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
