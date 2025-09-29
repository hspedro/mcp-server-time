# Technical Context - MCP Time Server

## Technology Stack

### Language & Runtime
- **Go 1.23+**: Modern Go with generics and enhanced performance
- **MCP Go SDK v0.8.0**: Official Model Context Protocol SDK for Go
- **Standard Library**: Leverages Go's excellent built-in capabilities for HTTP, JSON, context, time

### Core Libraries
- **[MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)**: Official MCP implementation for servers and clients
- **[Viper](https://github.com/spf13/viper)**: Configuration management with environment variable support
- **[Zap](https://github.com/uber-go/zap)**: High-performance structured logging
- **[Prometheus](https://github.com/prometheus/client_golang)**: Metrics collection and exposure (excluding Go runtime metrics)
- **net/http**: Standard HTTP server and client (no external web framework)
- **time**: Go's comprehensive time package for all time operations

### Library Philosophy
- **Official MCP SDK**: Use the official Go SDK for maximum compatibility
- **Well-established**: Prefer battle-tested libraries with strong community adoption
- **Minimal dependencies**: Favor composition over heavyweight frameworks
- **Standard library first**: Use Go's excellent standard library when sufficient
- **Observability focused**: Built-in metrics, logging from day one

### External Dependencies
- **None**: Self-contained time operations using Go's standard time package
- **MCP Proxy**: Communicates via SSE and Streamable transports

## Deployment & Development

### Target Environment
- **Container-first**: Docker images built for multi-architecture support (ARM64/AMD64)
- **Cloud-native**: Designed for Kubernetes with proper health checks and observability
- **Standalone**: Can run independently or as part of MCP proxy ecosystem

### Configuration Strategy
- **Hierarchical Config**: YAML files with environment variable overrides
- **Kubernetes-ready**: All configuration externalizable via ConfigMaps and Secrets
- **Local Development**: Standalone operation with docker-compose integration

### Project Structure
This is a standalone MCP server following Go best practices:
```
mcp-server-time/
├── cmd/main.go              # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── time/               # Core time service
│   ├── handlers/           # MCP tool handlers
│   ├── features/           # MCP tool and prompt definitions
│   ├── metrics/            # Prometheus metrics
│   └── transport/          # SSE and Streamable transports
├── docs/                   # Documentation
├── .github/workflows/      # GitHub Actions CI/CD
└── config.yaml            # Default configuration
```

## Architecture Overview

This service follows **Clean Architecture** principles with clear separation of concerns:

```
internal/
├── time/                   # Core business logic
│   ├── service.go         # Time operations interface & implementation
│   ├── service_test.go    # Time service tests
│   ├── types.go           # Time types and constants
│   └── mocks/             # Generated mocks
├── handlers/              # MCP tool handlers
│   ├── get_time.go        # Current time tool
│   ├── format_time.go     # Time formatting tool
│   ├── parse_time.go      # Time parsing tool
│   └── timezone_info.go   # Timezone information tool
├── transport/             # Communication layer
│   ├── sse.go            # Server-Sent Events transport
│   └── streamable.go     # Streamable HTTP transport
└── config/
    └── config.go          # Configuration management
```

## Key Design Principles

### 1. MCP Protocol Compliance
- Implements official MCP Go SDK patterns
- Supports both SSE and Streamable transports
- Proper tool and prompt definitions
- Standard MCP error handling

### 2. Interface-Based Design
- Core time operations defined as interfaces
- Enables easy testing with mocks
- Clear boundaries between components
- Dependency injection throughout

### 3. Testability
- Mock interfaces for unit testing
- Inline test files alongside code
- No external dependencies for core logic
- Comprehensive test coverage

## MCP Tools Implementation

### Tool Categories
The server provides four core MCP tools based on common time-related needs:

1. **get_time**: Get current time with optional timezone and format
2. **format_time**: Format a timestamp using custom formats
3. **parse_time**: Parse time strings in various formats
4. **timezone_info**: Get timezone information and perform conversions

### Time Service Interface
```go
// TimeService defines core time operations
type TimeService interface {
    GetCurrentTime(timezone string) (time.Time, error)
    FormatTime(t time.Time, format string) (string, error)
    ParseTime(timeStr, format string) (time.Time, error)
    GetTimezoneInfo(timezone string) (*TimezoneInfo, error)
    ConvertTimezone(t time.Time, fromTZ, toTZ string) (time.Time, error)
}
```

### Transport Layer
```go
// Both SSE and Streamable transports for MCP Proxy compatibility
type Transport interface {
    Handler() http.Handler
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

## Technical Implementation Details

### Configuration Management
```go
// Uses Viper for hierarchical configuration
// Supports: config files, environment variables, defaults
type Config struct {
    Server  ServerConfig  `mapstructure:"server"`
    Time    TimeConfig    `mapstructure:"time"`
    Logging LogConfig     `mapstructure:"logging"`
    Metrics MetricsConfig `mapstructure:"metrics"`
}
```

**Configuration Hierarchy** (highest priority first):
1. Environment variables (e.g., `TIME_SERVER_PORT`)
2. YAML configuration file (`config.yaml`)
3. Default values in code

**Example YAML Config**:
```yaml
server:
  name: "time-mcp-server"
  version: "1.0.0"
  host: "localhost"
  port: 8080

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

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"
```

### Logging & Observability
```go
// Structured logging with Zap
logger.Info("Processing time tool request",
    zap.String("tool", "get_time"),
    zap.String("timezone", timezone),
    zap.String("format", format))

// Prometheus metrics with labels
timeToolsTotal.WithLabelValues("get_time", "success").Inc()
timeToolDuration.WithLabelValues("get_time").Observe(duration.Seconds())
```

### HTTP Server Architecture
```go
// Standard net/http with MCP transports
mux := http.NewServeMux()
mux.Handle("/sse", sseTransport.Handler())
mux.Handle("/streamable", streamableTransport.Handler())
mux.Handle("/mcp", streamableTransport.Handler())
mux.HandleFunc("/health", handlers.HealthCheck)
mux.Handle("/metrics", promhttp.Handler())

server := &http.Server{
    Handler: mux,
    Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
}
```

### Time Operations
```go
// Core time service implementation
type timeService struct {
    defaultTimezone string
    supportedFormats []string
    logger          *zap.Logger
}

// Get current time with timezone support
func (s *timeService) GetCurrentTime(timezone string) (time.Time, error) {
    if timezone == "" {
        timezone = s.defaultTimezone
    }

    loc, err := time.LoadLocation(timezone)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
    }

    return time.Now().In(loc), nil
}
```

### Testing Strategy

### Unit Tests
- Time service: Test with various timezones and formats
- Handlers: Test with mocked time service using testify/mock
- Transports: Test HTTP endpoints with httptest
- Config: Test configuration loading and validation

### Integration Tests
- Full MCP server stack testing
- Transport protocol compliance
- Tool execution end-to-end

### Testing Libraries
- **[testify](https://github.com/stretchr/testify)**: Assertions and mocking
- **[mockgen](https://github.com/golang/mock)**: Interface mocking
- **httptest**: Standard library HTTP testing utilities

## Development Workflow

### Makefile-Driven Development
All development tasks are orchestrated through a comprehensive `Makefile`:

```bash
# Core development commands
make tidy          # Update dependencies
make build         # Build binary
make run           # Run locally with config
make test          # Run all tests
make fmt           # Format code
make lint          # Run linters

# Code generation
make mocks         # Generate test mocks
make tools         # Install development tools

# Docker operations
make docker-build  # Build container image
make docker-run    # Run in container

# Verification
make verify        # Full verification (lint + test + build)
```

### GitHub Actions Pipeline
Multi-architecture build pipeline:

**Test Stage**:
```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.23'
    - run: make verify
```

**Build Stage**:
- ARM64 build on self-hosted ARM runner
- AMD64 build on GitHub-hosted runner
- Multi-arch manifest creation and publishing

### Local Development
Standalone development with optional integration:

```bash
# Run standalone
make run

# Run with docker-compose (integrated with troubleshooter stack)
docker-compose up mcp-server-time

# Test with MCP client
curl http://localhost:8080/health
curl http://localhost:8080/sse  # SSE transport
curl http://localhost:8080/streamable  # Streamable transport
```

**Environment Configuration**:
```bash
# Create .env file (optional)
TIME_SERVER_PORT=8080
TIME_DEFAULT_TIMEZONE=America/New_York
TIME_LOGGING_LEVEL=debug
```

## MCP Integration

### Tool Definitions
Each tool is defined following MCP standards:

```go
var GetTimeTool = &mcp.Tool{
    Name:        "get_time",
    Description: "Get current time with optional timezone and format",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "timezone": map[string]any{
                "type":        "string",
                "description": "Timezone name (IANA format, e.g., 'America/New_York')",
                "default":     "UTC",
            },
            "format": map[string]any{
                "type":        "string",
                "description": "Output format (RFC3339, Unix, UnixMilli, etc.)",
                "default":     "RFC3339",
            },
        },
    },
}
```

### Transport Compatibility
- **SSE Transport**: Real-time streaming for MCP Proxy
- **Streamable Transport**: Stateless HTTP requests
- **Health Checks**: Kubernetes-ready endpoints

## Benefits of This Architecture

1. **MCP Compliance**: Full compatibility with MCP ecosystem
2. **Simplicity**: Focused on time operations without external dependencies
3. **Performance**: Efficient Go implementation with minimal overhead
4. **Testability**: Interface-based design with comprehensive mocking
5. **Observability**: Built-in metrics and structured logging
6. **Portability**: Runs anywhere Docker containers are supported

## Common Patterns

### Error Handling
```go
// Domain-specific errors
var (
    ErrInvalidTimezone = errors.New("invalid timezone")
    ErrInvalidFormat   = errors.New("invalid time format")
    ErrParseFailure    = errors.New("failed to parse time string")
)

// Wrapped errors with context
return fmt.Errorf("failed to get current time for timezone %s: %w", timezone, err)
```

### Dependency Injection
```go
// Constructor pattern
func NewTimeService(config TimeConfig, logger *zap.Logger) TimeService {
    return &timeService{
        defaultTimezone:  config.DefaultTimezone,
        supportedFormats: config.SupportedFormats,
        logger:          logger,
    }
}
```

### Configuration
```go
// Environment variable precedence
viper.SetEnvPrefix("TIME")
viper.AutomaticEnv()
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
```

## Related Documentation

- [agent_workflow.md](./agent_workflow.md) - Development workflow for AI agents
- [CLAUDE.md](./CLAUDE.md) - Quick reference guide
- [README.md](../README.md) - Project setup and usage
