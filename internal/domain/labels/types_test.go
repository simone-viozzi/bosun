package labels

import (
	"testing"
	"time"
)

func TestTypes(t *testing.T) {
	// Test DefaultLabelPrefix constant
	if DefaultLabelPrefix != "bosun." {
		t.Errorf("Expected DefaultLabelPrefix to be 'bosun.', got %s", DefaultLabelPrefix)
	}

	// Test Kind constants
	if KindContainer != "container" {
		t.Errorf("Expected KindContainer to be 'container', got %s", KindContainer)
	}
	if KindVolume != "volume" {
		t.Errorf("Expected KindVolume to be 'volume', got %s", KindVolume)
	}
	if KindNetwork != "network" {
		t.Errorf("Expected KindNetwork to be 'network', got %s", KindNetwork)
	}

	// Test LabeledEntity
	entity := LabeledEntity{
		Kind:   KindContainer,
		ID:     "test-id",
		Name:   "test-name",
		Labels: map[string]string{"key": "value"},
		Meta:   map[string]string{"project": "test"},
	}
	if entity.Kind != KindContainer {
		t.Errorf("Expected entity.Kind to be KindContainer")
	}

	// Test Snapshot
	snapshot := Snapshot{
		Entities: []LabeledEntity{entity},
		TakenAt:  time.Now(),
	}
	if len(snapshot.Entities) != 1 {
		t.Errorf("Expected snapshot to have 1 entity")
	}
}
