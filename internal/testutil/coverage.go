//go:build integration

package testutil

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// RunBosunWithCoverage executes the bosun binary with coverage collection enabled.
// If coverDir is empty, runs without coverage (GOCOVERDIR not set).
// Returns the exec.Cmd configured but not started.
func RunBosunWithCoverage(ctx context.Context, binPath, coverDir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binPath, args...)

	if coverDir != "" {
		// Set GOCOVERDIR to collect coverage data
		cmd.Env = append(os.Environ(), "GOCOVERDIR="+coverDir)
	}

	return cmd
}

// BuildBosunBinary builds the bosun binary for testing.
// If withCoverage is true, builds with -cover flag and returns coverage directory.
// Returns binPath and optionally coverDir (empty string if coverage disabled).
func BuildBosunBinary(t *testing.T, withCoverage bool) (binPath, coverDir string) {
	t.Helper()

	// Create temp directory for binary
	tmpDir := t.TempDir()
	binPath = filepath.Join(tmpDir, "bosun")

	// Build the binary
	args := []string{"build", "-o", binPath}
	if withCoverage {
		args = append(args, "-cover")
		coverDir = filepath.Join(tmpDir, "covdata")
		if err := os.MkdirAll(coverDir, 0755); err != nil {
			t.Fatalf("Failed to create coverage directory: %v", err)
		}
	}
	args = append(args, "./cmd/bosun")

	cmd := exec.Command("go", args...)
	cmd.Dir = findProjectRoot(t)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build bosun: %v\n%s", err, output)
	}

	return binPath, coverDir
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Look for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
