//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	// TODO find a more robust way to ensure containers are ready
	// Give containers time to start and be labeled
	time.Sleep(5 * time.Second)

	// Test HostPort utility
	port := testutil.HostPort(t, ctx, stack.Project, "test-service", 80)
	log.Printf("Service available on port %d", port)

	// Verify service is responding
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	log.Printf("HTTP check passed")

	// Dump logs for demonstration (normally used on failure)
	logDir := t.TempDir()
	testutil.DumpLogs(t, ctx, stack.Project, logDir)
	log.Printf("Logs dumped to %s", logDir)

	log.Printf("Integration smoke test completed successfully with project: %s", stack.Project)
}
