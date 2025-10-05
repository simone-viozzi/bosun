package labels

import "time"

// DefaultLabelPrefix is the standard prefix for Bosun-managed labels.
const DefaultLabelPrefix = "bosun."

// TODO this cannot be here, we need a better way of handling this
// LabelInstance is the label key for instance identification.
const LabelInstance = DefaultLabelPrefix + "instance"

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
