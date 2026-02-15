//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
)

var globalCoverDir string

// coverageEnabled returns true if integration test coverage is enabled via env var
func coverageEnabled() bool {
	return os.Getenv("BOSUN_TEST_COVERAGE") == "1"
}

func TestMain(m *testing.M) {
	// Set up global coverage directory if coverage is enabled
	if coverageEnabled() {
		var err error
		globalCoverDir, err = os.MkdirTemp("", "bosun-integration-coverage-*")
		if err != nil {
			panic("Failed to create global coverage directory: " + err.Error())
		}
		// Make the path absolute to avoid issues with working directory changes
		globalCoverDir, err = filepath.Abs(globalCoverDir)
		if err != nil {
			panic("Failed to get absolute path for coverage directory: " + err.Error())
		}

		// For Makefile integration: if GOCOVERDIR is set, use that instead
		if envCoverDir := os.Getenv("GOCOVERDIR"); envCoverDir != "" {
			globalCoverDir = envCoverDir
		}
	}

	code := m.Run()

	// Note: We don't clean up globalCoverDir here because the Makefile needs it
	os.Exit(code)
}
