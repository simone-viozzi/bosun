package state

import (
	"context"
	"sync"

	"github.com/simone-viozzi/bosun/internal/ports"
)

// InMemoryStateStore implements ports.JobStateStore with no durability.
// State is lost on daemon restart. This is the M4 default.
// Issue #177 replaces this with a durable adapter (BoltDB).
type InMemoryStateStore struct {
	states sync.Map // map[string]*ports.JobState
}

// NewInMemoryStateStore creates a new in-memory state store.
func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{}
}

func (s *InMemoryStateStore) SaveJobState(_ context.Context, state ports.JobState) error {
	s.states.Store(state.JobName, &state)
	return nil
}

func (s *InMemoryStateStore) LoadJobState(_ context.Context, jobName string) (ports.JobState, error) {
	val, ok := s.states.Load(jobName)
	if !ok {
		return ports.JobState{}, ports.ErrJobStateNotFound
	}
	return *val.(*ports.JobState), nil
}

func (s *InMemoryStateStore) ListJobStates(_ context.Context) ([]ports.JobState, error) {
	var result []ports.JobState
	s.states.Range(func(_, value any) bool {
		result = append(result, *value.(*ports.JobState))
		return true
	})
	return result, nil
}

func (s *InMemoryStateStore) DeleteJobState(_ context.Context, jobName string) error {
	s.states.Delete(jobName)
	return nil
}
