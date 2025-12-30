package compose

import (
	"errors"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
	"github.com/simone-viozzi/bosun/internal/ports"
)

func TestParseDependsOn_Empty(t *testing.T) {
	result := parseDependsOn("")
	if result != nil {
		t.Errorf("parseDependsOn(\"\") = %v, want nil", result)
	}
}

func TestParseDependsOn_Single(t *testing.T) {
	result := parseDependsOn("db")
	if len(result) != 1 || result[0] != "db" {
		t.Errorf("parseDependsOn(\"db\") = %v, want [\"db\"]", result)
	}
}

func TestParseDependsOn_Multiple(t *testing.T) {
	result := parseDependsOn("db,redis,cache")
	expected := []string{"db", "redis", "cache"}
	if len(result) != len(expected) {
		t.Fatalf("parseDependsOn length = %d, want %d", len(result), len(expected))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("parseDependsOn[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestParseDependsOn_WithSpaces(t *testing.T) {
	result := parseDependsOn("  db  , redis , cache  ")
	expected := []string{"db", "redis", "cache"}
	if len(result) != len(expected) {
		t.Fatalf("parseDependsOn length = %d, want %d", len(result), len(expected))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("parseDependsOn[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestTopologicalSort_Empty(t *testing.T) {
	result, err := topologicalSort(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("topologicalSort(nil) = %v, want nil", result)
	}
}

func TestTopologicalSort_SingleContainer(t *testing.T) {
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "web", State: "running"},
	}

	result, err := topologicalSort(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("result length = %d, want 1", len(result))
	}
	if result[0].ServiceName != "web" {
		t.Errorf("result[0].ServiceName = %q, want %q", result[0].ServiceName, "web")
	}
}

func TestTopologicalSort_LinearDependency(t *testing.T) {
	// web -> api -> db (web depends on api, api depends on db)
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "web", State: "running", DependsOn: []string{"api"}},
		{ID: "c2", ServiceName: "api", State: "running", DependsOn: []string{"db"}},
		{ID: "c3", ServiceName: "db", State: "running"},
	}

	result, err := topologicalSort(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("result length = %d, want 3", len(result))
	}

	// Expected order: db, api, web (dependencies first)
	if result[0].ServiceName != "db" {
		t.Errorf("result[0].ServiceName = %q, want %q", result[0].ServiceName, "db")
	}
	if result[1].ServiceName != "api" {
		t.Errorf("result[1].ServiceName = %q, want %q", result[1].ServiceName, "api")
	}
	if result[2].ServiceName != "web" {
		t.Errorf("result[2].ServiceName = %q, want %q", result[2].ServiceName, "web")
	}
}

func TestTopologicalSort_DiamondDependency(t *testing.T) {
	// Diamond: web depends on api and worker, both depend on db
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "web", State: "running", DependsOn: []string{"api", "worker"}},
		{ID: "c2", ServiceName: "api", State: "running", DependsOn: []string{"db"}},
		{ID: "c3", ServiceName: "worker", State: "running", DependsOn: []string{"db"}},
		{ID: "c4", ServiceName: "db", State: "running"},
	}

	result, err := topologicalSort(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("result length = %d, want 4", len(result))
	}

	// db must be first, web must be last
	if result[0].ServiceName != "db" {
		t.Errorf("result[0].ServiceName = %q, want %q", result[0].ServiceName, "db")
	}
	if result[3].ServiceName != "web" {
		t.Errorf("result[3].ServiceName = %q, want %q", result[3].ServiceName, "web")
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	// Create a cycle: a -> b -> c -> a
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "a", State: "running", DependsOn: []string{"c"}},
		{ID: "c2", ServiceName: "b", State: "running", DependsOn: []string{"a"}},
		{ID: "c3", ServiceName: "c", State: "running", DependsOn: []string{"b"}},
	}

	_, err := topologicalSort(containers)
	if err == nil {
		t.Error("expected error for cycle, got nil")
	}
}

func TestTopologicalSort_ExternalDependency(t *testing.T) {
	// web depends on external (not in this stack)
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "web", State: "running", DependsOn: []string{"external-db"}},
	}

	result, err := topologicalSort(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("result length = %d, want 1", len(result))
	}
}

func TestReverse_Empty(t *testing.T) {
	result := reverse(nil)
	if len(result) != 0 {
		t.Errorf("reverse(nil) length = %d, want 0", len(result))
	}
}

