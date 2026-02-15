package jobs_test

import (
	"testing"

	"github.com/simone-viozzi/bosun/internal/domain/jobs"
)

func TestValidateOverlapPolicy_Valid(t *testing.T) {
	tests := []struct {
		name   string
		policy jobs.OverlapPolicy
	}{
		{"queue", jobs.OverlapPolicyQueue},
		{"skip", jobs.OverlapPolicySkip},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := jobs.ValidateOverlapPolicy(tt.policy); err != nil {
				t.Errorf("ValidateOverlapPolicy(%q) = %v, want nil", tt.policy, err)
			}
		})
	}
}

func TestValidateOverlapPolicy_DeferredCancelRestart(t *testing.T) {
	err := jobs.ValidateOverlapPolicy(jobs.OverlapPolicyCancelRestart)
	if err == nil {
		t.Fatal("ValidateOverlapPolicy(cancel-and-restart) should return error (deferred to #176)")
	}
}

func TestValidateOverlapPolicy_Invalid(t *testing.T) {
	err := jobs.ValidateOverlapPolicy(jobs.OverlapPolicy("bogus"))
	if err == nil {
		t.Fatal("ValidateOverlapPolicy(bogus) should return error")
	}
}

func TestParseOverlapPolicy_Empty(t *testing.T) {
	got := jobs.ParseOverlapPolicy("")
	if got != jobs.DefaultOverlapPolicy {
		t.Errorf("ParseOverlapPolicy(\"\") = %q, want %q", got, jobs.DefaultOverlapPolicy)
	}
}

func TestParseOverlapPolicy_Passthrough(t *testing.T) {
	got := jobs.ParseOverlapPolicy("skip")
	if got != jobs.OverlapPolicySkip {
		t.Errorf("ParseOverlapPolicy(\"skip\") = %q, want %q", got, jobs.OverlapPolicySkip)
	}
}
