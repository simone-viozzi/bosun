package loader

import (
	"testing"

	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

func TestValidateJobLabels_ValidContainer(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled":      "true",
				"bosun.job.name":         "backup-job",
				"bosun.job.schedule":     "0 2 * * *",
				"bosun.job.worker.image": "backup:latest",
			},
		},
	}

	result := ValidateJobLabels(entities)

	if !result.IsValid() {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateJobLabels_InvalidEnabled(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled": "maybe",
				"bosun.job.name":    "backup-job",
			},
		},
	}

	result := ValidateJobLabels(entities)

	if result.IsValid() {
		t.Fatal("Expected invalid result for 'maybe' as enabled value")
	}

	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Code != JobErrorInvalidEnabled {
		t.Errorf("Expected code %q, got %q", JobErrorInvalidEnabled, result.Errors[0].Code)
	}
}

func TestValidateJobLabels_InvalidSchedule(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled":  "true",
				"bosun.job.name":     "backup-job",
				"bosun.job.schedule": "not a cron",
			},
		},
	}

	result := ValidateJobLabels(entities)

	if result.IsValid() {
		t.Fatal("Expected invalid result for bad cron expression")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == JobErrorInvalidSchedule {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected JobErrorInvalidSchedule error")
	}
}

func TestValidateJobLabels_MissingName(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled":  "true",
				"bosun.job.schedule": "0 2 * * *",
				// Missing bosun.job.name
			},
		},
	}

	result := ValidateJobLabels(entities)

	if result.IsValid() {
		t.Fatal("Expected invalid result for missing name when enabled=true")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == JobErrorMissingName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected JobErrorMissingName error")
	}
}

func TestValidateJobLabels_ConflictingSchedule(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app1",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled":  "true",
				"bosun.job.name":     "backup-job",
				"bosun.job.schedule": "0 2 * * *",
			},
		},
		{
			Kind: dlabels.KindContainer,
			Name: "app2",
			ID:   "container2",
			Labels: map[string]string{
				"bosun.job.enabled":  "true",
				"bosun.job.name":     "backup-job",
				"bosun.job.schedule": "0 3 * * *", // Different schedule, same job
			},
		},
	}

	result := ValidateJobLabels(entities)

	if result.IsValid() {
		t.Fatal("Expected invalid result for conflicting schedules")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == JobErrorConflictingField {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected JobErrorConflictingField error")
	}
}

func TestValidateJobLabels_ValidVolume(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled": "true",
				"bosun.job.name":    "backup-job",
			},
		},
		{
			Kind: dlabels.KindVolume,
			Name: "data-vol",
			ID:   "vol1",
			Labels: map[string]string{
				"bosun.job.attach":     "backup-job",
				"bosun.job.mount.path": "/data",
			},
		},
	}

	result := ValidateJobLabels(entities)

	if !result.IsValid() {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Warnings) > 0 {
		t.Errorf("Expected no warnings, got: %v", result.Warnings)
	}
}

func TestValidateJobLabels_VolumeAttachIsJobName(t *testing.T) {
	t.Parallel()

	// bosun.job.attach is the job name to attach to, not a boolean
	// Any string value is valid - it references a job name
	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled": "true",
				"bosun.job.name":    "backup-job",
			},
		},
		{
			Kind: dlabels.KindVolume,
			Name: "data-vol",
			ID:   "vol1",
			Labels: map[string]string{
				"bosun.job.attach": "backup-job", // Job name, not boolean
			},
		},
	}

	result := ValidateJobLabels(entities)

	if !result.IsValid() {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateJobLabels_VolumeOrphanedWithNoMatchingJob(t *testing.T) {
	t.Parallel()

	// Volume references a job that doesn't exist - should be a warning
	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindVolume,
			Name: "data-vol",
			ID:   "vol1",
			Labels: map[string]string{
				"bosun.job.attach": "nonexistent-job", // Job name that doesn't exist
			},
		},
	}

	result := ValidateJobLabels(entities)

	// Orphaned volumes are warnings, not errors
	if !result.IsValid() {
		t.Errorf("Expected valid result (orphan is a warning), got errors: %v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning for orphaned volume")
	}
}

