package events_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/simone-viozzi/bosun/internal/adapters/events"
)

func newTestEmitter() (*events.LogEmitter, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return events.NewLogEmitter(logger), &buf
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q, got:\n%s", needle, haystack)
	}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }

var errTest = &testError{}

func TestLogEmitter_EmitJobScheduled(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobScheduled(context.Background(), "daily-backup", "0 3 * * *", time.Now())
	out := buf.String()
	assertContains(t, out, "job.scheduled")
	assertContains(t, out, "daily-backup")
	assertContains(t, out, "0 3 * * *")
}

func TestLogEmitter_EmitJobStarted(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobStarted(context.Background(), "daily-backup", "run-123")
	out := buf.String()
	assertContains(t, out, "job.started")
	assertContains(t, out, "run-123")
}

func TestLogEmitter_EmitJobCompleted(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobCompleted(context.Background(), "daily-backup", "run-123", 5*time.Second)
	out := buf.String()
	assertContains(t, out, "job.completed")
	assertContains(t, out, "daily-backup")
}

func TestLogEmitter_EmitJobFailed(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobFailed(context.Background(), "daily-backup", "run-123", errTest, 10*time.Second)
	out := buf.String()
	assertContains(t, out, "job.failed")
	assertContains(t, out, "test error")
}

func TestLogEmitter_EmitJobSkipped(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobSkipped(context.Background(), "daily-backup", "overlap policy: skip")
	out := buf.String()
	assertContains(t, out, "job.skipped")
	assertContains(t, out, "overlap policy: skip")
}

func TestLogEmitter_EmitJobAdded(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobAdded(context.Background(), "new-job")
	out := buf.String()
	assertContains(t, out, "job.added")
	assertContains(t, out, "new-job")
}

func TestLogEmitter_EmitJobRemoved(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobRemoved(context.Background(), "old-job")
	out := buf.String()
	assertContains(t, out, "job.removed")
	assertContains(t, out, "old-job")
}

func TestLogEmitter_EmitJobChanged(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobChanged(context.Background(), "daily-backup", "0 3 * * *", "0 4 * * *")
	out := buf.String()
	assertContains(t, out, "job.changed")
	assertContains(t, out, "0 3 * * *")
	assertContains(t, out, "0 4 * * *")
}

func TestLogEmitter_EmitJobCircuitBroken(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobCircuitBroken(context.Background(), "flaky-job", 3)
	out := buf.String()
	assertContains(t, out, "job.circuit_broken")
	assertContains(t, out, "flaky-job")
}

func TestLogEmitter_EmitJobDuplicateName(t *testing.T) {
	emitter, buf := newTestEmitter()
	emitter.EmitJobDuplicateName(context.Background(), "dupe-job", "container-abc")
	out := buf.String()
	assertContains(t, out, "job.duplicate_name")
	assertContains(t, out, "container-abc")
}
