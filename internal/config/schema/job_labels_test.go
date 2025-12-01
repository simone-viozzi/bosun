package schema

import (
	"testing"
)

func TestJobLabelConfigParsing(t *testing.T) {
	spec := JobSpec()

	t.Run("contains all expected keys", func(t *testing.T) {
		expectedKeys := []string{
			"bosun.job.enabled",
			"bosun.job.name",
			"bosun.job.schedule",
			"bosun.job.worker.image",
			"bosun.job.attach",
			"bosun.job.mount.path",
			"bosun.job.mount.mode",
		}

		for _, key := range expectedKeys {
			if _, ok := spec.Get(key); !ok {
				t.Errorf("missing expected key %q", key)
			}
		}
	})

	t.Run("bosun.job.enabled has correct spec", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.enabled")
		if !ok {
			t.Fatal("missing bosun.job.enabled")
		}

		if fs.Scope != ScopeContainer {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeContainer)
		}
		if fs.Type != TypeBool {
			t.Errorf("Type = %q, want %q", fs.Type, TypeBool)
		}
		if fs.Default != "false" {
			t.Errorf("Default = %q, want %q", fs.Default, "false")
		}
	})

	t.Run("bosun.job.name has correct spec", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.name")
		if !ok {
			t.Fatal("missing bosun.job.name")
		}

		if fs.Scope != ScopeContainer {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeContainer)
		}
		if fs.Type != TypeString {
			t.Errorf("Type = %q, want %q", fs.Type, TypeString)
		}
		// Name has no default - it's required
		if fs.Default != "" {
			t.Errorf("Default = %q, want empty (required field)", fs.Default)
		}
	})

	t.Run("bosun.job.schedule has default midnight", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.schedule")
		if !ok {
			t.Fatal("missing bosun.job.schedule")
		}

		if fs.Scope != ScopeContainer {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeContainer)
		}
		if fs.Type != TypeString {
			t.Errorf("Type = %q, want %q", fs.Type, TypeString)
		}
		if fs.Default != "0 0 * * *" {
			t.Errorf("Default = %q, want %q", fs.Default, "0 0 * * *")
		}
	})

	t.Run("bosun.job.worker.image has correct spec", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.worker.image")
		if !ok {
			t.Fatal("missing bosun.job.worker.image")
		}

		if fs.Scope != ScopeContainer {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeContainer)
		}
		if fs.Default != "bosun-worker:local" {
			t.Errorf("Default = %q, want %q", fs.Default, "bosun-worker:local")
		}
	})

	t.Run("bosun.job.attach has correct scope", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.attach")
		if !ok {
			t.Fatal("missing bosun.job.attach")
		}

		if fs.Scope != ScopeVolume {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeVolume)
		}
		if fs.Type != TypeString {
			t.Errorf("Type = %q, want %q", fs.Type, TypeString)
		}
	})

	t.Run("bosun.job.mount.path has correct spec", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.mount.path")
		if !ok {
			t.Fatal("missing bosun.job.mount.path")
		}

		if fs.Scope != ScopeVolume {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeVolume)
		}
		if fs.Type != TypeString {
			t.Errorf("Type = %q, want %q", fs.Type, TypeString)
		}
	})

	t.Run("bosun.job.mount.mode is enum with ro default", func(t *testing.T) {
		fs, ok := spec.Get("bosun.job.mount.mode")
		if !ok {
			t.Fatal("missing bosun.job.mount.mode")
		}

		if fs.Scope != ScopeVolume {
			t.Errorf("Scope = %q, want %q", fs.Scope, ScopeVolume)
		}
		if fs.Type != TypeEnum {
			t.Errorf("Type = %q, want %q", fs.Type, TypeEnum)
		}
		if fs.Default != "ro" {
			t.Errorf("Default = %q, want %q", fs.Default, "ro")
		}

		// Check enum values
		hasRO := false
		hasRW := false
		for _, v := range fs.Enum {
			if v == "ro" {
				hasRO = true
			}
			if v == "rw" {
				hasRW = true
			}
		}
		if !hasRO {
			t.Error("Enum missing 'ro' value")
		}
		if !hasRW {
			t.Error("Enum missing 'rw' value")
		}
	})
}

func TestJobSpecImmutability(t *testing.T) {
	// Calling JobSpec multiple times should return consistent specs
	spec1 := JobSpec()
	spec2 := JobSpec()

	if len(spec1) != len(spec2) {
		t.Errorf("Spec length mismatch: %d != %d", len(spec1), len(spec2))
	}

	for key := range spec1 {
		if _, ok := spec2[key]; !ok {
			t.Errorf("Key %q present in first spec but not second", key)
		}
	}
}

