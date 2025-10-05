package dockerlabels

import (
	"reflect"
	"testing"

	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
)

func TestFilterByPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		prefixes []string
		expected map[string]string
	}{
		{
			name:     "no prefixes returns empty map",
			input:    map[string]string{"bosun.key": "value", "other.key": "value2"},
			prefixes: []string{},
			expected: map[string]string{},
		},
		{
			name:     "single prefix filters correctly",
			input:    map[string]string{"bosun.key1": "value1", "bosun.key2": "value2", "other.key": "value3"},
			prefixes: []string{dlabels.DefaultLabelPrefix},
			expected: map[string]string{"bosun.key1": "value1", "bosun.key2": "value2"},
		},
		{
			name:     "multiple prefixes filter correctly",
			input:    map[string]string{"bosun.key": "value1", "docker.key": "value2", "other.key": "value3"},
			prefixes: []string{dlabels.DefaultLabelPrefix, "docker."},
			expected: map[string]string{"bosun.key": "value1", "docker.key": "value2"},
		},
		{
			name:     "empty values are dropped",
			input:    map[string]string{"bosun.key1": "value1", "bosun.key2": "", "bosun.key3": "value3"},
			prefixes: []string{dlabels.DefaultLabelPrefix},
			expected: map[string]string{"bosun.key1": "value1", "bosun.key3": "value3"},
		},
		{
			name:     "whitespace values are dropped",
			input:    map[string]string{"bosun.key1": "value1", "bosun.key2": "   ", "bosun.key3": "\t\n"},
			prefixes: []string{dlabels.DefaultLabelPrefix},
			expected: map[string]string{"bosun.key1": "value1"},
		},
		{
			name:     "no matching prefixes returns empty",
			input:    map[string]string{"bosun.key": "value", "other.key": "value2"},
			prefixes: []string{"nomatch."},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of input to verify it's not mutated
			originalInput := make(map[string]string)
			for k, v := range tt.input {
				originalInput[k] = v
			}

			result := FilterByPrefixes(tt.input, tt.prefixes)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterByPrefixes() = %v, expected %v", result, tt.expected)
			}

			// Verify input map was not mutated
			if !reflect.DeepEqual(tt.input, originalInput) {
				t.Errorf("Input map was mutated: original %v, current %v", originalInput, tt.input)
			}
		})
	}
}
