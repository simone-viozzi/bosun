package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- T063: Text table output formatting (columns, alignment) ---

func TestRenderTextTable_MultipleJobs(t *testing.T) {
	t.Parallel()

	entries := []jobListEntry{
		{Name: "backup-db", Schedule: "0 3 * * *", OverlapPolicy: "queue", Enabled: true, Stacks: "postgres"},
		{Name: "cleanup-logs", Schedule: "0 0 * * 0", OverlapPolicy: "skip", Enabled: true, Stacks: "app, monitoring"},
		{Name: "sync-data", Schedule: "*/15 * * * *", OverlapPolicy: "queue", Enabled: false, Stacks: "-"},
	}

	var buf bytes.Buffer
	renderTextTable(&buf, entries)
	output := buf.String()

	// Check header row.
	if !strings.Contains(output, "NAME") {
		t.Error("missing NAME header")
	}
	if !strings.Contains(output, "SCHEDULE") {
		t.Error("missing SCHEDULE header")
	}
	if !strings.Contains(output, "OVERLAP") {
		t.Error("missing OVERLAP header")
	}
	if !strings.Contains(output, "ENABLED") {
		t.Error("missing ENABLED header")
	}
	if !strings.Contains(output, "STACKS") {
		t.Error("missing STACKS header")
	}

	// Check data rows.
	if !strings.Contains(output, "backup-db") {
		t.Error("missing backup-db in output")
	}
	if !strings.Contains(output, "cleanup-logs") {
		t.Error("missing cleanup-logs in output")
	}
	if !strings.Contains(output, "sync-data") {
		t.Error("missing sync-data in output")
	}

	// Verify alignment: all lines should have the same number of columns.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 { // 1 header + 3 data rows
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestRenderTextTable_EmptyList(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	renderTextTable(&buf, nil)

	if !strings.Contains(buf.String(), "No jobs discovered") {
		t.Errorf("expected 'No jobs discovered' message, got %q", buf.String())
	}
}

func TestRenderTextTable_ColumnAlignment(t *testing.T) {
	t.Parallel()

	entries := []jobListEntry{
		{Name: "a", Schedule: "* * * * *", OverlapPolicy: "queue", Enabled: true, Stacks: "-"},
		{Name: "long-job-name", Schedule: "0 0 1 1 *", OverlapPolicy: "skip", Enabled: false, Stacks: "my-stack"},
	}

	var buf bytes.Buffer
	renderTextTable(&buf, entries)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Verify NAME column is left-aligned and padded consistently.
	// The header "NAME" and data "a" and "long-job-name" should all
	// start at the same position (column 0).
	for _, line := range lines {
		if len(line) == 0 || line[0] == ' ' {
			t.Errorf("line should start at column 0: %q", line)
		}
	}
}

// --- T064: JSON and YAML output formats ---

func TestRenderJSON_ValidOutput(t *testing.T) {
	t.Parallel()

	entries := []jobListEntry{
		{Name: "backup-db", Schedule: "0 3 * * *", OverlapPolicy: "queue", Enabled: true, Stacks: "postgres"},
	}

	var buf bytes.Buffer
	if err := renderJSON(&buf, entries); err != nil {
		t.Fatalf("renderJSON: %v", err)
	}

	// Parse JSON to verify structure.
	var parsed []jobListEntry
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0].Name != "backup-db" {
		t.Errorf("name = %q, want %q", parsed[0].Name, "backup-db")
	}
	if parsed[0].Schedule != "0 3 * * *" {
		t.Errorf("schedule = %q, want %q", parsed[0].Schedule, "0 3 * * *")
	}
}

func TestRenderJSON_EmptyList(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := renderJSON(&buf, []jobListEntry{}); err != nil {
		t.Fatalf("renderJSON: %v", err)
	}

	var parsed []jobListEntry
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty array, got %d entries", len(parsed))
	}
}

func TestRenderYAML_ValidOutput(t *testing.T) {
	t.Parallel()

	entries := []jobListEntry{
		{Name: "backup-db", Schedule: "0 3 * * *", OverlapPolicy: "queue", Enabled: true, Stacks: "postgres"},
	}

	var buf bytes.Buffer
	if err := renderYAML(&buf, entries); err != nil {
		t.Fatalf("renderYAML: %v", err)
	}

	// Parse YAML to verify structure.
	var parsed []jobListEntry
	if err := yaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid YAML: %v\noutput: %s", err, buf.String())
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0].Name != "backup-db" {
		t.Errorf("name = %q, want %q", parsed[0].Name, "backup-db")
	}
}

func TestRenderYAML_MultipleEntries(t *testing.T) {
	t.Parallel()

	entries := []jobListEntry{
		{Name: "job-a", Schedule: "0 0 * * *", OverlapPolicy: "queue", Enabled: true, Stacks: "-"},
		{Name: "job-b", Schedule: "0 6 * * *", OverlapPolicy: "skip", Enabled: false, Stacks: "stack-1"},
	}

	var buf bytes.Buffer
	if err := renderYAML(&buf, entries); err != nil {
		t.Fatalf("renderYAML: %v", err)
	}

	var parsed []jobListEntry
	if err := yaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid YAML: %v", err)
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 entries, got %d", len(parsed))
	}
}

// --- Command integration ---

func TestNewJobListCmd_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewJobListCmd()

	fFlag := cmd.Flags().Lookup("format")
	if fFlag == nil {
		t.Fatal("missing --format flag")
	} else {
		if fFlag.DefValue != "text" {
			t.Errorf("--format default = %q, want %q", fFlag.DefValue, "text")
		}
		if fFlag.Shorthand != "f" {
			t.Errorf("--format shorthand = %q, want %q", fFlag.Shorthand, "f")
		}
	}
}
