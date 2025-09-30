//go:build integration
// +build integration

package testutil

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/compose"
)

//go:embed compose/*.yaml
var ComposeFS embed.FS

type Stack struct {
	Project string
	Files   []string
	Cmp     compose.ComposeStack // NOTE: interface, not LocalDockerCompose
	T       *testing.T
}

func StartCompose(t *testing.T, ctx context.Context, files ...string) *Stack {
	t.Helper()

	project := sanitize(fmt.Sprintf("bosun-%s-%d", t.Name(), time.Now().UnixNano()))

	// Feed compose from embedded files via readers (no temp files needed)
	var readers []io.Reader
	for _, f := range files {
		b, err := ComposeFS.ReadFile(filepath.Join("compose", f))
		if err != nil {
			t.Fatalf("read compose %s: %v", f, err)
		}
		readers = append(readers, bytes.NewReader(b))
	}

	stack, err := compose.NewDockerComposeWith(
		compose.StackIdentifier(project),
		compose.WithStackReaders(readers...),
	)
	if err != nil {
		t.Fatalf("compose create: %v", err)
	}

	// Optional: declare readiness for critical services if they lack HEALTHCHECK
	// stack = stack.WaitForService("db", wait.ForListeningPort("5432/tcp"))

	if err := stack.Up(ctx, compose.Wait(true)); err != nil { // waits for running/healthy
		t.Fatalf("compose up: %v", err)
	}

	st := &Stack{Project: project, Files: files, Cmp: stack, T: t}
	t.Cleanup(func() {
		_ = st.Cmp.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			// compose.RemoveImagesLocal, // add if you really want images removed
		)
	})
	return st
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, s)
}
