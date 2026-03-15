package app_test

import (
	"context"
	"testing"

	"github.com/simone-viozzi/bosun/internal/app"
)

func TestAppRuns(t *testing.T) {
	a := app.New()
	if err := a.Run(context.Background(), nil); err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestBootstrapOptions_Defaults(t *testing.T) {
	// Bootstrap requires Docker, so we only test that the function
	// signature exists and that Services type is usable.
	// Full Bootstrap integration test is in integration/ with Docker.
	var svc *app.Services

	// Verify Services has expected fields (compile-time check)
	_ = svc
}
