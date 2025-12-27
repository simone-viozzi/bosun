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
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
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

func TestSelectorFilters(t *testing.T) {
	t.Run("empty filters match all", func(t *testing.T) {
		selector := Selector{
			Prefixes:       []string{dlabels.DefaultLabelPrefix},
			IncludeStopped: false,
			ProjectFilter:  nil,
			StackFilter:    nil,
		}

		// Verify Prefixes and IncludeStopped are set correctly
		if len(selector.Prefixes) != 1 || selector.Prefixes[0] != dlabels.DefaultLabelPrefix {
			t.Errorf("Expected Prefixes=[%q], got %v", dlabels.DefaultLabelPrefix, selector.Prefixes)
		}
		if selector.IncludeStopped {
			t.Errorf("Expected IncludeStopped=false")
		}
		// Empty filters should not restrict matching
		if len(selector.ProjectFilter) != 0 {
			t.Errorf("Expected empty ProjectFilter, got %v", selector.ProjectFilter)
		}
		if len(selector.StackFilter) != 0 {
			t.Errorf("Expected empty StackFilter, got %v", selector.StackFilter)
		}
	})

	t.Run("both filters can be set", func(t *testing.T) {
		selector := Selector{
			Prefixes:       []string{dlabels.DefaultLabelPrefix},
			IncludeStopped: false,
			ProjectFilter:  []string{"myproject"},
			StackFilter:    []string{"production"},
		}

		// Verify Prefixes and IncludeStopped are set correctly
		if len(selector.Prefixes) != 1 || selector.Prefixes[0] != dlabels.DefaultLabelPrefix {
			t.Errorf("Expected Prefixes=[%q], got %v", dlabels.DefaultLabelPrefix, selector.Prefixes)
		}
		if selector.IncludeStopped {
			t.Errorf("Expected IncludeStopped=false")
		}
		if len(selector.ProjectFilter) != 1 || selector.ProjectFilter[0] != "myproject" {
			t.Errorf("Expected ProjectFilter=['myproject'], got %v", selector.ProjectFilter)
		}
		if len(selector.StackFilter) != 1 || selector.StackFilter[0] != "production" {
			t.Errorf("Expected StackFilter=['production'], got %v", selector.StackFilter)
		}
	})

	t.Run("multiple values in filters", func(t *testing.T) {
		selector := Selector{
			Prefixes:       []string{dlabels.DefaultLabelPrefix},
			IncludeStopped: false,
			ProjectFilter:  []string{"app-a", "app-b"},
			StackFilter:    []string{"staging", "production"},
		}

		// Verify Prefixes and IncludeStopped are set correctly
		if len(selector.Prefixes) != 1 || selector.Prefixes[0] != dlabels.DefaultLabelPrefix {
			t.Errorf("Expected Prefixes=[%q], got %v", dlabels.DefaultLabelPrefix, selector.Prefixes)
		}
		if selector.IncludeStopped {
			t.Errorf("Expected IncludeStopped=false")
		}
		if len(selector.ProjectFilter) != 2 {
			t.Errorf("Expected 2 project filters, got %d", len(selector.ProjectFilter))
		}
		if len(selector.StackFilter) != 2 {
			t.Errorf("Expected 2 stack filters, got %d", len(selector.StackFilter))
		}
	})
}
