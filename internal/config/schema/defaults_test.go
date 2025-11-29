package schema

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
		{"yes", true, false},
		{"no", false, false},
		{"Yes", true, false},
		{"No", false, false},
		{"", false, true},
		{"invalid", false, true},
		{"truee", false, true},
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
		want    int64
		wantErr bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"-1", -1, false},
		{"42", 42, false},
		{"9223372036854775807", 9223372036854775807, false},   // max int64
		{"-9223372036854775808", -9223372036854775808, false}, // min int64
		{"", 0, true},
		{"abc", 0, true},
		{"1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
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
		{"1h30m", 90 * time.Minute, false},
		{"500ms", 500 * time.Millisecond, false},
		{"2h45m30s", 2*time.Hour + 45*time.Minute + 30*time.Second, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"30", 0, true}, // missing unit
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
		{"1", 1, false},
		{"1B", 1, false},
		{"1KB", 1024, false},
		{"1MB", 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"1TB", 1024 * 1024 * 1024 * 1024, false},
		{"500MB", 500 * 1024 * 1024, false},
		{"1.5GB", int64(1.5 * 1024 * 1024 * 1024), false},
		// Note: units.RAMInBytes returns -1 for errors, not 0
		{"", -1, true},
		{"invalid", -1, true},
		{"-1GB", -1, true}, // negative not allowed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseSize(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseList(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", []string{}},
		{"a", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"  a  ,  b  ,  c  ", []string{"a", "b", "c"}},
		{"one,two", []string{"one", "two"}},
		{"a,,b", []string{"a", "b"}}, // empty parts are skipped
		{",a,", []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseList(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseList(%q) returned %d items, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseList(%q)[%d] = %q, want %q", tt.input, i, v, tt.want[i])
				}
			}
		})
	}
}

// Test structs for DefaultOf tests.
type testDefaultsConfig struct {
	Name    string        `bosun:"key=bosun.name,scope=global,type=string,default=test-name"`
	Timeout time.Duration `bosun:"key=bosun.timeout,scope=container,type=duration,default=30s"`
	Enabled bool          `bosun:"key=bosun.enabled,scope=container,type=bool,default=true"`
	Count   int           `bosun:"key=bosun.count,scope=global,type=int,default=42"`
}

type defaultsWithSizeAndList struct {
	MaxSize int64    `bosun:"key=bosun.maxSize,scope=volume,type=size,default=1GB"`
	Tags    []string `bosun:"key=bosun.tags,scope=container,type=list,default='a,b,c'"`
}

type defaultsWithEnum struct {
	Level string `bosun:"key=bosun.level,scope=container,type=enum,enum=debug|info|warn|error,default=info"`
}

type defaultsWithNoDefaults struct {
	Name    string `bosun:"key=bosun.name,scope=global,type=string"`
	Enabled bool   `bosun:"key=bosun.enabled,scope=container,type=bool"`
}

type defaultsWithMixed struct {
	WithDefault    string `bosun:"key=bosun.withDefault,scope=global,type=string,default=has-default"`
	WithoutDefault string `bosun:"key=bosun.withoutDefault,scope=global,type=string"`
}

type embeddedDefaults struct {
	testDefaultsConfig
	Extra string `bosun:"key=bosun.extra,scope=global,type=string,default=extra-value"`
}

type invalidDefaultDuration struct {
	Timeout time.Duration `bosun:"key=bosun.timeout,scope=container,type=duration,default=invalid"`
}

type invalidDefaultInt struct {
	Count int `bosun:"key=bosun.count,scope=global,type=int,default=not-a-number"`
}

type invalidDefaultSize struct {
	MaxSize int64 `bosun:"key=bosun.maxSize,scope=volume,type=size,default=invalid"`
}

