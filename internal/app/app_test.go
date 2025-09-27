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
