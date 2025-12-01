// Package testutil provides test utilities and infrastructure for Bosun.
//
// # Testing Philosophy
//
// Bosun follows a layered testing strategy with clear responsibilities:
//
// ## Unit Tests
//
// Unit tests (files ending in _test.go without build tags) verify:
//   - Individual function behavior with mocked dependencies
//   - Edge cases and error conditions
//   - Business logic in domain/ports/adapters layers
//   - Input validation and parsing
//
// Unit tests should:
//   - Be fast (no Docker, no network)
//   - Use table-driven test patterns
//   - Cover error paths thoroughly
//   - Mock external dependencies
//
// ## Integration Tests
//
// Integration tests (files with //go:build integration) verify:
//   - End-to-end CLI command behavior
//   - Docker API interactions
//   - Real compose stack orchestration
//   - Output format correctness (JSON/YAML/text parsing)
//
// Integration tests should:
//   - Use the --project flag for isolation between parallel tests
//   - Focus on happy path scenarios
//   - Avoid duplicating edge case coverage from unit tests
//   - Use compose fixtures from testutil/compose/
//
// ## Test Isolation
//
// Integration tests use unique Docker Compose project names (generated from
// test name + timestamp) to prevent interference when running in parallel.
// The StartCompose helper returns a Stack with a Project field that should
// be passed to CLI commands via --project flag.
//
// ## Compose Fixtures
//
// Test compose files live in internal/testutil/compose/:
//   - dockerlabels-compose.yaml: Basic containers with bosun.* config labels
//   - joblabels-compose.yaml: Containers and volumes with bosun.job.* labels
//   - joblabels-invalid-compose.yaml: Invalid labels for validation testing
//
// Each compose file should have a header comment explaining its purpose
// and which tests use it.
package testutil
