package ports

import (
	"context"
	"testing"
	"time"

	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

// mockLabelSource implements LabelSource for testing
type mockLabelSource struct{}

func (m *mockLabelSource) Snapshot(ctx context.Context, sel Selector) (dlabels.Snapshot, error) {
	return dlabels.Snapshot{
		Entities: []dlabels.LabeledEntity{
			{
				Kind:   dlabels.KindContainer,
				ID:     "test-container",
				Name:   "test",
				Labels: map[string]string{"bosun.test": "true"},
				Meta:   map[string]string{"project": "test"},
			},
		},
		TakenAt: time.Now(),
	}, nil
}

func TestInterfaces(t *testing.T) {
	// Test Selector struct
	selector := Selector{
		Prefixes:       []string{"bosun."},
		IncludeStopped: false,
		ProjectFilter:  []string{"test"},
	}
	if len(selector.Prefixes) != 1 {
		t.Errorf("Expected selector to have 1 prefix")
	}

	// Test LabelSource interface
	var source LabelSource = &mockLabelSource{}
	snapshot, err := source.Snapshot(context.Background(), selector)
	if err != nil {
		t.Errorf("Expected no error from Snapshot, got %v", err)
	}
	if len(snapshot.Entities) != 1 {
		t.Errorf("Expected snapshot to have 1 entity")
	}
}
