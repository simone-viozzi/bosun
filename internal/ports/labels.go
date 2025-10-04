package ports

import (
	"context"

	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

type Selector struct {
	Prefixes       []string
	IncludeStopped bool
	ProjectFilter  []string // optional filter by compose project
}

type LabelSource interface {
	Snapshot(ctx context.Context, sel Selector) (dlabels.Snapshot, error)
}
