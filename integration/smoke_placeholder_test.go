//go:build integration
// +build integration

package integration

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/testutil"
)

// Test_Integration_Smoke_Placeholder verifies that the testutil harness can successfully
// start and stop a Docker Compose stack with unique project naming and automatic cleanup.
// This serves as a basic smoke test for the integration testing infrastructure.
func Test_Integration_Smoke_Placeholder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Printf("Starting integration smoke test")
	// Start a basic compose stack for smoke testing
	stack := testutil.StartCompose(t, ctx, "docker-compose.yaml")
	if stack.Project == "" {
		t.Fatal("expected non-empty project name")
	}

	// Basic validation that the stack started
	if len(stack.Files) == 0 {
		t.Fatal("expected at least one compose file")
	}

	log.Printf("Integration smoke test completed successfully with project: %s", stack.Project)
}
