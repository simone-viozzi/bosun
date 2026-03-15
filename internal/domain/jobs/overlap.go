package jobs

import "fmt"

// OverlapPolicy defines behavior when a job is scheduled while a previous run is active.
type OverlapPolicy string

const (
	// OverlapPolicyQueue delays the next run until the current one completes.
	// Uses robfig/cron/v3 DelayIfStillRunning wrapper.
	OverlapPolicyQueue OverlapPolicy = "queue"

	// OverlapPolicySkip drops the next run if the current one is still active.
	// Uses robfig/cron/v3 SkipIfStillRunning wrapper.
	OverlapPolicySkip OverlapPolicy = "skip"

	// OverlapPolicyCancelRestart stops the current run and starts a fresh one.
	// DEFERRED: Not implemented in M4, tracked in #176.
	OverlapPolicyCancelRestart OverlapPolicy = "cancel-and-restart"
)

// DefaultOverlapPolicy is used when no overlap policy is specified.
const DefaultOverlapPolicy = OverlapPolicyQueue

// ValidateOverlapPolicy returns an error if the policy is invalid or not yet implemented.
func ValidateOverlapPolicy(policy OverlapPolicy) error {
	switch policy {
	case OverlapPolicyQueue, OverlapPolicySkip:
		return nil
	case OverlapPolicyCancelRestart:
		return fmt.Errorf("overlap policy %q is not implemented (tracked in #176)", policy)
	default:
		return fmt.Errorf("invalid overlap policy %q (must be %q or %q)", policy, OverlapPolicyQueue, OverlapPolicySkip)
	}
}

// ParseOverlapPolicy converts a string to an OverlapPolicy.
// Returns DefaultOverlapPolicy for empty strings.
func ParseOverlapPolicy(s string) OverlapPolicy {
	if s == "" {
		return DefaultOverlapPolicy
	}
	return OverlapPolicy(s)
}
