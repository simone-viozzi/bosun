package planner

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// Planner generates execution plans for discovered jobs.
// The planner is a pure function - it takes a job and returns an execution plan
// with no side effects or Docker API calls.
type Planner struct{}

// New creates a new Planner instance.
func New() *Planner {
	return &Planner{}
}

// Plan generates an ExecutionPlan for the given job.
// The plan is deterministic: same job input produces identical output.
// Steps are ordered: stop → run-worker.
func (p *Planner) Plan(ctx context.Context, job jobs.Job) (jobs.ExecutionPlan, error) {
	select {
	case <-ctx.Done():
		return jobs.ExecutionPlan{}, ctx.Err()
	default:
	}

	var steps []jobs.PlanStep

	// Sort containers for deterministic output
	targetContainers := make([]string, len(job.TargetContainers))
	copy(targetContainers, job.TargetContainers)
	sort.Strings(targetContainers)

	// Sort volumes for deterministic output
	attachVolumes := make([]jobs.VolumeAttachment, len(job.AttachVolumes))
	copy(attachVolumes, job.AttachVolumes)
	sort.Slice(attachVolumes, func(i, j int) bool {
		return attachVolumes[i].Name < attachVolumes[j].Name
	})

	// Determine compose usage once for both stop and start steps.
	// TODO: DESIGN ISSUE - Current logic is simplistic: assumes useCompose=true if
	// len(TargetStacks)==1, but doesn't verify that ALL target containers actually
	// belong to that stack AND are ALL containers in the stack.
	// Correct logic: "compose stop" only if all containers belong to same stack
	// AND they are ALL containers in that stack. Otherwise, stop individually
	// while being mindful of container interdependencies.
	useCompose := len(job.TargetStacks) == 1
	composeProject := ""
	if useCompose {
		composeProject = job.TargetStacks[0]
	}

	// Step 1: Stop containers (if any)
	if len(targetContainers) > 0 {
		containerNames := extractContainerNames(targetContainers)

		stopStep := jobs.PlanStep{
			Type:           jobs.StepTypeStopContainers,
			Description:    generateStopDescription(containerNames, useCompose, composeProject),
			ContainerIDs:   targetContainers,
			ContainerNames: containerNames,
			UseCompose:     useCompose,
			ComposeProject: composeProject,
		}
		steps = append(steps, stopStep)
	}

	// Step 2: Run worker container
	runWorkerStep := jobs.PlanStep{
		Type:         jobs.StepTypeRunWorker,
		Description:  generateRunWorkerDescription(job.WorkerImage, attachVolumes),
		WorkerImage:  job.WorkerImage,
		VolumeMounts: attachVolumes,
	}
	steps = append(steps, runWorkerStep)

	// Step 3: Start containers (if any were stopped)
	if len(targetContainers) > 0 {
		containerNames := extractContainerNames(targetContainers)

		// Reuse useCompose decision from stop step to ensure consistency
		startStep := jobs.PlanStep{
			Type:           jobs.StepTypeStartContainers,
			Description:    generateStartDescription(containerNames, useCompose, composeProject),
			ContainerIDs:   targetContainers,
			ContainerNames: containerNames,
			UseCompose:     useCompose,
			ComposeProject: composeProject,
		}
		steps = append(steps, startStep)
	}

	plan := jobs.ExecutionPlan{
		JobName:   job.Name,
		Steps:     steps,
		CreatedAt: time.Now().UTC(),
	}

	return plan, nil
}

// extractContainerNames extracts human-readable names from container IDs.
// In a real implementation, these would come from the snapshot metadata.
// For now, we use truncated IDs as placeholders.
func extractContainerNames(containerIDs []string) []string {
	names := make([]string, len(containerIDs))
	for i, id := range containerIDs {
		// Use truncated ID as name placeholder
		// In actual use, the names would come from the snapshot
		if len(id) > 12 {
			names[i] = id[:12]
		} else {
			names[i] = id
		}
	}
	return names
}

// generateStopDescription creates a human-readable description for a stop step.
func generateStopDescription(containerNames []string, useComposeStop bool, composeProject string) string {
	if len(containerNames) == 0 {
		return "No containers to stop"
	}

	if useComposeStop && composeProject != "" {
		return fmt.Sprintf("Stop stack %q using docker compose stop (%d container(s))",
			composeProject, len(containerNames))
	}

	if len(containerNames) == 1 {
		return fmt.Sprintf("Stop container %q", containerNames[0])
	}

	// List first few containers, then summarize
	maxShow := 3
	if len(containerNames) <= maxShow {
		return fmt.Sprintf("Stop %d containers: %s",
			len(containerNames), strings.Join(containerNames, ", "))
	}

	shown := strings.Join(containerNames[:maxShow], ", ")
	return fmt.Sprintf("Stop %d containers: %s, and %d more",
		len(containerNames), shown, len(containerNames)-maxShow)
}

// generateRunWorkerDescription creates a human-readable description for a run-worker step.
func generateRunWorkerDescription(workerImage string, volumes []jobs.VolumeAttachment) string {
	if len(volumes) == 0 {
		return fmt.Sprintf("Run worker container %q", workerImage)
	}

	if len(volumes) == 1 {
		return fmt.Sprintf("Run worker %q with volume %q mounted at %s (%s)",
			workerImage, volumes[0].Name, volumes[0].MountPath, volumes[0].Mode)
	}

	return fmt.Sprintf("Run worker %q with %d volumes attached",
		workerImage, len(volumes))
}

// generateStartDescription creates a human-readable description for a start step.
func generateStartDescription(containerNames []string, useComposeStart bool, composeProject string) string {
	if len(containerNames) == 0 {
		return "No containers to start"
	}

	if useComposeStart && composeProject != "" {
		return fmt.Sprintf("Start stack %q using docker compose start (%d container(s))",
			composeProject, len(containerNames))
	}

	if len(containerNames) == 1 {
		return fmt.Sprintf("Start container %q", containerNames[0])
	}

	// List first few containers, then summarize
	maxShow := 3
	if len(containerNames) <= maxShow {
		return fmt.Sprintf("Start %d containers: %s",
			len(containerNames), strings.Join(containerNames, ", "))
	}

	shown := strings.Join(containerNames[:maxShow], ", ")
	return fmt.Sprintf("Start %d containers: %s, and %d more",
		len(containerNames), shown, len(containerNames)-maxShow)
}

// Ensure Planner implements JobPlanner interface.
var _ ports.JobPlanner = (*Planner)(nil)
