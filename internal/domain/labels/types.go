package labels

import "time"

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
