---
applyTo: "**/docker-compose.yml,**/docker-compose.yaml,**/*.docker-compose.yml,**/*.docker-compose.yaml"
---

# Docker Compose Guidelines

When working with Docker Compose files in Bosun:

## Usage in Bosun

Docker Compose files are primarily used for:
1. Integration testing (in `internal/testutil/compose/`)
2. Local development environments
3. Test fixtures for label discovery features

## Best Practices

### Version
- Use Compose file version 3.8 or later (or omit for latest spec)
- Modern syntax doesn't require explicit version field

### Service Naming
- Use descriptive, lowercase names
- Avoid special characters (stick to alphanumeric and hyphens)
- Match service names to their purpose (e.g., `nginx`, `postgres`, `redis`)

### Networks
- Define explicit networks when services need to communicate
- Use bridge network for isolation during tests
- Name networks descriptively

### Volumes
- Use named volumes for persistent data
- Use bind mounts for development only
- Document volume purposes in comments

### Ports
- Always use `"host:container"` format for clarity
- Consider using `expose` instead of `ports` for internal-only services
- Document non-standard ports

### Labels
- For Bosun testing, use `bosun.*` prefix for labels
- Labels are key to testing label discovery features
- Example: `bosun.role: "web"`, `bosun.env: "test"`

### Environment Variables
- Use `.env` files for local development
- Don't commit sensitive values
- Provide `.env.example` templates
- Use `${VAR:-default}` syntax for defaults

### Health Checks
- Define health checks for critical services
- Keep intervals reasonable (10-30s)
- Use appropriate timeouts and retries

## Testing Pattern

For integration tests, keep compose files simple:

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80"  # Random host port assigned
    labels:
      bosun.test: "true"
      bosun.service: "web"
```

## Don't
- Don't use `latest` tags (pin specific versions)
- Don't include production credentials
- Don't make tests depend on external services when possible
- Don't use complex networking without good reason

## Integration with testutil

When creating compose files for tests:
- Store in `internal/testutil/compose/`
- Use embedded filesystem (`ComposeFS`) for access
- Files are referenced by name in `testutil.StartCompose()`
- Cleanup happens automatically via testcontainers
