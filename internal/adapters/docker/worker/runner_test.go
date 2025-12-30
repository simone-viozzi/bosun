package worker

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/mount"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

func TestConvertMounts_Empty(t *testing.T) {
	result := convertMounts(nil)
	if result != nil {
		t.Errorf("convertMounts(nil) = %v, want nil", result)
	}

	result = convertMounts([]ports.VolumeMount{})
	if result != nil {
		t.Errorf("convertMounts([]) = %v, want nil", result)
	}
}

func TestConvertMounts_Single(t *testing.T) {
	mounts := []ports.VolumeMount{
		{Source: "pgdata", Target: "/data/postgres", ReadOnly: true},
	}

	result := convertMounts(mounts)
	if len(result) != 1 {
		t.Fatalf("result length = %d, want 1", len(result))
	}

	m := result[0]
	if m.Type != mount.TypeVolume {
		t.Errorf("Type = %v, want %v", m.Type, mount.TypeVolume)
	}
	if m.Source != "pgdata" {
		t.Errorf("Source = %q, want %q", m.Source, "pgdata")
	}
	if m.Target != "/data/postgres" {
		t.Errorf("Target = %q, want %q", m.Target, "/data/postgres")
	}
	if !m.ReadOnly {
		t.Error("ReadOnly = false, want true")
	}
}

func TestConvertMounts_Multiple(t *testing.T) {
	mounts := []ports.VolumeMount{
		{Source: "pgdata", Target: "/data/postgres", ReadOnly: true},
		{Source: "redis", Target: "/data/redis", ReadOnly: false},
	}

	result := convertMounts(mounts)
	if len(result) != 2 {
		t.Fatalf("result length = %d, want 2", len(result))
	}

	// First mount
	if result[0].Source != "pgdata" {
		t.Errorf("result[0].Source = %q, want %q", result[0].Source, "pgdata")
	}
	if !result[0].ReadOnly {
		t.Error("result[0].ReadOnly = false, want true")
	}

	// Second mount
	if result[1].Source != "redis" {
		t.Errorf("result[1].Source = %q, want %q", result[1].Source, "redis")
	}
	if result[1].ReadOnly {
		t.Error("result[1].ReadOnly = true, want false")
	}
}

func TestDefaultWorkerConfig(t *testing.T) {
	config := ports.DefaultWorkerConfig()

	if config.Env == nil {
		t.Error("Env is nil, want non-nil map")
	}
	if config.Mounts != nil {
		t.Errorf("Mounts = %v, want nil", config.Mounts)
	}
	if config.Timeout != 1*time.Hour {
		t.Errorf("Timeout = %v, want %v", config.Timeout, 1*time.Hour)
	}
	if config.KeepOnFailure {
		t.Error("KeepOnFailure = true, want false")
	}
	if config.DryRun {
		t.Error("DryRun = true, want false")
	}
}

func TestWorkerConfig_BuildEnv(t *testing.T) {
	config := ports.WorkerConfig{
		JobName:   "daily-backup",
		RunID:     "abc123-def456-ghi789",
		StackName: "mystack",
		DryRun:    false,
		Env: map[string]string{
			"CUSTOM_VAR": "custom_value",
		},
	}

	env := config.BuildEnv()

	// Should have BOSUN_* vars plus custom vars
	if len(env) < 5 {
		t.Errorf("env length = %d, want at least 5", len(env))
	}

	// Check for required BOSUN_* environment variables
	hasJobName := false
	hasRunID := false
	hasStack := false
	hasDryRun := false
	hasCustom := false

	for _, e := range env {
		switch e {
		case "BOSUN_JOB_NAME=daily-backup":
			hasJobName = true
		case "BOSUN_RUN_ID=abc123-def456-ghi789":
			hasRunID = true
		case "BOSUN_STACK=mystack":
			hasStack = true
		case "BOSUN_DRY_RUN=false":
			hasDryRun = true
		case "CUSTOM_VAR=custom_value":
			hasCustom = true
		}
	}

	if !hasJobName {
		t.Error("missing BOSUN_JOB_NAME in env")
	}
	if !hasRunID {
		t.Error("missing BOSUN_RUN_ID in env")
	}
	if !hasStack {
		t.Error("missing BOSUN_STACK in env")
	}
	if !hasDryRun {
		t.Error("missing BOSUN_DRY_RUN in env")
	}
	if !hasCustom {
		t.Error("missing CUSTOM_VAR in env")
	}
}

