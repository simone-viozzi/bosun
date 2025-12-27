package schema

import (
	"reflect"
	"testing"
	"time"
)

func TestParseTagValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "empty string",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "single key-value",
			input: "key=bosun.test",
			want:  map[string]string{"key": "bosun.test"},
		},
		{
			name:  "multiple key-values",
			input: "key=bosun.test,scope=container,type=string",
			want: map[string]string{
				"key":   "bosun.test",
				"scope": "container",
				"type":  "string",
			},
		},
		{
			name:  "quoted doc with comma",
			input: "key=bosun.test,doc='Hello, world'",
			want: map[string]string{
				"key": "bosun.test",
				"doc": "Hello, world",
			},
		},
		{
			name:  "quoted doc with equals",
			input: "key=bosun.test,doc='x=y'",
			want: map[string]string{
				"key": "bosun.test",
				"doc": "x=y",
			},
		},
		{
			name:  "full tag",
			input: "key=bosun.container.stop,scope=container,type=duration,default=30s,doc='Grace period'",
			want: map[string]string{
				"key":     "bosun.container.stop",
				"scope":   "container",
				"type":    "duration",
				"default": "30s",
				"doc":     "Grace period",
			},
		},
		{
			name:  "enum values",
			input: "key=bosun.level,type=enum,enum=debug|info|warn|error",
			want: map[string]string{
				"key":  "bosun.level",
				"type": "enum",
				"enum": "debug|info|warn|error",
			},
		},
		{
			name:  "boolean flags",
			input: "key=bosun.req,required=true,deprecated=true",
			want: map[string]string{
				"key":        "bosun.req",
				"required":   "true",
				"deprecated": "true",
			},
		},
		{
			name:  "whitespace handling",
			input: "key = bosun.test , scope = container",
			want: map[string]string{
				"key":   "bosun.test",
				"scope": "container",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTagValue(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseTagValue() got %d pairs, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseTagValue()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestParseFieldSpec(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		goType    reflect.Type
		tagParts  map[string]string
		want      FieldSpec
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid string field",
			fieldName: "Name",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "bosun.name",
				"scope": "global",
				"type":  "string",
			},
			want: FieldSpec{
				Key:       "bosun.name",
				Scope:     ScopeGlobal,
				Type:      TypeString,
				FieldName: "Name",
			},
		},
		{
			name:      "valid duration field with default",
			fieldName: "Timeout",
			goType:    reflect.TypeOf(time.Duration(0)),
			tagParts: map[string]string{
				"key":     "bosun.timeout",
				"scope":   "container",
				"type":    "duration",
				"default": "30s",
				"doc":     "Request timeout",
			},
			want: FieldSpec{
				Key:       "bosun.timeout",
				Scope:     ScopeContainer,
				Type:      TypeDuration,
				Default:   "30s",
				Doc:       "Request timeout",
				FieldName: "Timeout",
			},
		},
		{
			name:      "valid enum field",
			fieldName: "LogLevel",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":     "bosun.logLevel",
				"scope":   "container",
				"type":    "enum",
				"enum":    "debug|info|warn|error",
				"default": "info",
			},
			want: FieldSpec{
				Key:       "bosun.logLevel",
				Scope:     ScopeContainer,
				Type:      TypeEnum,
				Default:   "info",
				Enum:      []string{"debug", "info", "warn", "error"},
				FieldName: "LogLevel",
			},
		},
		{
			name:      "valid bool field with required",
			fieldName: "Enabled",
			goType:    reflect.TypeOf(true),
			tagParts: map[string]string{
				"key":      "bosun.enabled",
				"scope":    "volume",
				"type":     "bool",
				"required": "true",
			},
			want: FieldSpec{
				Key:       "bosun.enabled",
				Scope:     ScopeVolume,
				Type:      TypeBool,
				Required:  true,
				FieldName: "Enabled",
			},
		},
		{
			name:      "valid int field",
			fieldName: "Priority",
			goType:    reflect.TypeOf(0),
			tagParts: map[string]string{
				"key":     "bosun.priority",
				"scope":   "network",
				"type":    "int",
				"default": "10",
			},
			want: FieldSpec{
				Key:       "bosun.priority",
				Scope:     ScopeNetwork,
				Type:      TypeInt,
				Default:   "10",
				FieldName: "Priority",
			},
		},
		{
			name:      "valid size field",
			fieldName: "MaxSize",
			goType:    reflect.TypeOf(int64(0)),
			tagParts: map[string]string{
				"key":     "bosun.maxSize",
				"scope":   "volume",
				"type":    "size",
				"default": "1GB",
			},
			want: FieldSpec{
				Key:       "bosun.maxSize",
				Scope:     ScopeVolume,
				Type:      TypeSize,
				Default:   "1GB",
				FieldName: "MaxSize",
			},
		},
		{
			name:      "valid list field",
			fieldName: "Tags",
			goType:    reflect.TypeOf([]string{}),
			tagParts: map[string]string{
				"key":     "bosun.tags",
				"scope":   "container",
				"type":    "list",
				"default": "a,b,c",
			},
			want: FieldSpec{
				Key:       "bosun.tags",
				Scope:     ScopeContainer,
				Type:      TypeList,
				Default:   "a,b,c",
				FieldName: "Tags",
			},
		},
		{
			name:      "deprecated field",
			fieldName: "OldField",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":        "bosun.old",
				"scope":      "global",
				"type":       "string",
				"deprecated": "true",
			},
			want: FieldSpec{
				Key:        "bosun.old",
				Scope:      ScopeGlobal,
				Type:       TypeString,
				Deprecated: true,
				FieldName:  "OldField",
			},
		},
		// Error cases
		{
			name:      "missing key",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"scope": "global",
				"type":  "string",
			},
			wantErr: true,
			errMsg:  "missing required tag component 'key'",
		},
		{
			name:      "missing scope",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":  "bosun.test",
				"type": "string",
			},
			wantErr: true,
			errMsg:  "missing required tag component 'scope'",
		},
		{
			name:      "missing type",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "bosun.test",
				"scope": "global",
			},
			wantErr: true,
			errMsg:  "missing required tag component 'type'",
		},
		{
			name:      "invalid scope",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "bosun.test",
				"scope": "invalid",
				"type":  "string",
			},
			wantErr: true,
			errMsg:  "invalid scope",
		},
		{
			name:      "invalid type",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "bosun.test",
				"scope": "global",
				"type":  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name:      "enum without values",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "bosun.test",
				"scope": "global",
				"type":  "enum",
			},
			wantErr: true,
			errMsg:  "type=enum requires enum= component",
		},
		{
			name:      "key without bosun prefix",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "test.key",
				"scope": "global",
				"type":  "string",
			},
			wantErr: true,
			errMsg:  "must start with 'bosun.'",
		},
		{
			name:      "type mismatch - string field with bool type",
			fieldName: "Test",
			goType:    reflect.TypeOf(""),
			tagParts: map[string]string{
				"key":   "bosun.test",
				"scope": "global",
				"type":  "bool",
			},
			wantErr: true,
			errMsg:  "requires Go type bool",
		},
		{
			name:      "type mismatch - int field with duration type",
			fieldName: "Test",
			goType:    reflect.TypeOf(0),
			tagParts: map[string]string{
				"key":   "bosun.test",
				"scope": "global",
				"type":  "duration",
			},
			wantErr: true,
			errMsg:  "requires Go type time.Duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFieldSpec(tt.fieldName, tt.goType, tt.tagParts)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseFieldSpec() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("parseFieldSpec() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("parseFieldSpec() unexpected error: %v", err)
				return
			}
			// Compare relevant fields
			if got.Key != tt.want.Key {
				t.Errorf("Key = %q, want %q", got.Key, tt.want.Key)
			}
			if got.Scope != tt.want.Scope {
				t.Errorf("Scope = %q, want %q", got.Scope, tt.want.Scope)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Default != tt.want.Default {
				t.Errorf("Default = %q, want %q", got.Default, tt.want.Default)
			}
			if got.Required != tt.want.Required {
				t.Errorf("Required = %v, want %v", got.Required, tt.want.Required)
			}
			if got.Deprecated != tt.want.Deprecated {
				t.Errorf("Deprecated = %v, want %v", got.Deprecated, tt.want.Deprecated)
			}
			if got.Doc != tt.want.Doc {
				t.Errorf("Doc = %q, want %q", got.Doc, tt.want.Doc)
			}
			if got.FieldName != tt.want.FieldName {
				t.Errorf("FieldName = %q, want %q", got.FieldName, tt.want.FieldName)
			}
			if len(got.Enum) != len(tt.want.Enum) {
				t.Errorf("Enum length = %d, want %d", len(got.Enum), len(tt.want.Enum))
			} else {
				for i, v := range got.Enum {
					if v != tt.want.Enum[i] {
						t.Errorf("Enum[%d] = %q, want %q", i, v, tt.want.Enum[i])
					}
				}
			}
		})
	}
}

