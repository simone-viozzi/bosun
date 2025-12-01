package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/simone-viozzi/bosun/internal/adapters/dockerlabels"
	"github.com/simone-viozzi/bosun/internal/config/loader"
	"github.com/simone-viozzi/bosun/internal/config/merge"
	"github.com/simone-viozzi/bosun/internal/config/schema"
	dlabels "github.com/simone-viozzi/bosun/internal/domain/labels"
	"github.com/simone-viozzi/bosun/internal/ports"
)

// ConfigSource indicates where config values come from
type ConfigSource string

const (
	// SourceAuto merges all sources (default)
	SourceAuto ConfigSource = "auto"
	// SourceLabels uses Docker labels only
	SourceLabels ConfigSource = "labels"
	// SourceFile uses config file only (future)
	SourceFile ConfigSource = "file"
)

// ValidateOptions holds CLI flag values
type ValidateOptions struct {
	Source         ConfigSource // --from flag
	Scope          string       // --scope flag (empty = all)
	PrintConfig    bool         // --print flag
	ConfigFile     string       // --config flag (future)
	IncludeStopped bool         // --stopped flag
}

// EntityValidationError wraps validation errors with entity context
type EntityValidationError struct {
	Entity dlabels.LabeledEntity   // The entity that failed
	Errors loader.ValidationErrors // Validation errors for this entity
}

// ValidationResult holds the outcome of validation
type ValidationResult struct {
	Valid        bool                       // Overall success
	MergedConfig *schema.ConfigV1           // Merged config (if valid or for --print)
	EntityErrors []EntityValidationError    // Per-entity errors (config labels)
	JobErrors    loader.JobValidationErrors // Job label validation errors
	Warnings     []string                   // Non-fatal warnings
}

