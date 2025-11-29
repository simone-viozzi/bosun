package schema

import (
	"testing"
	"time"
)

func TestConfigV1_ParseTags(t *testing.T) {
	spec, err := ParseTags[ConfigV1]()
	if err != nil {
		t.Fatalf("ParseTags[ConfigV1]() error: %v", err)
	}

	// Verify expected number of fields
	expectedFields := 8 // 1 global + 4 container + 2 volume + 1 network
	if len(spec) != expectedFields {
		t.Errorf("ParseTags returned %d fields, want %d", len(spec), expectedFields)
	}

	// Verify all expected keys are present
	expectedKeys := []string{
		"bosun.instance",
		"bosun.container.stopGracePeriod",
		"bosun.container.healthCheckInterval",
		"bosun.container.autoRestart",
		"bosun.container.logLevel",
		"bosun.volume.backupEnabled",
		"bosun.volume.maxSize",
		"bosun.network.priority",
	}
	for _, key := range expectedKeys {
		if _, ok := spec.Get(key); !ok {
			t.Errorf("missing expected key: %s", key)
		}
	}

	// Verify specific field metadata
	t.Run("global fields", func(t *testing.T) {
		fs, ok := spec.Get("bosun.instance")
		if !ok {
			t.Fatal("missing bosun.instance")
		}
		if fs.Scope != ScopeGlobal {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeGlobal)
		}
		if fs.Type != TypeString {
			t.Errorf("Type = %q, want %q", fs.Type, TypeString)
		}
	})

	t.Run("container fields", func(t *testing.T) {
		fs, ok := spec.Get("bosun.container.stopGracePeriod")
		if !ok {
			t.Fatal("missing bosun.container.stopGracePeriod")
		}
		if fs.Scope != ScopeContainer {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeContainer)
		}
		if fs.Type != TypeDuration {
			t.Errorf("Type = %q, want %q", fs.Type, TypeDuration)
		}
		if fs.Default != "30s" {
			t.Errorf("Default = %q, want %q", fs.Default, "30s")
		}

		// Verify enum field
		fs, ok = spec.Get("bosun.container.logLevel")
		if !ok {
			t.Fatal("missing bosun.container.logLevel")
		}
		if fs.Type != TypeEnum {
			t.Errorf("Type = %q, want %q", fs.Type, TypeEnum)
		}
		expectedEnum := []string{"debug", "info", "warn", "error"}
		if len(fs.Enum) != len(expectedEnum) {
			t.Errorf("Enum has %d values, want %d", len(fs.Enum), len(expectedEnum))
		}
	})

	t.Run("volume fields", func(t *testing.T) {
		fs, ok := spec.Get("bosun.volume.maxSize")
		if !ok {
			t.Fatal("missing bosun.volume.maxSize")
		}
		if fs.Scope != ScopeVolume {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeVolume)
		}
		if fs.Type != TypeSize {
			t.Errorf("Type = %q, want %q", fs.Type, TypeSize)
		}
	})

	t.Run("network fields", func(t *testing.T) {
		fs, ok := spec.Get("bosun.network.priority")
		if !ok {
			t.Fatal("missing bosun.network.priority")
		}
		if fs.Scope != ScopeNetwork {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeNetwork)
		}
		if fs.Type != TypeInt {
			t.Errorf("Type = %q, want %q", fs.Type, TypeInt)
		}
	})

	// Verify Scopes() grouping
	t.Run("scopes grouping", func(t *testing.T) {
		scopes := spec.Scopes()
		if len(scopes[ScopeGlobal]) != 1 {
			t.Errorf("Global scope has %d fields, want 1", len(scopes[ScopeGlobal]))
		}
		if len(scopes[ScopeContainer]) != 4 {
			t.Errorf("Container scope has %d fields, want 4", len(scopes[ScopeContainer]))
		}
		if len(scopes[ScopeVolume]) != 2 {
			t.Errorf("Volume scope has %d fields, want 2", len(scopes[ScopeVolume]))
		}
		if len(scopes[ScopeNetwork]) != 1 {
			t.Errorf("Network scope has %d fields, want 1", len(scopes[ScopeNetwork]))
		}
	})
}

func TestConfigV1_DefaultOf(t *testing.T) {
	cfg, err := DefaultOf[ConfigV1]()
	if err != nil {
		t.Fatalf("DefaultOf[ConfigV1]() error: %v", err)
	}

	// Verify global defaults
	t.Run("global defaults", func(t *testing.T) {
		// Instance has no default, should be empty
		if cfg.Instance != "" {
			t.Errorf("Instance = %q, want empty", cfg.Instance)
		}
	})

	// Verify container defaults
	t.Run("container defaults", func(t *testing.T) {
		if cfg.StopGracePeriod != 30*time.Second {
			t.Errorf("StopGracePeriod = %v, want %v", cfg.StopGracePeriod, 30*time.Second)
		}
		if cfg.HealthCheckInterval != 30*time.Second {
			t.Errorf("HealthCheckInterval = %v, want %v", cfg.HealthCheckInterval, 30*time.Second)
		}
		if cfg.AutoRestart != true {
			t.Errorf("AutoRestart = %v, want true", cfg.AutoRestart)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
		}
	})

	// Verify volume defaults
	t.Run("volume defaults", func(t *testing.T) {
		if cfg.BackupEnabled != false {
			t.Errorf("BackupEnabled = %v, want false", cfg.BackupEnabled)
		}
		expectedSize := int64(10 * 1024 * 1024 * 1024) // 10GB
		if cfg.MaxSize != expectedSize {
			t.Errorf("MaxSize = %v, want %v", cfg.MaxSize, expectedSize)
		}
	})

	// Verify network defaults
	t.Run("network defaults", func(t *testing.T) {
		if cfg.Priority != 100 {
			t.Errorf("Priority = %v, want 100", cfg.Priority)
		}
	})
}

func TestV1Spec(t *testing.T) {
	// Should not panic
	spec := V1Spec()
	if len(spec) == 0 {
		t.Error("V1Spec() returned empty spec")
	}
}

func TestV1Defaults(t *testing.T) {
	// Should not panic
	cfg := V1Defaults()

	// Verify a default is set
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}
