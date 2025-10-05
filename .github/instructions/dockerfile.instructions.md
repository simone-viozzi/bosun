---
applyTo: "**/Dockerfile,**/dockerfile,**/*.dockerfile"
---

# Dockerfile Guidelines

When modifying Dockerfiles in the Bosun project:

## Current Setup

The main Dockerfile uses a multi-stage build pattern optimized for Go applications.

## Best Practices

### Multi-stage Builds
- Use multi-stage builds to minimize final image size
- Separate build stage from runtime stage
- Only copy necessary artifacts to final stage

### Base Images
- Use official Go images for build stage
- Use minimal base images for runtime (e.g., `alpine`, `distroless`, `scratch`)
- Pin versions for reproducibility: `golang:1.25-alpine` not `golang:latest`

### Build Optimization
- Leverage Docker layer caching by copying `go.mod` and `go.sum` first
- Run `go mod download` before copying source code
- Use `.dockerignore` to exclude unnecessary files

### Example Pattern
```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o bosun ./cmd/bosun

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/bosun /usr/local/bin/
ENTRYPOINT ["bosun"]
```

### Security
- Run as non-root user when possible
- Scan images for vulnerabilities
- Keep base images updated
- Don't include secrets or credentials

### Labels
- Use standard labels for metadata:
  - `org.opencontainers.image.source`
  - `org.opencontainers.image.description`
  - `org.opencontainers.image.version`

### Static Binaries
- Build with `CGO_ENABLED=0` for static binaries
- Use `-trimpath` for reproducible builds
- Consider `-ldflags="-s -w"` to reduce binary size

## Don't
- Don't use `latest` tags in production
- Don't run as root unless absolutely necessary
- Don't include development tools in runtime images
- Don't copy entire project if only specific files are needed
