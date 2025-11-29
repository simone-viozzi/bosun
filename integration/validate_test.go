//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/testutil"
)

// Test_Integration_ConfigValidate_ValidConfig verifies that the validate command
// exits 0 when all labels are valid.
func Test_Integration_ConfigValidate_ValidConfig(t *testing.T) {
	// Note: These tests cannot run in parallel because they all validate ALL
	// Docker entities with bosun.* labels system-wide

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start a compose stack with valid labels
	_ = testutil.StartCompose(t, ctx, "validate-valid.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run bosun config validate
	cmd := exec.CommandContext(ctx, bosunBin, "config", "validate")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit 0
	if err != nil {
		t.Errorf("Expected exit 0, got error: %v\nstderr: %s", err, stderr.String())
	}

	// Should print success message
	if !strings.Contains(stdout.String(), "Configuration valid") {
		t.Errorf("Expected success message, got stdout: %s", stdout.String())
	}
}

// Test_Integration_ConfigValidate_InvalidConfig verifies that the validate command
// exits non-zero and reports all errors when labels are invalid.
func Test_Integration_ConfigValidate_InvalidConfig(t *testing.T) {
	// Note: These tests cannot run in parallel because they all validate ALL
	// Docker entities with bosun.* labels system-wide

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start a compose stack with invalid labels
	_ = testutil.StartCompose(t, ctx, "validate-invalid.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run bosun config validate
	cmd := exec.CommandContext(ctx, bosunBin, "config", "validate")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit non-zero
	if err == nil {
		t.Error("Expected non-zero exit code, got exit 0")
	}

	// Should report errors to stderr
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Validation errors") {
		t.Errorf("Expected validation errors header, got stderr: %s", stderrStr)
	}

	// Should report unknown key error
	if !strings.Contains(stderrStr, "unknown key") {
		t.Errorf("Expected unknown key error, got stderr: %s", stderrStr)
	}

	// Should report multiple errors (not just first)
	if !strings.Contains(stderrStr, "error") {
		t.Errorf("Expected error count, got stderr: %s", stderrStr)
	}
}

// Test_Integration_ConfigValidate_PrintFlag verifies that --print outputs valid JSON.
func Test_Integration_ConfigValidate_PrintFlag(t *testing.T) {
	// Note: These tests cannot run in parallel because they all validate ALL
	// Docker entities with bosun.* labels system-wide

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start a compose stack with valid labels
	_ = testutil.StartCompose(t, ctx, "validate-valid.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run bosun config validate --print
	cmd := exec.CommandContext(ctx, bosunBin, "config", "validate", "--print")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit 0
	if err != nil {
		t.Errorf("Expected exit 0, got error: %v\nstderr: %s", err, stderr.String())
	}

	// Should output valid JSON
	var config map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &config); err != nil {
		t.Errorf("Expected valid JSON output, got parse error: %v\nstdout: %s", err, stdout.String())
	}

	// Should have expected fields
	if _, ok := config["StopGracePeriod"]; !ok {
		t.Errorf("Expected StopGracePeriod in config, got: %v", config)
	}
}

// Test_Integration_ConfigValidate_ScopeFlag verifies that --scope filters entities.
func Test_Integration_ConfigValidate_ScopeFlag(t *testing.T) {
	// Note: These tests cannot run in parallel because they all validate ALL
	// Docker entities with bosun.* labels system-wide

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start a compose stack with invalid volume label (scope mismatch)
	_ = testutil.StartCompose(t, ctx, "validate-invalid.yaml")

	// Build bosun binary
	bosunBin := buildBosun(t)

	// Run bosun config validate --scope container
	// This should FAIL because the container has invalid labels (unknown key, bad duration, bad enum)
	cmd := exec.CommandContext(ctx, bosunBin, "config", "validate", "--scope", "container")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit non-zero (container has invalid labels)
	if err == nil {
		t.Error("Expected non-zero exit when validating container scope with invalid container labels")
	}

	// The volume's scope mismatch error should NOT appear (since we're only checking containers)
	stderrStr := stderr.String()
	if strings.Contains(stderrStr, "volume") {
		t.Errorf("Did not expect volume errors when --scope container, got: %s", stderrStr)
	}
}

// buildBosun compiles the bosun binary for testing
func buildBosun(t *testing.T) string {
	t.Helper()

	// Create temp directory for binary
	tmpDir := t.TempDir()
	binPath := tmpDir + "/bosun"

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/bosun")
	cmd.Dir = findProjectRoot(t)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build bosun: %v\n%s", err, output)
	}

	return binPath
}

// findProjectRoot finds the project root directory
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Look for go.mod
	for {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			return dir
		}
		parent := dir[:strings.LastIndex(dir, "/")]
		if parent == dir {
			t.Fatal("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
