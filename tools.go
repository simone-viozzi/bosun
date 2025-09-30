//go:build tools
// +build tools

// Package tools imports dependencies that are used for tooling and testing
// but not directly in the application code. This ensures they remain pinned
// in go.mod until they are actually imported elsewhere.
//
// TODO: Remove this file once the dependencies are used in the actual codebase
// (e.g., in integration tests or adapters).
package tools

import (
	_ "github.com/docker/docker/client"
)
