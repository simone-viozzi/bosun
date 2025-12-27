package labels

import "time"

// DefaultLabelPrefix is the standard prefix for Bosun-managed labels.
const DefaultLabelPrefix = "bosun."

// LabelInstance is the label key for instance identification.
// TODO: Consider a better way of handling shared label constants.
const LabelInstance = DefaultLabelPrefix + "instance"

// LabelStack is the label key for stack grouping (used for filtering).
const LabelStack = DefaultLabelPrefix + "stack"

type Kind string

const (
	KindContainer Kind = "container"
	KindVolume    Kind = "volume"
	KindNetwork   Kind = "network"
)

type LabeledEntity struct {
	Kind   Kind
	ID     string
	Name   string
	Labels map[string]string
	Meta   map[string]string // e.g., "compose.project", "compose.service", "image", "networks"
}

type Snapshot struct {
	Entities []LabeledEntity
	TakenAt  time.Time
}
