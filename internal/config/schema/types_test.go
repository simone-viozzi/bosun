package schema

import (
	"testing"
)

func TestIsValidScope(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"container is valid", "container", true},
		{"volume is valid", "volume", true},
		{"network is valid", "network", true},
		{"global is valid", "global", true},
		{"empty is invalid", "", false},
		{"unknown is invalid", "unknown", false},
		{"CONTAINER uppercase is invalid", "CONTAINER", false},
		{"Container mixed case is invalid", "Container", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidScope(tt.input)
			if got != tt.want {
				t.Errorf("IsValidScope(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidConfigType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"string is valid", "string", true},
		{"bool is valid", "bool", true},
		{"int is valid", "int", true},
		{"duration is valid", "duration", true},
		{"size is valid", "size", true},
		{"enum is valid", "enum", true},
		{"list is valid", "list", true},
		{"empty is invalid", "", false},
		{"unknown is invalid", "unknown", false},
		{"STRING uppercase is invalid", "STRING", false},
		{"integer is invalid", "integer", false},
		{"boolean is invalid", "boolean", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidConfigType(tt.input)
			if got != tt.want {
				t.Errorf("IsValidConfigType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSpec_Keys(t *testing.T) {
	spec := Spec{
		"bosun.z.last":   FieldSpec{Key: "bosun.z.last"},
		"bosun.a.first":  FieldSpec{Key: "bosun.a.first"},
		"bosun.m.middle": FieldSpec{Key: "bosun.m.middle"},
	}

	keys := spec.Keys()

	if len(keys) != 3 {
		t.Fatalf("Keys() returned %d keys, want 3", len(keys))
	}

	// Keys should be sorted
	expected := []string{"bosun.a.first", "bosun.m.middle", "bosun.z.last"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, key, expected[i])
		}
	}
}

func TestSpec_Get(t *testing.T) {
	spec := Spec{
		"bosun.test.key": FieldSpec{
			Key:   "bosun.test.key",
			Scope: ScopeContainer,
			Type:  TypeString,
		},
	}

	t.Run("existing key", func(t *testing.T) {
		fs, ok := spec.Get("bosun.test.key")
		if !ok {
			t.Fatal("Get() returned ok=false for existing key")
		}
		if fs.Key != "bosun.test.key" {
			t.Errorf("Get() Key = %q, want %q", fs.Key, "bosun.test.key")
		}
		if fs.Scope != ScopeContainer {
			t.Errorf("Get() Scope = %q, want %q", fs.Scope, ScopeContainer)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		_, ok := spec.Get("bosun.missing.key")
		if ok {
			t.Error("Get() returned ok=true for missing key")
		}
	})
}

func TestSpec_Scopes(t *testing.T) {
	spec := Spec{
		"bosun.container.a": FieldSpec{Key: "bosun.container.a", Scope: ScopeContainer},
		"bosun.container.b": FieldSpec{Key: "bosun.container.b", Scope: ScopeContainer},
		"bosun.volume.a":    FieldSpec{Key: "bosun.volume.a", Scope: ScopeVolume},
		"bosun.global.a":    FieldSpec{Key: "bosun.global.a", Scope: ScopeGlobal},
	}

	scopes := spec.Scopes()

	// Check container scope
	if len(scopes[ScopeContainer]) != 2 {
		t.Errorf("Scopes()[container] has %d fields, want 2", len(scopes[ScopeContainer]))
	}

	// Check volume scope
	if len(scopes[ScopeVolume]) != 1 {
		t.Errorf("Scopes()[volume] has %d fields, want 1", len(scopes[ScopeVolume]))
	}

	// Check global scope
	if len(scopes[ScopeGlobal]) != 1 {
		t.Errorf("Scopes()[global] has %d fields, want 1", len(scopes[ScopeGlobal]))
	}

	// Check network scope (should be empty/nil)
	if len(scopes[ScopeNetwork]) != 0 {
		t.Errorf("Scopes()[network] has %d fields, want 0", len(scopes[ScopeNetwork]))
	}

	// Check container fields are sorted
	containerFields := scopes[ScopeContainer]
	if containerFields[0].Key != "bosun.container.a" || containerFields[1].Key != "bosun.container.b" {
		t.Error("Scopes() container fields are not sorted by key")
	}
}
