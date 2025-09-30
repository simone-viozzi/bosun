//go:build integration
// +build integration

package testutil

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tccompose "github.com/testcontainers/testcontainers-go/modules/compose"
)

//go:embed compose/*.yaml
var ComposeFS embed.FS

type Stack struct {
	Project string
	Files   []string
	Cmp     *tccompose.LocalDockerCompose
	T       *testing.T
}

func StartCompose(t *testing.T, ctx context.Context, files ...string) *Stack {
	t.Helper()

	project := sanitize(fmt.Sprintf("bosun_%s_%d", t.Name(), time.Now().UnixNano()))
	tmpDir := t.TempDir()

	var paths []string
	for _, f := range files {
		b, err := ComposeFS.ReadFile(filepath.Join("compose", f))
		if err != nil {
			t.Fatalf("read compose %s: %v", f, err)
		}
		p := filepath.Join(tmpDir, f)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
		paths = append(paths, p)
	}

	log.Printf("Starting compose stack with project: %s, files: %v", project, files)
	cmp := tccompose.NewLocalDockerCompose(paths, project)

	if err := cmp.Invoke().Error; err != nil {
		t.Fatalf("compose up: %v", err)
	}
	log.Printf("Compose stack %s started successfully", project)

	st := &Stack{Project: project, Files: paths, Cmp: cmp, T: t}
	t.Cleanup(func() {
		log.Printf("Stopping compose stack: %s", project)
		_ = st.Cmp.Down()
		log.Printf("Compose stack %s stopped", project)
	})
	return st
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '-'
	}, s)
	return s
}
