//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Placeholders for future janitor and shared setup.
	// e.g., configure logging dir, global timeouts, etc.

	code := m.Run()
	os.Exit(code)
}
