# Bosun Configuration Reference

> Auto-generated from ConfigV1 schema. Do not edit manually.

## Global Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.instance` | string | - | No | - | Unique instance identifier |

## Container Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.container.autoRestart` | boolean | true | No | - | Automatically restart containers |
| `bosun.container.healthCheckInterval` | duration | 30s | No | - | Interval between health checks |
| `bosun.container.logLevel` | enum | info | No | debug \| info \| warn \| error | Logging verbosity level |
| `bosun.container.stopGracePeriod` | duration | 30s | No | - | Grace period before force stopping |

## Volume Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.volume.enabled` | boolean | false | No | - | Enable volume processing |
| `bosun.volume.maxSize` | byte size | 10GB | No | - | Maximum volume size |

## Network Configuration

| Key | Type | Default | Required | Allowed Values | Description |
|-----|------|---------|----------|----------------|-------------|
| `bosun.network.priority` | integer | 100 | No | - | Network priority (lower = higher priority) |

## Value Formats

### Duration

Go duration syntax. Supported units: `ns`, `us`/`µs`, `ms`, `s`, `m`, `h`.

**Examples:** `30s`, `5m`, `1h30m`, `100ms`

### Byte Size

Docker/go-units byte size syntax. Supports base-10 (KB, MB, GB, TB) and base-2 (KiB, MiB, GiB, TiB) units.

**Examples:** `100MB`, `1GiB`, `500KB`, `10GB`

### List

List values can be specified as CSV (comma-separated) or JSON array syntax.

**Examples:** `value1,value2,value3`, `["value1", "value2"]`
