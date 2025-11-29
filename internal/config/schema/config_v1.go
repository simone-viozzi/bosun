package schema

import "time"

// GlobalConfig contains configuration fields that apply to all entity types.
type GlobalConfig struct {
	// Instance is a unique identifier for this Bosun instance.
	Instance string `bosun:"key=bosun.instance,scope=global,type=string,doc='Unique instance identifier'"`
}

// ContainerConfig contains configuration fields specific to containers.
type ContainerConfig struct {
	// StopGracePeriod is the time to wait before forcefully stopping a container.
	StopGracePeriod time.Duration `bosun:"key=bosun.container.stopGracePeriod,scope=container,type=duration,default=30s,doc='Grace period before force stopping'"`

	// HealthCheckInterval is the interval between health checks.
	HealthCheckInterval time.Duration `bosun:"key=bosun.container.healthCheckInterval,scope=container,type=duration,default=30s,doc='Interval between health checks'"`

	// AutoRestart determines whether containers should be automatically restarted.
	AutoRestart bool `bosun:"key=bosun.container.autoRestart,scope=container,type=bool,default=true,doc='Automatically restart containers'"`

	// LogLevel is the logging verbosity level.
	LogLevel string `bosun:"key=bosun.container.logLevel,scope=container,type=enum,enum=debug|info|warn|error,default=info,doc='Logging verbosity level'"`
}

// VolumeConfig contains configuration fields specific to volumes.
type VolumeConfig struct {
	// BackupEnabled determines whether volume backups are enabled.
	BackupEnabled bool `bosun:"key=bosun.volume.backupEnabled,scope=volume,type=bool,default=false,doc='Enable volume backups'"`

	// MaxSize is the maximum size limit for the volume in bytes.
	MaxSize int64 `bosun:"key=bosun.volume.maxSize,scope=volume,type=size,default=10GB,doc='Maximum volume size'"`
}

// NetworkConfig contains configuration fields specific to networks.
type NetworkConfig struct {
	// Priority is the network priority (lower values = higher priority).
	Priority int `bosun:"key=bosun.network.priority,scope=network,type=int,default=100,doc='Network priority (lower = higher priority)'"`
}

// ConfigV1 is the v1 configuration schema for Bosun.
// It demonstrates all supported config types and serves as the source of truth
// for the configuration spec used by the loader, merger, and documentation generator.
//
// Example usage:
//
//	spec, err := schema.ParseTags[ConfigV1]()
//	if err != nil {
//		log.Fatal(err)
//	}
//	// spec now contains all field metadata for config loading
//
//	cfg, err := schema.DefaultOf[ConfigV1]()
//	if err != nil {
//		log.Fatal(err)
//	}
//	// cfg now has all default values populated
type ConfigV1 struct {
	GlobalConfig
	ContainerConfig
	VolumeConfig
	NetworkConfig
}

// V1Spec returns the parsed specification for ConfigV1.
// This is a convenience function that panics on error (for use in init()).
func V1Spec() Spec {
	spec, err := ParseTags[ConfigV1]()
	if err != nil {
		panic("failed to parse ConfigV1 tags: " + err.Error())
	}
	return spec
}

// V1Defaults returns a ConfigV1 instance with all default values populated.
// This is a convenience function that panics on error (for use in init()).
func V1Defaults() ConfigV1 {
	cfg, err := DefaultOf[ConfigV1]()
	if err != nil {
		panic("failed to apply ConfigV1 defaults: " + err.Error())
	}
	return cfg
}
