// Package loader provides configuration loading and validation from Docker labels and files.
package loader

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

// JobValidationError represents a validation error specific to job labels.
type JobValidationError struct {
	Entity  dlabels.LabeledEntity // The entity with the error
	Field   string                // The label field that failed
	Message string                // Human-readable error message
	Code    JobErrorCode          // Machine-readable error code
}

// Error implements the error interface.
func (e JobValidationError) Error() string {
	return fmt.Sprintf("%s %q: %s", e.Entity.Kind, e.Entity.Name, e.Message)
}

// JobErrorCode represents machine-readable error codes for job validation.
type JobErrorCode string

const (
	// JobErrorInvalidEnabled indicates bosun.job.enabled is not a valid boolean.
	JobErrorInvalidEnabled JobErrorCode = "invalid_enabled"
	// JobErrorInvalidSchedule indicates bosun.job.schedule is not a valid cron expression.
	JobErrorInvalidSchedule JobErrorCode = "invalid_schedule"
	// JobErrorMissingName indicates bosun.job.name is missing when enabled=true.
	JobErrorMissingName JobErrorCode = "missing_name"
	// JobErrorConflictingField indicates conflicting values for the same job field.
	JobErrorConflictingField JobErrorCode = "conflicting_field"
	// JobErrorOrphanedVolume indicates a volume references a non-existent job.
	JobErrorOrphanedVolume JobErrorCode = "orphaned_volume"
	// JobErrorInvalidAttach indicates bosun.job.attach is not a valid boolean.
	JobErrorInvalidAttach JobErrorCode = "invalid_attach"
	// JobErrorInvalidMountPath indicates bosun.job.mountpath is invalid.
	JobErrorInvalidMountPath JobErrorCode = "invalid_mount_path"
)

// JobValidationErrors is a collection of job validation errors.
type JobValidationErrors []JobValidationError

// Error implements the error interface.
func (e JobValidationErrors) Error() string {
	if len(e) == 0 {
		return "no errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d job validation errors", len(e))
}

// JobValidationResult holds the outcome of job label validation.
type JobValidationResult struct {
	Errors   JobValidationErrors // All validation errors
	Warnings []string            // Non-fatal warnings (e.g., orphaned volumes)
}

// IsValid returns true if there are no errors.
func (r JobValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidateJobLabels validates job labels across all entities.
// It checks individual label validity and cross-entity consistency.
func ValidateJobLabels(entities []dlabels.LabeledEntity) JobValidationResult {
	result := JobValidationResult{}

	// Track job names and their field values for conflict detection
	jobFields := make(map[string]map[string]fieldValue)    // job name -> field -> first value seen
	jobVolumes := make(map[string][]dlabels.LabeledEntity) // job name -> volumes attached

	for _, entity := range entities {
		switch entity.Kind {
		case dlabels.KindContainer:
			validateContainerJobLabels(entity, &result, jobFields)
		case dlabels.KindVolume:
			validateVolumeJobLabels(entity, &result, jobVolumes)
		}
	}

	// Check for orphaned volumes (volumes attached to non-existent jobs)
	checkOrphanedVolumes(jobFields, jobVolumes, &result)

	return result
}

// fieldValue tracks where a field value was first seen.
type fieldValue struct {
	value  string
	entity dlabels.LabeledEntity
}

// validateContainerJobLabels validates job labels on a container.
func validateContainerJobLabels(entity dlabels.LabeledEntity, result *JobValidationResult, jobFields map[string]map[string]fieldValue) {
	labels := entity.Labels

	// Check if container has any job labels
	var hasJobLabels bool
	for key := range labels {
		if strings.HasPrefix(key, "bosun.job.") {
			hasJobLabels = true
			break
		}
	}
	if !hasJobLabels {
		return
	}

	// Validate bosun.job.enabled (must be valid boolean)
	enabledStr, hasEnabled := labels["bosun.job.enabled"]
	if hasEnabled && !isValidBoolean(enabledStr) {
		result.Errors = append(result.Errors, JobValidationError{
			Entity:  entity,
			Field:   "bosun.job.enabled",
			Message: fmt.Sprintf("invalid boolean value %q, expected 'true' or 'false'", enabledStr),
			Code:    JobErrorInvalidEnabled,
		})
	}

	// Parse key values
	enabled := hasEnabled && isTrue(enabledStr)
	jobName := labels["bosun.job.name"]
	schedule := labels["bosun.job.schedule"]

	// If enabled=true, name must be present
	if enabled && jobName == "" {
		result.Errors = append(result.Errors, JobValidationError{
			Entity:  entity,
			Field:   "bosun.job.name",
			Message: "bosun.job.name is required when bosun.job.enabled=true",
			Code:    JobErrorMissingName,
		})
	}

	// Validate cron schedule if present
	if schedule != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(schedule); err != nil {
			result.Errors = append(result.Errors, JobValidationError{
				Entity:  entity,
				Field:   "bosun.job.schedule",
				Message: fmt.Sprintf("invalid cron expression %q: %v", schedule, err),
				Code:    JobErrorInvalidSchedule,
			})
		}
	}

	// Track field values for conflict detection
	if jobName != "" {
		if jobFields[jobName] == nil {
			jobFields[jobName] = make(map[string]fieldValue)
		}
		fields := jobFields[jobName]

		// Check each field for conflicts
		checkFieldConflict(entity, "schedule", schedule, fields, result)
		checkFieldConflict(entity, "worker.image", labels["bosun.job.worker.image"], fields, result)
		checkFieldConflict(entity, "use_compose_stop", labels["bosun.job.use_compose_stop"], fields, result)
	}
}

