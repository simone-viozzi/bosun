package ports

import (
	"context"

	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

type Selector struct {
	Prefixes       []string
	IncludeStopped bool
	ProjectFilter  []string // optional filter by com.docker.compose.project label
	StackFilter    []string // optional filter by bosun.stack label
}

type LabelSource interface {
	Snapshot(ctx context.Context, sel Selector) (dlabels.Snapshot, error)
}