// Helper function for error message checking
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test structs for ParseTags tests
type simpleTestConfig struct {
	Name    string        `bosun:"key=bosun.name,scope=global,type=string"`
	Timeout time.Duration `bosun:"key=bosun.timeout,scope=container,type=duration,default=30s"`
	Enabled bool          `bosun:"key=bosun.enabled,scope=container,type=bool,default=true"`
}

type embeddedConfig struct {
	simpleTestConfig
	Extra int `bosun:"key=bosun.extra,scope=global,type=int,default=42"`
}

type nestedEmbeddedConfig struct {
	embeddedConfig
	Deep string `bosun:"key=bosun.deep,scope=global,type=string"`
}

type configWithUntaggedFields struct {
	Tagged   string `bosun:"key=bosun.tagged,scope=global,type=string"`
	Untagged string
	private  string //nolint:unused
}

type configWithEnumField struct {
	Level string `bosun:"key=bosun.level,scope=container,type=enum,enum=debug|info|warn|error,default=info"`
}

type configWithListField struct {
	Tags []string `bosun:"key=bosun.tags,scope=container,type=list,default=a,b,c"`
}

type configWithSizeField struct {
	MaxSize int64 `bosun:"key=bosun.maxSize,scope=volume,type=size,default=1GB"`
}

