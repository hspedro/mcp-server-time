# Claude Assistant Guide - MCP Time Server

## Project Overview

The MCP Time Server is a Go-based Model Context Protocol (MCP) server that provides time-related tools and utilities. It implements the official MCP Go SDK to offer timezone-aware time operations, formatting, parsing, and conversion capabilities.

## Architecture

This project follows **Clean Architecture** principles with interface-based design. Key principle: **Dependencies point inward** - the time service has no external dependencies except Go's standard time package.

### Quick Structure
```
internal/
├── time/        # Core time service interface and implementation
├── handlers/    # MCP tool handlers (get_time, format_time, etc.)
├── transport/   # SSE and Streamable transports for MCP communication
├── features/    # MCP tool and prompt definitions
├── metrics/     # Prometheus metrics (excluding Go runtime)
└── config/      # Configuration management with Viper
```

## Essential Documentation

### Must Read First
1. **[tech_context.md](./tech_context.md)** - MCP patterns, Go architecture, and library choices
2. **[agent_workflow.md](./agent_workflow.md)** - **CRITICAL**: Standard workflow for AI agents
3. **[README.md](../README.md)** - Project overview and setup

### Agent Workflow Requirements
**IMPORTANT**: All AI agents must follow the [agent_workflow.md](./agent_workflow.md) process:
1. Load tech_context.md and understand MCP patterns
2. Analyze user request against project architecture
3. Create comprehensive plan including code, tests, config, and docs
4. **Get user validation BEFORE making any changes**
5. Implement in proper order with quality gates

## Key Design Patterns

### 1. MCP Protocol Compliance
- Uses official MCP Go SDK v0.8.0
- Implements both SSE and Streamable transports
- Standard tool and prompt definitions
- Proper error response formatting

### 2. Interface-Based Time Service
```go
// Core time operations interface
type TimeService interface {
    GetCurrentTime(timezone string) (time.Time, error)
    FormatTime(t time.Time, format string) (string, error)
    ParseTime(timeStr, format string) (time.Time, error)
    GetTimezoneInfo(timezone string) (*TimezoneInfo, error)
}

// Implementation with dependency injection
type timeService struct {
    defaultTimezone  string
    supportedFormats []string
    logger          *zap.Logger
}
```

### 3. MCP Tool Handlers
```go
// Each tool follows MCP patterns
func (h *GetTimeHandler) Handle(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Parse input according to tool schema
    // Call time service
    // Return MCP-compliant response
}
```

## MCP Tools

Current tools implemented:

### get_time
Get current time with optional timezone and format
- **Input**: `timezone` (string, optional), `format` (string, optional)
- **Output**: Current time in specified timezone and format
- **Default**: UTC timezone, RFC3339 format

### format_time
Format a given timestamp using custom formats
- **Input**: `timestamp` (string/number), `format` (string)
- **Output**: Formatted time string
- **Supports**: RFC3339, Unix timestamps, custom Go layouts

### parse_time
Parse time strings in various formats
- **Input**: `time_string` (string), `format` (string, optional)
- **Output**: Parsed time in RFC3339 format
- **Auto-detection**: Common formats if format not specified

### timezone_info
Get timezone information and perform conversions
- **Input**: `timezone` (string), `reference_time` (string, optional)
- **Output**: Timezone details, offset, DST info
- **Features**: All IANA timezones supported

## Common Tasks

### Adding a New Time Tool

1. **Define tool schema** in `internal/features/tools.go`
2. **Create handler** in `internal/handlers/new_tool.go`
3. **Add handler tests** in `internal/handlers/new_tool_test.go`
4. **Register tool** in main.go
5. **Update documentation**

### Testing

```bash
make test              # Run all tests
make test-unit         # Unit tests only
make test-integration  # Integration tests
make mocks             # Regenerate mocks
```

### Local Development

```bash
make run               # Run server locally
make fmt               # Format code
make lint              # Run linters
make build             # Build binary
```

## Configuration

### Environment Variables

