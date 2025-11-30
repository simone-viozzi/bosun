package configdoc

import (
	"time"

	"github.com/simone-viozzi/bosun/internal/config/schema"
)

// testSpec creates a test schema.Spec with all field types for testing.
func testSpec() schema.Spec {
	return schema.Spec{
		"bosun.container.autoRestart": schema.FieldSpec{
			Key:       "bosun.container.autoRestart",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeBool,
			Default:   "true",
			Doc:       "Automatically restart containers",
			FieldName: "AutoRestart",
		},
		"bosun.container.logLevel": schema.FieldSpec{
			Key:       "bosun.container.logLevel",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeEnum,
			Enum:      []string{"debug", "info", "warn", "error"},
			Default:   "info",
			Doc:       "Logging verbosity level",
			FieldName: "LogLevel",
		},
		"bosun.container.stopGracePeriod": schema.FieldSpec{
			Key:       "bosun.container.stopGracePeriod",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeDuration,
			Default:   "30s",
			Doc:       "Grace period before force stopping",
			FieldName: "StopGracePeriod",
		},
		"bosun.instance": schema.FieldSpec{
			Key:       "bosun.instance",
			Scope:     schema.ScopeGlobal,
			Type:      schema.TypeString,
			Doc:       "Unique instance identifier",
			FieldName: "Instance",
		},
		"bosun.volume.maxSize": schema.FieldSpec{
			Key:       "bosun.volume.maxSize",
			Scope:     schema.ScopeVolume,
			Type:      schema.TypeSize,
			Default:   "10GB",
			Doc:       "Maximum volume size",
			FieldName: "MaxSize",
		},
		"bosun.network.priority": schema.FieldSpec{
			Key:       "bosun.network.priority",
			Scope:     schema.ScopeNetwork,
			Type:      schema.TypeInt,
			Default:   "100",
			Doc:       "Network priority (lower = higher priority)",
			FieldName: "Priority",
		},
	}
}

// testSpecSingleField creates a spec with a single field for simple tests.
func testSpecSingleField() schema.Spec {
	return schema.Spec{
		"bosun.test.field": schema.FieldSpec{
			Key:       "bosun.test.field",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			Default:   "default",
			Doc:       "A test field",
			FieldName: "TestField",
		},
	}
}

// testSpecDeprecated creates a spec with a deprecated field.
func testSpecDeprecated() schema.Spec {
	return schema.Spec{
		"bosun.deprecated.field": schema.FieldSpec{
			Key:        "bosun.deprecated.field",
			Scope:      schema.ScopeContainer,
			Type:       schema.TypeString,
			Default:    "old",
			Doc:        "This field is deprecated",
			Deprecated: true,
			FieldName:  "DeprecatedField",
		},
	}
}

// testSpecAllTypes creates a spec with all config types for comprehensive testing.
func testSpecAllTypes() schema.Spec {
	return schema.Spec{
		"bosun.type.string": schema.FieldSpec{
			Key:       "bosun.type.string",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeString,
			Default:   "hello",
			Doc:       "String type field",
			FieldName: "StringField",
		},
		"bosun.type.bool": schema.FieldSpec{
			Key:       "bosun.type.bool",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeBool,
			Default:   "true",
			Doc:       "Boolean type field",
			FieldName: "BoolField",
		},
		"bosun.type.int": schema.FieldSpec{
			Key:       "bosun.type.int",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeInt,
			Default:   "42",
			Doc:       "Integer type field",
			FieldName: "IntField",
		},
		"bosun.type.duration": schema.FieldSpec{
			Key:       "bosun.type.duration",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeDuration,
			Default:   "5m",
			Doc:       "Duration type field",
			FieldName: "DurationField",
		},
		"bosun.type.size": schema.FieldSpec{
			Key:       "bosun.type.size",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeSize,
			Default:   "1GB",
			Doc:       "Size type field",
			FieldName: "SizeField",
		},
		"bosun.type.enum": schema.FieldSpec{
			Key:       "bosun.type.enum",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeEnum,
			Enum:      []string{"a", "b", "c"},
			Default:   "a",
			Doc:       "Enum type field",
			FieldName: "EnumField",
		},
		"bosun.type.list": schema.FieldSpec{
			Key:       "bosun.type.list",
			Scope:     schema.ScopeContainer,
			Type:      schema.TypeList,
			Default:   "x,y,z",
			Doc:       "List type field",
			FieldName: "ListField",
		},
	}
}

// Ensure time package is available for duration tests
var _ = time.Second
