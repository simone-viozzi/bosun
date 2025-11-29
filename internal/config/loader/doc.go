// Package loader provides functionality for parsing Docker labels into typed
// configuration values according to a schema specification.
//
// The main entry point is [FromLabels], which takes a [schema.Spec] describing
// the configuration schema, a map of Docker labels, and a [schema.Scope]
// indicating the entity type (container, volume, network), and returns a
// populated [schema.ConfigV1] struct.
//
// # Label Filtering
//
// Only labels with the "bosun." prefix are processed. All other labels are
// silently ignored. This allows bosun configuration to coexist with other
// Docker label systems.
//
// # Validation
//
// The loader performs comprehensive validation:
//   - Unknown keys (bosun.* labels not in the schema) are rejected
//   - Scope mismatches (e.g., container label on volume) are rejected
//   - Type parse failures (e.g., invalid duration) are rejected
//
// All validation errors are collected and returned together as [ValidationErrors],
// enabling users to see all issues at once rather than fixing them one at a time.
//
// # Example
//
//	spec, _ := schema.V1Spec()
//	labels := map[string]string{
//	    "bosun.container.stopGracePeriod": "30s",
//	    "bosun.container.autoRestart":     "true",
//	}
//
//	cfg, err := loader.FromLabels(spec, labels, schema.ScopeContainer)
//	if err != nil {
//	    var verrs loader.ValidationErrors
//	    if errors.As(err, &verrs) {
//	        for _, ve := range verrs {
//	            fmt.Printf("Validation error: %s\n", ve.Message)
//	        }
//	    }
//	    return
//	}
//	// Use cfg...
package loader
