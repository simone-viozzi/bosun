package schema

import (
	"sort"
)

// Scope represents where a Docker label can be applied.
type Scope string

// Valid scope values.
const (
	ScopeContainer Scope = "container"
	ScopeVolume    Scope = "volume"
	ScopeNetwork   Scope = "network"
	ScopeGlobal    Scope = "global"
)

// validScopes is the set of valid scope values.
var validScopes = map[Scope]bool{
	ScopeContainer: true,
	ScopeVolume:    true,
	ScopeNetwork:   true,
	ScopeGlobal:    true,
}

// IsValidScope returns true if s is a valid scope value.
func IsValidScope(s string) bool {
	return validScopes[Scope(s)]
}

// ConfigType represents the type of a configuration value.
type ConfigType string

// Valid configuration types.
const (
	TypeString   ConfigType = "string"
	TypeBool     ConfigType = "bool"
	TypeInt      ConfigType = "int"
	TypeDuration ConfigType = "duration"
	TypeSize     ConfigType = "size"
	TypeEnum     ConfigType = "enum"
	TypeList     ConfigType = "list"
)

// validConfigTypes is the set of valid configuration type values.
var validConfigTypes = map[ConfigType]bool{
	TypeString:   true,
	TypeBool:     true,
	TypeInt:      true,
	TypeDuration: true,
	TypeSize:     true,
	TypeEnum:     true,
	TypeList:     true,
}

// IsValidConfigType returns true if t is a valid configuration type value.
func IsValidConfigType(t string) bool {
	return validConfigTypes[ConfigType(t)]
}

// FieldSpec contains metadata for a single configuration field.
type FieldSpec struct {
	// Key is the Docker label key (e.g., "bosun.container.stopGracePeriod").
	Key string

	// Scope indicates where this label can be applied.
	Scope Scope

	// Type specifies how to parse the value.
	Type ConfigType

	// Default is the default value as a string (parsed according to Type).
	Default string

	// Enum contains allowed values if Type is TypeEnum.
	Enum []string

	// Required indicates whether this field must be present.
	Required bool

	// Doc is the human-readable description.
	Doc string

	// Deprecated indicates whether this field is deprecated.
	Deprecated bool

	// FieldName is the Go struct field name (for reflection).
	FieldName string
}

// Spec maps label keys to their field specifications.
type Spec map[string]FieldSpec

// Keys returns all keys in the spec in deterministic (sorted) order.
func (s Spec) Keys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Get returns the FieldSpec for the given key and a boolean indicating
// whether the key exists.
func (s Spec) Get(key string) (FieldSpec, bool) {
	fs, ok := s[key]
	return fs, ok
}

// Scopes returns a map grouping fields by their scope.
func (s Spec) Scopes() map[Scope][]FieldSpec {
	result := make(map[Scope][]FieldSpec)
	for _, fs := range s {
		result[fs.Scope] = append(result[fs.Scope], fs)
	}
	// Sort each scope's fields by key for deterministic output
	for scope := range result {
		sort.Slice(result[scope], func(i, j int) bool {
			return result[scope][i].Key < result[scope][j].Key
		})
	}
	return result
}