Core configs:
```bash
TIME_SERVER_PORT=8080
TIME_SERVER_HOST=localhost
TIME_DEFAULT_TIMEZONE=UTC
TIME_LOGGING_LEVEL=info
TIME_METRICS_ENABLED=true
```

### YAML Configuration
```yaml
server:
  name: "time-mcp-server"
  port: 8080
  host: "localhost"

time:
  default_timezone: "UTC"
  default_format: "RFC3339"
  supported_formats:
    - "RFC3339"
    - "Unix"
    - "UnixMilli"

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"
```

## MCP Integration

### Transports
- **SSE Transport**: `/sse` - Server-Sent Events for real-time communication
- **Streamable Transport**: `/streamable` or `/mcp` - HTTP request/response
- **Health Check**: `/health` - Kubernetes-ready health endpoint
- **Metrics**: `/metrics` - Prometheus metrics endpoint

### Tool Registration
```go
// In main.go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "time-mcp-server",
    Version: "1.0.0",
}, nil)

// Register tools
mcp.AddTool(server, features.GetTimeTool, handlers.NewGetTimeHandler(timeService))
mcp.AddTool(server, features.FormatTimeTool, handlers.NewFormatTimeHandler(timeService))
// ...
```

## Data Flow

1. **MCP Request** → `transport` → `handler` → `time service` → `response`
2. **Health Check** → `health endpoint` → `status`
3. **Metrics** → `prometheus endpoint` → `metrics data`

## Code Style

- Use interface-based design for testability
- Follow MCP protocol standards strictly
- Handle timezone operations carefully
- Add structured logging with context
- Write tests for all time operations
- Validate all time inputs thoroughly

## Monitoring

Key metrics:
- Request latency by tool type
- Time operation success/failure rates
- Invalid timezone/format attempts
- Concurrent request handling

Available at `/metrics`:
- `mcp_time_requests_total{tool="get_time",error="false"}`
- `mcp_time_request_duration_seconds{tool="format_time"}`
- `mcp_time_operations_total{operation="timezone_conversion"}`

## Troubleshooting

Common issues:

1. **"Invalid timezone"** - Check IANA timezone database
2. **"Parse error"** - Verify time format strings
3. **"SSE connection failed"** - Check transport configuration
4. **"Tool not found"** - Verify tool registration in main.go

### Debug Mode
```bash
TIME_LOGGING_LEVEL=debug make run
```

## MCP Client Integration

### Cursor IDE Integration
Add to MCP configuration:
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

### With MCP Proxy
```yaml
# In MCP proxy configuration
servers:
  time:
    url: "http://mcp-server-time:8080/sse"
    transport: "sse"
```

## Development Workflow

### Quality Gates
1. `make fmt` - Code formatting
2. `make lint` - Linting passes
3. `make test` - All tests pass
4. `make mocks` - Mocks are current
5. `make build` - Build succeeds

### File Organization
- Tests alongside code (`handler_test.go` next to `handler.go`)
- Mocks in dedicated `mocks/` directories
- Configuration in `config.yaml` with env overrides
- Documentation in `docs/` directory

## Getting Help

1. Check existing documentation in `docs/`
2. Review test files for usage examples
3. Reference MCP Go SDK documentation
4. Look at git history for implementation patterns
5. Ask about design decisions before major changes

## Quick Reference

### Time Formats
- `RFC3339`: `2006-01-02T15:04:05Z07:00`
- `Unix`: `1609459200` (seconds since epoch)
- `UnixMilli`: `1609459200000` (milliseconds since epoch)
- `Custom`: Go time layout strings

### Common Timezones
- `UTC`: Coordinated Universal Time
- `America/New_York`: US Eastern Time
- `Europe/London`: UK Time
- `Asia/Tokyo`: Japan Standard Time
- `Local`: System local timezone

### Error Handling
All errors return MCP-compliant responses with:
- `IsError: true`
- Descriptive error messages
- Proper HTTP status codes
- Structured error information
