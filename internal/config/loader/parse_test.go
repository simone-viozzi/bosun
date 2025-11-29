package loader

import (
	"testing"
	"time"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"True", true, false},
		{"False", false, false},
		{"TRUE", true, false},
		{"FALSE", false, false},
		{"1", true, false},
		{"0", false, false},
		{"t", true, false},
		{"f", false, false},
		{"", false, true},
		{"yes", false, true},   // Note: strconv.ParseBool doesn't accept "yes"
		{"no", false, true},    // Note: strconv.ParseBool doesn't accept "no"
		{"maybe", false, true}, // Invalid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"-1", -1, false},
		{"100", 100, false},
		{"9999", 9999, false},
		{"-9999", -9999, false},
		{"", 0, true},
		{"abc", 0, true},
		{"1.5", 0, true},
		{"1e10", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"0", 0, false},
		{"1s", time.Second, false},
		{"30s", 30 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"1h", time.Hour, false},
		{"1h30m", time.Hour + 30*time.Minute, false},
		{"500ms", 500 * time.Millisecond, false},
		{"100us", 100 * time.Microsecond, false},
		{"2h45m30s", 2*time.Hour + 45*time.Minute + 30*time.Second, false},
		{"", 0, true},
		{"abc", 0, true},
		{"30", 0, true}, // Missing unit
		{"-1s", -time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"0", 0, false},
		{"1024", 1024, false},
		{"1K", 1024, false},
		{"1KB", 1024, false},
		{"1M", 1024 * 1024, false},
		{"1MB", 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"10GB", 10 * 1024 * 1024 * 1024, false},
		{"500MB", 500 * 1024 * 1024, false},
		{"", -1, true},
		{"abc", -1, true},
		{"-1", -1, true}, // Negative size
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSize(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseEnum(t *testing.T) {
	allowed := []string{"debug", "info", "warn", "error"}

	tests := []struct {
		input   string
		wantVal string
		wantOk  bool
	}{
		{"debug", "debug", true},
		{"info", "info", true},
		{"warn", "warn", true},
		{"error", "error", true},
		{"", "", false},
		{"verbose", "", false},
		{"DEBUG", "", false}, // Case-sensitive
		{"Info", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotVal, gotOk := parseEnum(tt.input, allowed)
			if gotOk != tt.wantOk {
				t.Errorf("parseEnum(%q) ok = %v, want %v", tt.input, gotOk, tt.wantOk)
			}
			if gotVal != tt.wantVal {
				t.Errorf("parseEnum(%q) val = %v, want %v", tt.input, gotVal, tt.wantVal)
			}
		})
	}
}

func TestParseList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  nil,
		},
		{
			name:  "csv single",
			input: "a",
			want:  []string{"a"},
		},
		{
			name:  "csv multiple",
			input: "a,b,c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "csv with spaces",
			input: "a, b, c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "csv with empty items",
			input: "a,,b",
			want:  []string{"a", "b"},
		},
		{
			name:  "json array",
			input: `["a", "b", "c"]`,
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "json array single",
			input: `["hello"]`,
			want:  []string{"hello"},
		},
		{
			name:  "json array empty",
			input: `[]`,
			want:  []string{},
		},
		{
			name:  "json array with special chars",
			input: `["hello,world", "foo", "bar"]`,
			want:  []string{"hello,world", "foo", "bar"},
		},
		{
			name:    "invalid json",
			input:   `["incomplete`,
			wantErr: true,
		},
		{
			name:  "json with leading spaces",
			input: `  ["a", "b"]`,
			want:  []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseList(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("parseList(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
