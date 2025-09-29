# MCP Time Server

[![Go Reference](https://pkg.go.dev/badge/github.com/hspedro/mcp-server-time.svg)](https://pkg.go.dev/github.com/hspedro/mcp-make server-time)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A Model Context Protocol (MCP) server providing time and timezone tools. Built with Go and the official MCP Go SDK, supporting both SSE and Streamable transports.

## Features

### 🕐 **Time Operations**
- **Current Time**: Get current time in any timezone with flexible formatting
- **Time Formatting**: Convert timestamps between different formats (RFC3339, Unix, custom layouts)
- **Time Parsing**: Parse time strings with auto-detection or explicit formats
- **Timezone Info**: Comprehensive timezone information including DST transitions

### 🌐 **Protocol Support**
- **SSE Transport**: Real-time Server-Sent Events for persistent connections
- **Streamable Transport**: HTTP request/response for stateless operations
- **MCP Compliant**: Full compatibility with MCP protocol v0.8.0

### 📊 **Observability**
- **Prometheus Metrics**: Detailed metrics for requests, operations, and errors
- **Structured Logging**: JSON and console logging with configurable levels
- **Health Checks**: Kubernetes-ready health endpoints

### 🏗️ **Production Ready**
- **Multi-Architecture**: ARM64 and AMD64 Docker images
- **Graceful Shutdown**: Proper signal handling and connection draining
- **Configuration**: YAML config with environment variable overrides
- **Security**: Non-root container execution

## Quick Start

### Using Docker

```bash
# Run with default configuration
docker run -p 8080:8080 -p 9080:9080 ghcr.io/hspedro/mcp-server-time:latest

# Run with custom configuration
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml ghcr.io/hspedro/mcp-server-time:latest
```

### Using Go

```bash
# Clone and build
git clone https://github.com/hspedro/mcp-server-time.git
cd mcp-server-time
make build

# Run locally
./mcp-server-time
```

## MCP Tools

### `get_time`
Get current time with optional timezone and format specification.

**Input:**
```json
{
  "timezone": "America/New_York",  // Optional, defaults to UTC
  "format": "RFC3339"              // Optional, defaults to RFC3339
}
```

**Output:**
```json
{
  "current_time": "2023-12-25T10:30:45-05:00",
  "timezone": "America/New_York",
  "format": "RFC3339",
  "timestamp_utc": "2023-12-25T15:30:45Z",
  "unix_timestamp": 1703520645
}
```

### `format_time`
Format a timestamp using custom formats with optional timezone conversion.

**Input:**
```json
{
  "timestamp": "2023-12-25T15:30:45Z",  // Required: string or number
  "format": "Unix",                    // Required: output format
  "timezone": "America/New_York"       // Optional: target timezone
}
```

### `parse_time`
Parse time strings with auto-detection or explicit format specification.

**Input:**
```json
{
  "time_string": "December 25, 2023 3:30 PM",  // Required
  "format": "",                                // Optional: auto-detect if empty
  "timezone": "America/New_York"               // Optional: assume timezone
}
```

### `timezone_info`
Get comprehensive timezone information including DST transitions.

**Input:**
```json
{
  "timezone": "America/New_York",              // Required
  "reference_time": "2023-12-25T15:30:45Z"    // Optional: defaults to now
}
```

## Configuration

### YAML Configuration
```yaml
server:
  name: "mcp-server-time"
  version: "1.0.0"
  host: "localhost"
  port: 8080
  graceful_shutdown_timeout: 30s

time:
  default_timezone: "UTC"
  default_format: "RFC3339"
  supported_formats:
    - "RFC3339"
    - "RFC3339Nano"
    - "Unix"
    - "UnixMilli"
    - "UnixMicro"
    - "UnixNano"
    - "Layout"

logging:
  level: "info"        # debug, info, warn, error, fatal
  format: "json"       # json, console

metrics:
  enabled: true
  port: 9080
  path: "/metrics"
```

### Environment Variables
```bash
# Server configuration
MCP_SERVER_HOST=0.0.0.0
MCP_SERVER_PORT=8080

# Time service configuration
MCP_TIME_DEFAULT_TIMEZONE=America/New_York
MCP_TIME_DEFAULT_FORMAT=RFC3339

# Logging configuration
MCP_LOGGING_LEVEL=debug
MCP_LOGGING_FORMAT=console

# Metrics configuration
MCP_METRICS_ENABLED=true
MCP_METRICS_PORT=9080
```

## Endpoints

### MCP Transports
- **SSE**: `GET /sse` - Server-Sent Events transport
- **Streamable**: `POST /streamable` - HTTP request/response transport
- **MCP**: `POST /mcp` - Alias for streamable transport

### Monitoring
- **Health**: `GET /health` - Health check endpoint
- **Metrics**: `GET /metrics` - Prometheus metrics (if enabled)

## Development

### Prerequisites
- Go 1.23+
- Docker (optional)
- Make

### Development Commands
```bash
# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# Generate mocks
make mocks

# Build binary
make build

# Run locally
make run

# Complete verification
make verify
```

## MCP Client Integration

### Cursor IDE
Add to your MCP configuration:

```json
{
  "mcpServers": {
    "time": {
      "type": "url",
      "url": "http://localhost:8080/streamable",
      "transport": "streamable"
    }
  }
}
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