func TestJobSpecByScope(t *testing.T) {
	spec := JobSpec()

	scopes := spec.Scopes()
	containerKeys := scopes[ScopeContainer]
	volumeKeys := scopes[ScopeVolume]

	// Container scope should have: enabled, name, schedule, worker.image
	if len(containerKeys) != 4 {
		t.Errorf("Container scope keys = %d, want 4", len(containerKeys))
	}

	// Volume scope should have: attach, mount.path, mount.mode
	if len(volumeKeys) != 3 {
		t.Errorf("Volume scope keys = %d, want 3", len(volumeKeys))
	}
}

func TestJobLabelConstants(t *testing.T) {
	t.Run("label key constants match struct tags", func(t *testing.T) {
		spec := JobSpec()

		// Container labels
		if _, ok := spec.Get(LabelJobEnabled); !ok {
			t.Errorf("LabelJobEnabled %q not found in JobSpec", LabelJobEnabled)
		}
		if _, ok := spec.Get(LabelJobName); !ok {
			t.Errorf("LabelJobName %q not found in JobSpec", LabelJobName)
		}
		if _, ok := spec.Get(LabelJobSchedule); !ok {
			t.Errorf("LabelJobSchedule %q not found in JobSpec", LabelJobSchedule)
		}
		if _, ok := spec.Get(LabelJobWorkerImage); !ok {
			t.Errorf("LabelJobWorkerImage %q not found in JobSpec", LabelJobWorkerImage)
		}

		// Volume labels
		if _, ok := spec.Get(LabelJobAttach); !ok {
			t.Errorf("LabelJobAttach %q not found in JobSpec", LabelJobAttach)
		}
		if _, ok := spec.Get(LabelJobMountPath); !ok {
			t.Errorf("LabelJobMountPath %q not found in JobSpec", LabelJobMountPath)
		}
		if _, ok := spec.Get(LabelJobMountMode); !ok {
			t.Errorf("LabelJobMountMode %q not found in JobSpec", LabelJobMountMode)
		}
	})
}

func TestDefaultJobValues(t *testing.T) {
	t.Run("DefaultJobSchedule matches struct tag", func(t *testing.T) {
		spec := JobSpec()
		fs, _ := spec.Get(LabelJobSchedule)
		if DefaultJobSchedule() != fs.Default {
			t.Errorf("DefaultJobSchedule() = %q, spec.Default = %q",
				DefaultJobSchedule(), fs.Default)
		}
	})

	t.Run("DefaultJobWorkerImage matches struct tag", func(t *testing.T) {
		spec := JobSpec()
		fs, _ := spec.Get(LabelJobWorkerImage)
		if DefaultJobWorkerImage() != fs.Default {
			t.Errorf("DefaultJobWorkerImage() = %q, spec.Default = %q",
				DefaultJobWorkerImage(), fs.Default)
		}
	})

	t.Run("DefaultJobMountMode matches struct tag", func(t *testing.T) {
		spec := JobSpec()
		fs, _ := spec.Get(LabelJobMountMode)
		if DefaultJobMountMode() != fs.Default {
			t.Errorf("DefaultJobMountMode() = %q, spec.Default = %q",
				DefaultJobMountMode(), fs.Default)
		}
	})
}

func TestNormalizeMountMode(t *testing.T) {
	tests := []struct {
		input     string
		wantValue string
		wantValid bool
	}{
		// Valid lowercase
		{"ro", "ro", true},
		{"rw", "rw", true},

		// Valid uppercase (case-insensitive)
		{"RO", "ro", true},
		{"RW", "rw", true},

		// Valid mixed case
		{"Ro", "ro", true},
		{"rW", "rw", true},

		// Whitespace handling
		{"  ro  ", "ro", true},
		{"  RW  ", "rw", true},

		// Invalid values
		{"", "", false},
		{"read-only", "", false},
		{"readwrite", "", false},
		{"r", "", false},
		{"w", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotValue, gotValid := NormalizeMountMode(tt.input)
			if gotValue != tt.wantValue || gotValid != tt.wantValid {
				t.Errorf("NormalizeMountMode(%q) = (%q, %v), want (%q, %v)",
					tt.input, gotValue, gotValid, tt.wantValue, tt.wantValid)
			}
		})
	}
}