// NewValidateCmd creates the validate subcommand
func NewValidateCmd() *cobra.Command {
	opts := ValidateOptions{
		Source: SourceAuto,
	}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration from labels and files",
		Long: `Validates Bosun configuration from Docker labels and config files.

Checks for:
- Unknown label keys (typos)
- Invalid value types (e.g., bad duration format)
- Scope mismatches (container label on volume)
- Required fields

Exit codes:
  0 - Validation passed
  1 - Runtime error (Docker unavailable, etc.)
  2 - Validation failed (invalid config)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context(), opts)
		},
	}

	// Define flags
	cmd.Flags().StringVarP((*string)(&opts.Source), "from", "f", string(SourceAuto),
		"Config source: auto, labels, file")
	cmd.Flags().StringVarP(&opts.Scope, "scope", "s", "",
		"Validate only: container, volume, network, global (default: all)")
	cmd.Flags().BoolVarP(&opts.PrintConfig, "print", "p", false,
		"Print merged config as JSON")
	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "",
		"Path to config file")
	cmd.Flags().BoolVar(&opts.IncludeStopped, "stopped", false,
		"Include stopped containers")

	return cmd
}

// runValidate executes the validation logic
func runValidate(ctx context.Context, opts ValidateOptions) error {
	// Handle --from file (not implemented)
	if opts.Source == SourceFile {
		fmt.Fprintln(os.Stderr, "Warning: --from file is not yet implemented. Validating defaults only.")
	}

	// Connect to Docker
	source, err := dockerlabels.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to Docker: %v\nIs Docker running?\n", err)
		os.Exit(ExitRuntimeError)
		return nil // unreachable, but satisfies compiler
	}

	// Load snapshot
	selector := ports.Selector{
		Prefixes:       []string{dlabels.DefaultLabelPrefix},
		IncludeStopped: opts.IncludeStopped,
	}

	snapshot, err := source.Snapshot(ctx, selector)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get snapshot: %v\n", err)
		os.Exit(ExitRuntimeError)
		return nil
	}

	// Filter by scope if specified
	entities := filterByScope(snapshot.Entities, opts.Scope)

	// Validate each entity (config labels)
	result := validateEntities(entities, opts)

	// Validate job labels
	jobResult := loader.ValidateJobLabels(snapshot.Entities) // Always validate all entities for job cross-checks
	if !jobResult.IsValid() {
		result.Valid = false
		result.JobErrors = jobResult.Errors
	}
	result.Warnings = append(result.Warnings, jobResult.Warnings...)

	// Output results
	return outputResults(result, opts)
}

// entityKindToScope maps entity kinds to schema scopes
func entityKindToScope(kind dlabels.Kind) schema.Scope {
	switch kind {
	case dlabels.KindContainer:
		return schema.ScopeContainer
	case dlabels.KindVolume:
		return schema.ScopeVolume
	case dlabels.KindNetwork:
		return schema.ScopeNetwork
	default:
		return schema.ScopeGlobal
	}
}

// scopeStringToKind converts scope flag string to entity kind filter
func scopeStringToKind(scope string) (dlabels.Kind, bool) {
	switch scope {
	case "container":
		return dlabels.KindContainer, true
	case "volume":
		return dlabels.KindVolume, true
	case "network":
		return dlabels.KindNetwork, true
	case "global", "":
		return "", false // no filter
	default:
		return "", false
	}
}

// filterByScope filters entities by the --scope flag
func filterByScope(entities []dlabels.LabeledEntity, scope string) []dlabels.LabeledEntity {
	if scope == "" || scope == "global" {
		return entities // no filter, validate all
	}

	filterKind, hasFilter := scopeStringToKind(scope)
	if !hasFilter {
		return entities
	}

	var filtered []dlabels.LabeledEntity
	for _, e := range entities {
		if e.Kind == filterKind {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// filterJobLabels removes bosun.job.* labels from a map since they are validated separately.
func filterJobLabels(labels map[string]string) map[string]string {
	filtered := make(map[string]string)
	for k, v := range labels {
		if !strings.HasPrefix(k, "bosun.job.") {
			filtered[k] = v
		}
	}
	return filtered
}

// validateEntities validates all entities and collects errors
func validateEntities(entities []dlabels.LabeledEntity, opts ValidateOptions) ValidationResult {
	spec := schema.V1Spec()
	result := ValidationResult{
		Valid: true,
	}

	// Collect all entity configs for merging
	var labelConfigs []schema.ConfigV1

	for _, entity := range entities {
		scope := entityKindToScope(entity.Kind)

		// Filter out job labels - they are validated separately by ValidateJobLabels
		configLabels := filterJobLabels(entity.Labels)

		// Validate config labels for this entity
		cfg, err := loader.FromLabels(spec, configLabels, scope)
		if err != nil {
			var verrs loader.ValidationErrors
			if errors.As(err, &verrs) {
				result.Valid = false
				result.EntityErrors = append(result.EntityErrors, EntityValidationError{
					Entity: entity,
					Errors: verrs,
				})
			}
		}

		// Collect config for merging (even if there were errors)
		labelConfigs = append(labelConfigs, cfg)
	}

	// Skip labels if --from file
	if opts.Source == SourceFile {
		labelConfigs = nil
	}

	// Merge configurations
	defaults := schema.V1Defaults()
	var mergedLabels *schema.ConfigV1

	if len(labelConfigs) > 0 && opts.Source != SourceFile {
		// Merge all label configs together (last wins for conflicts)
		merged := defaults
		for _, cfg := range labelConfigs {
			merged, _ = merge.Merge(spec, merged, nil, nil, &cfg, merge.Options{})
		}
		mergedLabels = &merged
	}

	// Final merge: defaults -> file (future) -> labels
	finalConfig, _ := merge.Merge(spec, defaults, nil, nil, mergedLabels, merge.Options{})
	result.MergedConfig = &finalConfig

	return result
}

// outputResults prints validation results and returns appropriate error
func outputResults(result ValidationResult, opts ValidateOptions) error {
	if !result.Valid {
		// Print errors to stderr
		fmt.Fprintln(os.Stderr, "Validation errors:")
		fmt.Fprintln(os.Stderr)

		totalErrors := 0

		// Print config label errors (entity-grouped)
		for _, entityErr := range result.EntityErrors {
			fmt.Fprintf(os.Stderr, "%s %q (%s):\n",
				entityErr.Entity.Kind,
				entityErr.Entity.Name,
				truncateID(entityErr.Entity.ID))

			for _, verr := range entityErr.Errors {
				fmt.Fprintf(os.Stderr, "  - %s\n", verr.Message)
				totalErrors++
			}
			fmt.Fprintln(os.Stderr)
		}

		// Print job label errors
		if len(result.JobErrors) > 0 {
			fmt.Fprintln(os.Stderr, "Job label errors:")
			for _, jerr := range result.JobErrors {
				fmt.Fprintf(os.Stderr, "  %s %q: %s\n",
					jerr.Entity.Kind,
					jerr.Entity.Name,
					jerr.Message)
				totalErrors++
			}
			fmt.Fprintln(os.Stderr)
		}

		fmt.Fprintf(os.Stderr, "Found %d error(s)\n", totalErrors)

		os.Exit(ExitValidationError)
		return nil
	}

	// Print warnings (even on success)
	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Warnings:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Success path
	if opts.PrintConfig && result.MergedConfig != nil {
		// Print merged config as JSON
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result.MergedConfig); err != nil {
			return fmt.Errorf("failed to encode config: %w", err)
		}
	} else {
		// Print success message
		fmt.Printf("Configuration valid\n")
	}

	return nil
}

// truncateID returns first 12 chars of an ID (like Docker does)
func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
