---
applyTo: "**/*_test.go"
---

# Go Test Files Guidelines

When writing or modifying test files in Bosun, follow these patterns:

## Test Organization

### Unit Tests
- Located alongside source files (e.g., `app.go` → `app_test.go`)
- No build tags required
- Should be fast (< 100ms per test)
- Mock external dependencies (Docker, HTTP, databases)

### Integration Tests
- Located in `integration/` package
- **MUST** include build tag: `//go:build integration`
- Use real Docker services via `internal/testutil/` harness
- Test complete workflows end-to-end
- Can be slower (typically 1-2 minutes)

## Test Naming

- Test functions: `TestFunctionName` or `TestType_Method`
- Subtests: Use `t.Run("descriptive name", func(t *testing.T) { ... })`
- Integration tests: `Test_Integration_FeatureName` to clearly identify them

## Test Structure (AAA Pattern)

```go
func TestFeature(t *testing.T) {
    // Arrange - Set up test conditions
    input := setupInput()
    expected := expectedOutput()

    // Act - Execute the code under test
    result := doSomething(input)

    // Assert - Verify the outcome
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

## Integration Test Pattern

Use the testutil harness for Docker-based tests:

```go
//go:build integration

package integration

func Test_Integration_MyFeature(t *testing.T) {
    t.Parallel()  // Run tests in parallel when possible

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // Start Docker Compose stack
    stack := testutil.StartCompose(t, ctx, "docker-compose.yaml")

    // Get service port
    port := testutil.HostPort(t, ctx, stack.Project, "service-name", 80)

    // Test your feature...
    // Cleanup happens automatically via t.Cleanup()
}
```

## Assertions

- Use standard library `testing` package
- For complex assertions, consider helper functions
- Prefer `t.Errorf()` over `t.Fatalf()` unless test cannot continue
- Use `t.Helper()` in assertion helper functions

## Test Data

- Use table-driven tests for multiple similar cases:
```go
func TestMultipleCases(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"zero", 0, 0},
        {"positive", 5, 25},
        {"negative", -2, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := square(tt.input)
            if result != tt.expected {
                t.Errorf("got %d, want %d", result, tt.expected)
            }
        })
    }
}
```

## Mocking

- Use interfaces for dependencies that need mocking
- Keep mocks simple and focused
- Consider using `internal/ports/` interfaces for easy mocking

## Running Tests

```bash
# Unit tests only (fast)
make test

# Integration tests only (requires Docker)
make it

# Integration tests with verbose output
make itv

# All tests
make test && make it
```

## Test Utilities

- Use `internal/testutil/` for integration test helpers
- `StartCompose()` - Launch Docker Compose stacks
- `HostPort()` - Get published ports for services
- `DumpLogs()` - Save container logs for debugging

## Performance

- Keep unit tests under 100ms
- Use `t.Parallel()` for independent integration tests
- Avoid unnecessary sleeps (use proper synchronization)

## Coverage

- Aim for good coverage of critical paths
- Don't chase 100% coverage at the expense of test quality
- Focus on testing behavior, not implementation details
