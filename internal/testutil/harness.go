//go:build integration

// Package testutil provides test utilities for integration tests.
// It uses the docker compose CLI directly instead of the Go library
// to avoid dependency conflicts with the Docker/Moby module split.
package testutil

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gosimple/slug"
)

// TODO use testcontainers-go to up / down compose stacks. do not rely on CLI.

//go:embed compose/*.yaml
var ComposeFS embed.FS

// Stack represents a running Docker Compose stack for testing.
type Stack struct {
	Project    string
	ComposeDir string
	Files      []string
	T          *testing.T
}

// StartCompose starts a Docker Compose stack using the CLI.
// It creates a temporary directory with the compose files and runs `docker compose up`.
// The stack is automatically cleaned up when the test completes.
func StartCompose(t *testing.T, ctx context.Context, files ...string) *Stack {
	t.Helper()

	project := slug.Make(fmt.Sprintf("bosun-%s-%d", t.Name(), time.Now().UnixNano()))

	// Create temp directory for compose files
	tmpDir, err := os.MkdirTemp("", "bosun-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	// Write embedded compose files to temp directory
	for _, f := range files {
		content, err := ComposeFS.ReadFile(filepath.Join("compose", f))
		if err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("read compose %s: %v", f, err)
		}
		destPath := filepath.Join(tmpDir, f)
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("write compose %s: %v", f, err)
		}
	}

	st := &Stack{
		Project:    project,
		ComposeDir: tmpDir,
		Files:      files,
		T:          t,
	}

	// Build compose command args
	args := st.composeArgs("up", "-d", "--wait")

	// Run docker compose up
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = tmpDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	t.Logf("Starting compose stack %s with files: %v", project, files)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("docker compose up failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Small delay to ensure everything is ready
	time.Sleep(500 * time.Millisecond)

	// Register cleanup
	t.Cleanup(func() {
		st.Down(context.Background())
		os.RemoveAll(tmpDir)
	})

	return st
}

// composeArgs builds the docker compose command arguments.
func (s *Stack) composeArgs(subcommand string, extraArgs ...string) []string {
	args := []string{"compose", "-p", s.Project}
	for _, f := range s.Files {
		args = append(args, "-f", f)
	}
	args = append(args, subcommand)
	args = append(args, extraArgs...)
	return args
}

// Down stops and removes the compose stack.
func (s *Stack) Down(ctx context.Context) error {
	args := s.composeArgs("down", "-v", "--remove-orphans")

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = s.ComposeDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.T.Logf("docker compose down warning: %v\nstderr: %s", err, stderr.String())
		return err
	}
	return nil
}

// Exec runs a command in a service container.
func (s *Stack) Exec(ctx context.Context, service string, command ...string) (string, error) {
	args := s.composeArgs("exec", "-T", service)
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = s.ComposeDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("exec failed: %w\nstderr: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Logs returns the logs from a service.
func (s *Stack) Logs(ctx context.Context, service string) (string, error) {
	args := s.composeArgs("logs", service)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = s.ComposeDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("logs failed: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// WaitForHealthy waits for a service to be healthy.
func (s *Stack) WaitForHealthy(ctx context.Context, service string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		args := s.composeArgs("ps", "--format", "{{.Health}}", service)
		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Dir = s.ComposeDir
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "healthy") {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			// continue polling
		}
	}

	return fmt.Errorf("timeout waiting for %s to be healthy", service)
}
