package joblabels

import (
	"context"
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"

	"github.com/simone-viozzi/bosun/internal/config/schema"
	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// Stack labels used for job grouping.
const (
	LabelStack          = "bosun.stack"
	LabelComposeProject = "com.docker.compose.project"
)

// Discoverer implements ports.JobDiscoverer by extracting job definitions
// from Docker labels on containers and volumes.
type Discoverer struct {
	cronParser cron.Parser
}

// NewDiscoverer creates a new job label discoverer.
func NewDiscoverer() *Discoverer {
	return &Discoverer{
		// Use standard 5-field cron parser (minute hour dom month dow)
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// DiscoverJobs extracts job definitions from a label snapshot.
// It returns discovered jobs, any validation errors encountered, and any fatal error.
func (d *Discoverer) DiscoverJobs(ctx context.Context, snapshot labels.Snapshot) ([]jobs.Job, []ports.ValidationError, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	var validationErrors []ports.ValidationError

	// Phase 1: Collect job definitions from containers with bosun.job.enabled=true
	jobBuilders := make(map[string]*jobBuilder)

	for _, entity := range snapshot.Entities {
		if entity.Kind != labels.KindContainer {
			continue
		}

		// Check if job is enabled on this container
		if !isJobEnabled(entity.Labels) {
			continue
		}

		// Get job name - required when enabled
		jobName := entity.Labels[schema.LabelJobName]
		if jobName == "" {
			validationErrors = append(validationErrors, ports.ValidationError{
				EntityKind: string(entity.Kind),
				EntityID:   entity.ID,
				EntityName: entity.Name,
				Field:      schema.LabelJobName,
				Message:    "job name is required when bosun.job.enabled=true",
			})
			continue
		}

		// Get or create job builder
		builder, exists := jobBuilders[jobName]
		if !exists {
			builder = &jobBuilder{
				name:             jobName,
				sourceContainers: make([]string, 0),
				targetContainers: make([]string, 0),
				stacks:           make(map[string]struct{}),
			}
			jobBuilders[jobName] = builder
		}

		// Track source container
		builder.sourceContainers = append(builder.sourceContainers, entity.ID)
		builder.targetContainers = append(builder.targetContainers, entity.ID)

		// Resolve stack name (bosun.stack > com.docker.compose.project)
		stackName := resolveStackName(entity)
		if stackName != "" {
			builder.stacks[stackName] = struct{}{}
		}

		// Merge schedule (validate and detect conflicts)
		if schedule := entity.Labels[schema.LabelJobSchedule]; schedule != "" {
			if err := d.validateCronExpression(schedule); err != nil {
				validationErrors = append(validationErrors, ports.ValidationError{
					EntityKind: string(entity.Kind),
					EntityID:   entity.ID,
					EntityName: entity.Name,
					Field:      schema.LabelJobSchedule,
					Message:    fmt.Sprintf("invalid cron expression %q: %v", schedule, err),
				})
			} else if builder.schedule != "" && builder.schedule != schedule {
				validationErrors = append(validationErrors, ports.ValidationError{
					EntityKind: string(entity.Kind),
					EntityID:   entity.ID,
					EntityName: entity.Name,
					Field:      schema.LabelJobSchedule,
					Message:    fmt.Sprintf("conflicting schedule %q (previously %q)", schedule, builder.schedule),
				})
			} else {
				builder.schedule = schedule
			}
		}

		// Merge worker image (detect conflicts)
		if workerImage := entity.Labels[schema.LabelJobWorkerImage]; workerImage != "" {
			if builder.workerImage != "" && builder.workerImage != workerImage {
				validationErrors = append(validationErrors, ports.ValidationError{
					EntityKind: string(entity.Kind),
					EntityID:   entity.ID,
					EntityName: entity.Name,
					Field:      schema.LabelJobWorkerImage,
					Message:    fmt.Sprintf("conflicting worker image %q (previously %q)", workerImage, builder.workerImage),
				})
			} else {
				builder.workerImage = workerImage
			}
		}

		// Merge overlap policy (detect conflicts)
		if overlapPolicy := entity.Labels[schema.LabelJobOverlapPolicy]; overlapPolicy != "" {
			// Validate the overlap policy value before merging.
			if err := jobs.ValidateOverlapPolicy(jobs.OverlapPolicy(overlapPolicy)); err != nil {
				validationErrors = append(validationErrors, ports.ValidationError{
					EntityKind: string(entity.Kind),
					EntityID:   entity.ID,
					EntityName: entity.Name,
					Field:      schema.LabelJobOverlapPolicy,
					Message:    err.Error(),
				})
			} else if builder.overlapPolicy != "" && builder.overlapPolicy != overlapPolicy {
				validationErrors = append(validationErrors, ports.ValidationError{
					EntityKind: string(entity.Kind),
					EntityID:   entity.ID,
					EntityName: entity.Name,
					Field:      schema.LabelJobOverlapPolicy,
					Message:    fmt.Sprintf("conflicting overlap policy %q (previously %q)", overlapPolicy, builder.overlapPolicy),
				})
			} else {
				builder.overlapPolicy = overlapPolicy
			}
		}
	}

	// Phase 2: Attach volumes to jobs
	volumeAttachments := make(map[string][]jobs.VolumeAttachment) // jobName -> []VolumeAttachment

	for _, entity := range snapshot.Entities {
		if entity.Kind != labels.KindVolume {
			continue
		}

		attachTo := entity.Labels[schema.LabelJobAttach]
		if attachTo == "" {
			continue
		}

		// Check if the referenced job exists
		if _, exists := jobBuilders[attachTo]; !exists {
			validationErrors = append(validationErrors, ports.ValidationError{
				EntityKind: string(entity.Kind),
				EntityID:   entity.ID,
				EntityName: entity.Name,
				Field:      schema.LabelJobAttach,
				Message:    fmt.Sprintf("volume references unknown job %q", attachTo),
			})
			continue
		}

		mountPath := entity.Labels[schema.LabelJobMountPath]
		if mountPath == "" {
			// Default mount path based on volume name
			mountPath = "/mnt/" + entity.Name
		}

		mode := entity.Labels[schema.LabelJobMountMode]
		normalizedMode, valid := schema.NormalizeMountMode(mode)
		if mode == "" {
			normalizedMode = schema.DefaultJobMountMode()
		} else if !valid {
			validationErrors = append(validationErrors, ports.ValidationError{
				EntityKind: string(entity.Kind),
				EntityID:   entity.ID,
				EntityName: entity.Name,
				Field:      schema.LabelJobMountMode,
				Message:    fmt.Sprintf("invalid mount mode %q, must be 'ro' or 'rw'", mode),
			})
			normalizedMode = schema.DefaultJobMountMode()
		}

		volumeAttachments[attachTo] = append(volumeAttachments[attachTo], jobs.VolumeAttachment{
			Name:      entity.Name,
			MountPath: mountPath,
			Mode:      normalizedMode,
		})
	}

	// Phase 3: Build final jobs with defaults
	result := make([]jobs.Job, 0, len(jobBuilders))

	for name, builder := range jobBuilders {
		// Apply defaults
		schedule := builder.schedule
		if schedule == "" {
			schedule = schema.DefaultJobSchedule()
		}

		workerImage := builder.workerImage
		if workerImage == "" {
			workerImage = schema.DefaultJobWorkerImage()
		}

		// Collect stack names as slice
		targetStacks := make([]string, 0, len(builder.stacks))
		for stack := range builder.stacks {
			targetStacks = append(targetStacks, stack)
		}

		job := jobs.Job{
			Name:             name,
			Schedule:         schedule,
			OverlapPolicy:    jobs.ParseOverlapPolicy(builder.overlapPolicy),
			Enabled:          true, // Only enabled containers reach this point.
			TargetContainers: builder.targetContainers,
			TargetStacks:     targetStacks,
			WorkerImage:      workerImage,
			AttachVolumes:    volumeAttachments[name],
			SourceContainers: builder.sourceContainers,
		}

		result = append(result, job)
	}

	return result, validationErrors, nil
}

// jobBuilder accumulates job configuration from multiple containers.
type jobBuilder struct {
	name             string
	schedule         string
	overlapPolicy    string
	workerImage      string
	sourceContainers []string
	targetContainers []string
	stacks           map[string]struct{}
}

// isJobEnabled checks if the bosun.job.enabled label is set to true.
func isJobEnabled(lbls map[string]string) bool {
	val, ok := lbls[schema.LabelJobEnabled]
	if !ok {
		return false
	}
	return strings.EqualFold(val, "true") || val == "1"
}

// resolveStackName determines the stack name from entity metadata.
// Priority: bosun.stack label > com.docker.compose.project metadata
func resolveStackName(entity labels.LabeledEntity) string {
	// First check for explicit bosun.stack label
	if stack := entity.Labels[LabelStack]; stack != "" {
		return stack
	}

	// Fall back to compose project from metadata
	if project := entity.Meta[LabelComposeProject]; project != "" {
		return project
	}

	// Check compose.project in metadata (our normalized key)
	if project := entity.Meta["compose.project"]; project != "" {
		return project
	}

	return ""
}

// validateCronExpression validates a cron expression.
func (d *Discoverer) validateCronExpression(expr string) error {
	_, err := d.cronParser.Parse(expr)
	return err
}

// Ensure Discoverer implements JobDiscoverer interface.
var _ ports.JobDiscoverer = (*Discoverer)(nil)