func TestValidateJobLabels_InvalidMountPath(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled": "true",
				"bosun.job.name":    "backup-job",
			},
		},
		{
			Kind: dlabels.KindVolume,
			Name: "data-vol",
			ID:   "vol1",
			Labels: map[string]string{
				"bosun.job.attach":     "backup-job",
				"bosun.job.mount.path": "data", // Not absolute path
			},
		},
	}

	result := ValidateJobLabels(entities)

	if result.IsValid() {
		t.Fatal("Expected invalid result for non-absolute mount path")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == JobErrorInvalidMountPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected JobErrorInvalidMountPath error")
	}
}

func TestValidateJobLabels_OrphanedVolume(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindVolume,
			Name: "orphan-vol",
			ID:   "vol1",
			Labels: map[string]string{
				"bosun.job.attach":     "nonexistent-job",
				"bosun.job.mount.path": "/data",
			},
		},
	}

	result := ValidateJobLabels(entities)

	// Orphaned volumes are warnings, not errors
	if !result.IsValid() {
		t.Errorf("Expected valid result (orphan is a warning), got errors: %v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning for orphaned volume")
	}
}

func TestValidateJobLabels_NoJobLabels(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.config.log.level": "debug",
			},
		},
	}

	result := ValidateJobLabels(entities)

	if !result.IsValid() {
		t.Errorf("Expected valid result for container with no job labels, got errors: %v", result.Errors)
	}
}

func TestValidateJobLabels_DisabledJob(t *testing.T) {
	t.Parallel()

	entities := []dlabels.LabeledEntity{
		{
			Kind: dlabels.KindContainer,
			Name: "app",
			ID:   "container1",
			Labels: map[string]string{
				"bosun.job.enabled": "false",
				// No name required when disabled
			},
		},
	}

	result := ValidateJobLabels(entities)

	if !result.IsValid() {
		t.Errorf("Expected valid result for disabled job, got errors: %v", result.Errors)
	}
}

func TestValidateJobLabels_BooleanVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"true", "true", true},
		{"false", "false", true},
		{"TRUE", "TRUE", true},
		{"FALSE", "FALSE", true},
		{"1", "1", true},
		{"0", "0", true},
		{"yes", "yes", true},
		{"no", "no", true},
		{"YES", "YES", true},
		{"NO", "NO", true},
		{"invalid", "maybe", false},
		{"empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isValidBoolean(tc.value)
			if result != tc.isValid {
				t.Errorf("isValidBoolean(%q) = %v, want %v", tc.value, result, tc.isValid)
			}
		})
	}
}

func TestJobValidationError_Error(t *testing.T) {
	t.Parallel()

	err := JobValidationError{
		Entity: dlabels.LabeledEntity{
			Kind: dlabels.KindContainer,
			Name: "myapp",
		},
		Field:   "bosun.job.enabled",
		Message: "invalid boolean value",
		Code:    JobErrorInvalidEnabled,
	}

	got := err.Error()
	if got != `container "myapp": invalid boolean value` {
		t.Errorf("Error() = %q, unexpected format", got)
	}
}

func TestJobValidationErrors_Error(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var errs JobValidationErrors
		if errs.Error() != "no errors" {
			t.Errorf("Error() = %q, want %q", errs.Error(), "no errors")
		}
	})

	t.Run("single", func(t *testing.T) {
		t.Parallel()
		errs := JobValidationErrors{
			{
				Entity:  dlabels.LabeledEntity{Kind: dlabels.KindContainer, Name: "app"},
				Message: "test error",
			},
		}
		if errs.Error() != `container "app": test error` {
			t.Errorf("Error() = %q, unexpected format", errs.Error())
		}
	})

	t.Run("multiple", func(t *testing.T) {
		t.Parallel()
		errs := JobValidationErrors{
			{Entity: dlabels.LabeledEntity{Kind: dlabels.KindContainer, Name: "app1"}, Message: "error 1"},
			{Entity: dlabels.LabeledEntity{Kind: dlabels.KindContainer, Name: "app2"}, Message: "error 2"},
		}
		if errs.Error() != "2 job validation errors" {
			t.Errorf("Error() = %q, want %q", errs.Error(), "2 job validation errors")
		}
	})
}