func TestReverse_Single(t *testing.T) {
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "web"},
	}

	result := reverse(containers)
	if len(result) != 1 {
		t.Fatalf("result length = %d, want 1", len(result))
	}
	if result[0].ServiceName != "web" {
		t.Errorf("result[0].ServiceName = %q, want %q", result[0].ServiceName, "web")
	}
}

func TestReverse_Multiple(t *testing.T) {
	containers := []ports.StackContainer{
		{ID: "c1", ServiceName: "db"},
		{ID: "c2", ServiceName: "api"},
		{ID: "c3", ServiceName: "web"},
	}

	result := reverse(containers)
	if len(result) != 3 {
		t.Fatalf("result length = %d, want 3", len(result))
	}

	// Expected reverse order: web, api, db
	if result[0].ServiceName != "web" {
		t.Errorf("result[0].ServiceName = %q, want %q", result[0].ServiceName, "web")
	}
	if result[1].ServiceName != "api" {
		t.Errorf("result[1].ServiceName = %q, want %q", result[1].ServiceName, "api")
	}
	if result[2].ServiceName != "db" {
		t.Errorf("result[2].ServiceName = %q, want %q", result[2].ServiceName, "db")
	}
}

func TestDefaultStopOptions(t *testing.T) {
	opts := ports.DefaultStopOptions()
	if opts.Timeout != 30*time.Second {
		t.Errorf("DefaultStopOptions().Timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}
}

func TestDefaultStartOptions(t *testing.T) {
	opts := ports.DefaultStartOptions()
	if opts.Timeout != 30*time.Second {
		t.Errorf("DefaultStartOptions().Timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}
}

func TestStopError_Error(t *testing.T) {
	cause := errors.New("connection refused")
	err := &jobs.StopError{
		StackName:     "mystack",
		ContainerName: "web",
		ContainerID:   "abc123",
		Cause:         cause,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("StopError.Error() returned empty string")
	}

	// Should contain key information
	if !containsString(msg, "mystack") {
		t.Errorf("StopError.Error() should contain stack name, got: %q", msg)
	}
	if !containsString(msg, "web") {
		t.Errorf("StopError.Error() should contain container name, got: %q", msg)
	}
}

func TestStartError_Error(t *testing.T) {
	cause := errors.New("image not found")
	err := &jobs.StartError{
		StackName:     "mystack",
		ContainerName: "web",
		ContainerID:   "abc123",
		Cause:         cause,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("StartError.Error() returned empty string")
	}

	// Should contain key information
	if !containsString(msg, "mystack") {
		t.Errorf("StartError.Error() should contain stack name, got: %q", msg)
	}
	if !containsString(msg, "web") {
		t.Errorf("StartError.Error() should contain container name, got: %q", msg)
	}
}

func TestStopError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &jobs.StopError{
		StackName: "mystack",
		Cause:     cause,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("StopError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestStartError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &jobs.StartError{
		StackName: "mystack",
		Cause:     cause,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("StartError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestStackContainer_Fields(t *testing.T) {
	sc := ports.StackContainer{
		ID:          "container123",
		Name:        "mystack-web-1",
		ServiceName: "web",
		State:       "running",
		DependsOn:   []string{"db"},
		Labels: map[string]string{
			"com.docker.compose.project": "mystack",
			"com.docker.compose.service": "web",
		},
	}

	if sc.ID != "container123" {
		t.Errorf("ID = %q, want %q", sc.ID, "container123")
	}
	if sc.Name != "mystack-web-1" {
		t.Errorf("Name = %q, want %q", sc.Name, "mystack-web-1")
	}
	if sc.ServiceName != "web" {
		t.Errorf("ServiceName = %q, want %q", sc.ServiceName, "web")
	}
	if sc.State != "running" {
		t.Errorf("State = %q, want %q", sc.State, "running")
	}
	if len(sc.DependsOn) != 1 || sc.DependsOn[0] != "db" {
		t.Errorf("DependsOn = %v, want [\"db\"]", sc.DependsOn)
	}
	if sc.Labels["com.docker.compose.project"] != "mystack" {
		t.Errorf("Labels[project] = %q, want %q", sc.Labels["com.docker.compose.project"], "mystack")
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