func TestDefaultOf(t *testing.T) {
	t.Run("basic types with defaults", func(t *testing.T) {
		cfg, err := DefaultOf[testDefaultsConfig]()
		if err != nil {
			t.Fatalf("DefaultOf() error: %v", err)
		}

		if cfg.Name != "test-name" {
			t.Errorf("Name = %q, want %q", cfg.Name, "test-name")
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
		}
		if cfg.Enabled != true {
			t.Errorf("Enabled = %v, want %v", cfg.Enabled, true)
		}
		if cfg.Count != 42 {
			t.Errorf("Count = %v, want %v", cfg.Count, 42)
		}
	})

	t.Run("size and list types", func(t *testing.T) {
		cfg, err := DefaultOf[defaultsWithSizeAndList]()
		if err != nil {
			t.Fatalf("DefaultOf() error: %v", err)
		}

		expectedSize := int64(1024 * 1024 * 1024) // 1GB
		if cfg.MaxSize != expectedSize {
			t.Errorf("MaxSize = %v, want %v", cfg.MaxSize, expectedSize)
		}

		expectedTags := []string{"a", "b", "c"}
		if len(cfg.Tags) != len(expectedTags) {
			t.Errorf("Tags length = %d, want %d", len(cfg.Tags), len(expectedTags))
		} else {
			for i, v := range cfg.Tags {
				if v != expectedTags[i] {
					t.Errorf("Tags[%d] = %q, want %q", i, v, expectedTags[i])
				}
			}
		}
	})

	t.Run("enum type", func(t *testing.T) {
		cfg, err := DefaultOf[defaultsWithEnum]()
		if err != nil {
			t.Fatalf("DefaultOf() error: %v", err)
		}

		if cfg.Level != "info" {
			t.Errorf("Level = %q, want %q", cfg.Level, "info")
		}
	})

	t.Run("no defaults leaves zero values", func(t *testing.T) {
		cfg, err := DefaultOf[defaultsWithNoDefaults]()
		if err != nil {
			t.Fatalf("DefaultOf() error: %v", err)
		}

		if cfg.Name != "" {
			t.Errorf("Name = %q, want empty", cfg.Name)
		}
		if cfg.Enabled != false {
			t.Errorf("Enabled = %v, want false", cfg.Enabled)
		}
	})

	t.Run("mixed defaults and no-defaults", func(t *testing.T) {
		cfg, err := DefaultOf[defaultsWithMixed]()
		if err != nil {
			t.Fatalf("DefaultOf() error: %v", err)
		}

		if cfg.WithDefault != "has-default" {
			t.Errorf("WithDefault = %q, want %q", cfg.WithDefault, "has-default")
		}
		if cfg.WithoutDefault != "" {
			t.Errorf("WithoutDefault = %q, want empty", cfg.WithoutDefault)
		}
	})

	t.Run("embedded struct defaults", func(t *testing.T) {
		cfg, err := DefaultOf[embeddedDefaults]()
		if err != nil {
			t.Fatalf("DefaultOf() error: %v", err)
		}

		// Check embedded fields
		if cfg.Name != "test-name" {
			t.Errorf("Name = %q, want %q", cfg.Name, "test-name")
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
		}

		// Check direct field
		if cfg.Extra != "extra-value" {
			t.Errorf("Extra = %q, want %q", cfg.Extra, "extra-value")
		}
	})

	t.Run("error on invalid duration default", func(t *testing.T) {
		_, err := DefaultOf[invalidDefaultDuration]()
		if err == nil {
			t.Error("DefaultOf() expected error for invalid duration")
		}
	})

	t.Run("error on invalid int default", func(t *testing.T) {
		_, err := DefaultOf[invalidDefaultInt]()
		if err == nil {
			t.Error("DefaultOf() expected error for invalid int")
		}
	})

	t.Run("error on invalid size default", func(t *testing.T) {
		_, err := DefaultOf[invalidDefaultSize]()
		if err == nil {
			t.Error("DefaultOf() expected error for invalid size")
		}
	})

	t.Run("error on non-struct type", func(t *testing.T) {
		_, err := DefaultOf[string]()
		if err == nil {
			t.Error("DefaultOf() expected error for non-struct type")
		}
	})
}