type invalidConfigMissingKey struct {
	Test string `bosun:"scope=global,type=string"`
}

type invalidConfigDuplicateKey struct {
	First  string `bosun:"key=bosun.dup,scope=global,type=string"`
	Second string `bosun:"key=bosun.dup,scope=global,type=string"`
}

func TestParseTags(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		spec, err := ParseTags[simpleTestConfig]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		if len(spec) != 3 {
			t.Errorf("ParseTags() returned %d fields, want 3", len(spec))
		}

		// Check specific fields
		if fs, ok := spec["bosun.name"]; !ok {
			t.Error("missing key bosun.name")
		} else {
			if fs.Scope != ScopeGlobal {
				t.Errorf("bosun.name Scope = %q, want %q", fs.Scope, ScopeGlobal)
			}
			if fs.Type != TypeString {
				t.Errorf("bosun.name Type = %q, want %q", fs.Type, TypeString)
			}
		}

		if fs, ok := spec["bosun.timeout"]; !ok {
			t.Error("missing key bosun.timeout")
		} else {
			if fs.Default != "30s" {
				t.Errorf("bosun.timeout Default = %q, want %q", fs.Default, "30s")
			}
		}
	})

	t.Run("embedded struct", func(t *testing.T) {
		spec, err := ParseTags[embeddedConfig]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		// Should have fields from both embedded and direct
		if len(spec) != 4 {
			t.Errorf("ParseTags() returned %d fields, want 4", len(spec))
		}

		// Check embedded fields are present
		if _, ok := spec["bosun.name"]; !ok {
			t.Error("missing embedded key bosun.name")
		}
		if _, ok := spec["bosun.extra"]; !ok {
			t.Error("missing direct key bosun.extra")
		}
	})

	t.Run("nested embedded struct", func(t *testing.T) {
		spec, err := ParseTags[nestedEmbeddedConfig]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		// Should have all 5 fields (3 from simple + 1 from embedded + 1 direct)
		if len(spec) != 5 {
			t.Errorf("ParseTags() returned %d fields, want 5", len(spec))
		}
	})

	t.Run("untagged fields are skipped", func(t *testing.T) {
		spec, err := ParseTags[configWithUntaggedFields]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		if len(spec) != 1 {
			t.Errorf("ParseTags() returned %d fields, want 1 (only tagged)", len(spec))
		}
	})

	t.Run("enum field", func(t *testing.T) {
		spec, err := ParseTags[configWithEnumField]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		fs, ok := spec["bosun.level"]
		if !ok {
			t.Fatal("missing key bosun.level")
		}
		if fs.Type != TypeEnum {
			t.Errorf("Type = %q, want %q", fs.Type, TypeEnum)
		}
		if len(fs.Enum) != 4 {
			t.Errorf("Enum has %d values, want 4", len(fs.Enum))
		}
	})

	t.Run("list field", func(t *testing.T) {
		spec, err := ParseTags[configWithListField]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		fs, ok := spec["bosun.tags"]
		if !ok {
			t.Fatal("missing key bosun.tags")
		}
		if fs.Type != TypeList {
			t.Errorf("Type = %q, want %q", fs.Type, TypeList)
		}
	})

	t.Run("size field", func(t *testing.T) {
		spec, err := ParseTags[configWithSizeField]()
		if err != nil {
			t.Fatalf("ParseTags() error: %v", err)
		}

		fs, ok := spec["bosun.maxSize"]
		if !ok {
			t.Fatal("missing key bosun.maxSize")
		}
		if fs.Type != TypeSize {
			t.Errorf("Type = %q, want %q", fs.Type, TypeSize)
		}
		if fs.Default != "1GB" {
			t.Errorf("Default = %q, want %q", fs.Default, "1GB")
		}
	})

	t.Run("error on missing key", func(t *testing.T) {
		_, err := ParseTags[invalidConfigMissingKey]()
		if err == nil {
			t.Error("ParseTags() expected error for missing key")
		}
	})

	t.Run("error on duplicate key", func(t *testing.T) {
		_, err := ParseTags[invalidConfigDuplicateKey]()
		if err == nil {
			t.Error("ParseTags() expected error for duplicate key")
		}
		if err != nil && !contains(err.Error(), "duplicate key") {
			t.Errorf("error = %q, want error containing 'duplicate key'", err.Error())
		}
	})

	t.Run("non-struct type", func(t *testing.T) {
		_, err := ParseTags[string]()
		if err == nil {
			t.Error("ParseTags() expected error for non-struct type")
		}
	})
}
