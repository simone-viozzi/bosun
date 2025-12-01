// Package cmd provides the CLI command implementations for Bosun.
package cmd

// Exit codes for Bosun CLI commands.
//
// These constants ensure consistent exit code semantics across all commands:
//   - ExitSuccess: Command completed without errors
//   - ExitRuntimeError: Docker unavailable, I/O failure, network error
//   - ExitValidationError: Invalid labels, missing required fields, config errors
const (
	ExitSuccess         = 0
	ExitRuntimeError    = 1
	ExitValidationError = 2
)