// checkFieldConflict checks if a field value conflicts with a previously seen value.
func checkFieldConflict(entity dlabels.LabeledEntity, field, value string, fields map[string]fieldValue, result *JobValidationResult) {
	if value == "" {
		return
	}

	if existing, ok := fields[field]; ok {
		if existing.value != value {
			result.Errors = append(result.Errors, JobValidationError{
				Entity: entity,
				Field:  "bosun.job." + field,
				Message: fmt.Sprintf("conflicting value %q for field %q (previously %q on %s %q)",
					value, field, existing.value, existing.entity.Kind, existing.entity.Name),
				Code: JobErrorConflictingField,
			})
		}
	} else {
		fields[field] = fieldValue{value: value, entity: entity}
	}
}

// validateVolumeJobLabels validates job labels on a volume.
func validateVolumeJobLabels(entity dlabels.LabeledEntity, result *JobValidationResult, jobVolumes map[string][]dlabels.LabeledEntity) {
	labels := entity.Labels

	// Check bosun.job.attach - this is the job name to attach to, not a boolean
	attachJobName, hasAttach := labels["bosun.job.attach"]
	if !hasAttach || attachJobName == "" {
		return
	}

	// Validate mount path if present
	if mountPath, ok := labels["bosun.job.mount.path"]; ok && mountPath != "" {
		if !strings.HasPrefix(mountPath, "/") {
			result.Errors = append(result.Errors, JobValidationError{
				Entity:  entity,
				Field:   "bosun.job.mount.path",
				Message: fmt.Sprintf("mount path %q must be an absolute path (start with /)", mountPath),
				Code:    JobErrorInvalidMountPath,
			})
		}
	}

	// Track for orphan detection - the attach value IS the job name
	jobVolumes[attachJobName] = append(jobVolumes[attachJobName], entity)
}

// checkOrphanedVolumes checks for volumes attached to jobs that don't exist.
func checkOrphanedVolumes(jobFields map[string]map[string]fieldValue, jobVolumes map[string][]dlabels.LabeledEntity, result *JobValidationResult) {
	for jobName, volumes := range jobVolumes {
		if _, exists := jobFields[jobName]; !exists {
			// Job doesn't exist - this is a warning, not an error
			for _, vol := range volumes {
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"volume %q is attached to job %q but no container defines this job",
					vol.Name, jobName))
			}
		}
	}
}

// isValidBoolean checks if a string is a valid boolean representation.
func isValidBoolean(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "false" || s == "1" || s == "0" || s == "yes" || s == "no"
}

// isTrue checks if a string represents a true boolean value.
func isTrue(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}
