// Package merge provides functionality for combining configuration values
// from multiple sources with defined precedence.
//
// The main entry point is [Merge], which combines configuration from defaults,
// config file, environment variables, and Docker labels into a single
// [schema.ConfigV1] struct.
//
// # Precedence
//
// Configuration sources are merged with the following precedence (lowest to highest):
//  1. defaults - Built-in defaults from schema tags
//  2. file - Config file values (if provided)
//  3. env - Environment variables (if opts.EnableEnv && env != nil)
//  4. labels - Docker label values (if provided)
//
// Higher precedence sources override lower ones. Zero values in higher layers
// are treated as "not set" and don't override lower layers.
//
// # Example
//
//	spec, _ := schema.V1Spec()
//	defaults, _ := schema.V1Defaults()
//	labelsCfg, _ := loader.FromLabels(spec, labels, scope)
//
//	merged, err := merge.Merge(spec, defaults, nil, nil, &labelsCfg, merge.Options{})
//	if err != nil {
//	    return err
//	}
//	// Use merged config...
package merge