func TestWorkerConfig_BuildEnv_DryRun(t *testing.T) {
	config := ports.WorkerConfig{
		JobName:   "test-job",
		RunID:     "test-run-id",
		StackName: "test-stack",
		DryRun:    true,
		Env:       map[string]string{},
	}

	env := config.BuildEnv()

	hasDryRunTrue := false
	for _, e := range env {
		if e == "BOSUN_DRY_RUN=true" {
			hasDryRunTrue = true
			break
		}
	}

	if !hasDryRunTrue {
		t.Error("BOSUN_DRY_RUN should be 'true' when DryRun=true")
	}
}

func TestWorkerResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   ports.WorkerResult
		expected bool
	}{
		{
			name: "success - exit code 0, not timed out",
			result: ports.WorkerResult{
				ExitCode: 0,
				TimedOut: false,
			},
			expected: true,
		},
		{
			name: "failure - non-zero exit code",
			result: ports.WorkerResult{
				ExitCode: 1,
				TimedOut: false,
			},
			expected: false,
		},
		{
			name: "failure - timed out with exit 0",
			result: ports.WorkerResult{
				ExitCode: 0,
				TimedOut: true,
			},
			expected: false,
		},
		{
			name: "failure - timed out with exit 137 (SIGKILL)",
			result: ports.WorkerResult{
				ExitCode: 137,
				TimedOut: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWorkerContainerNameFormat(t *testing.T) {
	// Test that the name format constant is usable
	name := jobs.WorkerContainerNameFormat
	if name == "" {
		t.Error("WorkerContainerNameFormat is empty")
	}

	// Format should have placeholders
	if !containsString(name, "%s") {
		t.Errorf("WorkerContainerNameFormat = %q, should contain %%s placeholder", name)
	}
}

func TestVolumeMountFields(t *testing.T) {
	vm := ports.VolumeMount{
		Source:   "my-volume",
		Target:   "/data",
		ReadOnly: true,
	}

	if vm.Source != "my-volume" {
		t.Errorf("Source = %q, want %q", vm.Source, "my-volume")
	}
	if vm.Target != "/data" {
		t.Errorf("Target = %q, want %q", vm.Target, "/data")
	}
	if !vm.ReadOnly {
		t.Error("ReadOnly = false, want true")
	}
}

func TestWorkerResultFields(t *testing.T) {
	wr := ports.WorkerResult{
		ExitCode:    0,
		Logs:        "container logs here",
		ContainerID: "abc123",
		Duration:    5 * time.Second,
		TimedOut:    false,
	}

	if wr.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want %d", wr.ExitCode, 0)
	}
	if wr.Logs != "container logs here" {
		t.Errorf("Logs = %q, want %q", wr.Logs, "container logs here")
	}
	if wr.ContainerID != "abc123" {
		t.Errorf("ContainerID = %q, want %q", wr.ContainerID, "abc123")
	}
	if wr.Duration != 5*time.Second {
		t.Errorf("Duration = %v, want %v", wr.Duration, 5*time.Second)
	}
	if wr.TimedOut {
		t.Error("TimedOut = true, want false")
	}
}

func TestGracePeriodConstant(t *testing.T) {
	if jobs.GracePeriod != 10*time.Second {
		t.Errorf("GracePeriod = %v, want %v", jobs.GracePeriod, 10*time.Second)
	}
}

func TestDefaultTimeoutConstants(t *testing.T) {
	if jobs.DefaultStopTimeout != 30*time.Second {
		t.Errorf("DefaultStopTimeout = %v, want %v", jobs.DefaultStopTimeout, 30*time.Second)
	}
	if jobs.DefaultStartTimeout != 30*time.Second {
		t.Errorf("DefaultStartTimeout = %v, want %v", jobs.DefaultStartTimeout, 30*time.Second)
	}
	if jobs.DefaultWorkerTimeout != 1*time.Hour {
		t.Errorf("DefaultWorkerTimeout = %v, want %v", jobs.DefaultWorkerTimeout, 1*time.Hour)
	}
}

func TestEnvVarConstants(t *testing.T) {
	// Verify environment variable constants
	if jobs.EnvJobName != "BOSUN_JOB_NAME" {
		t.Errorf("EnvJobName = %q, want %q", jobs.EnvJobName, "BOSUN_JOB_NAME")
	}
	if jobs.EnvRunID != "BOSUN_RUN_ID" {
		t.Errorf("EnvRunID = %q, want %q", jobs.EnvRunID, "BOSUN_RUN_ID")
	}
	if jobs.EnvStack != "BOSUN_STACK" {
		t.Errorf("EnvStack = %q, want %q", jobs.EnvStack, "BOSUN_STACK")
	}
	if jobs.EnvDryRun != "BOSUN_DRY_RUN" {
		t.Errorf("EnvDryRun = %q, want %q", jobs.EnvDryRun, "BOSUN_DRY_RUN")
	}
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
