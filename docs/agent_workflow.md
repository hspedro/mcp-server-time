# Agent Workflow Guide - MCP Time Server

## Overview

This document defines the standard workflow that AI agents (Claude, Cursor, etc.) should follow when working on the MCP Time Server. This ensures consistent, high-quality changes that align with MCP patterns and Go best practices.

## Pre-Work: Context Loading

### 1. **Always Read Core Documentation**
Before analyzing any user request, agents must load:

- **[tech_context.md](./tech_context.md)** - MCP patterns, Go architecture, and design decisions
- **[CLAUDE.md](./CLAUDE.md)** - Project overview and quick reference
- **[README.md](../README.md)** - Project setup and usage information

### 2. **Load Relevant MCP Documentation**
For MCP-specific changes, reference:
- **[MCP Go SDK Documentation](https://github.com/modelcontextprotocol/go-sdk)** - Official SDK patterns
- **[MCP Protocol Specification](https://modelcontextprotocol.io)** - Protocol compliance requirements

## Standard Workflow Process

### Phase 1: Analysis & Understanding

1. **Analyze User Request**
   - Understand the functional requirements
   - Identify affected components (time service, handlers, transports)
   - Consider impact on MCP protocol compliance
   - Check for architectural implications

2. **Review Current Implementation**
   - Read relevant source files to understand current state
   - Check existing tests for coverage and patterns
   - Identify potential breaking changes to MCP tools
   - Consider backward compatibility

3. **Validate Against MCP Standards**
   - Ensure changes align with MCP protocol requirements
   - Verify tool definitions follow MCP schema standards
   - Check transport compatibility (SSE and Streamable)
   - Consider impact on MCP client integrations

### Phase 2: Planning & Design

**CRITICAL: Always present a comprehensive plan to the user for validation BEFORE making any changes.**

The plan must include:

#### A. **Code Changes**
```
1. Time Service Changes:
   - Core time operations in internal/time/
   - Interface modifications
   - Business logic updates

2. Handler Changes:
   - MCP tool handler implementations
   - Input validation and schema updates
   - Error handling improvements

3. Transport Changes:
   - SSE transport modifications
   - Streamable transport updates
   - Protocol compliance adjustments
```

#### B. **Testing Strategy**
```
1. Unit Tests:
   - Time service tests with mocked dependencies
   - Handler tests with mocked time service
   - Transport tests with HTTP test utilities

2. Integration Tests:
   - End-to-end MCP tool execution
   - Transport protocol compliance
   - Health check endpoint testing

3. Test Data:
   - Mock time data creation
   - Timezone test cases
   - Format validation scenarios
```

#### C. **Development Workflow**
```
1. Code Implementation Order:
   - Time service interfaces and types first
   - Service implementations
   - Handler updates
   - Transport modifications
   - Feature definitions

2. Makefile Commands:
   - make fmt (code formatting)
   - make lint (linting checks)
   - make test (run all tests)
   - make mocks (generate mocks)
   - make build (build verification)
```

#### D. **Configuration Changes**
```
1. YAML Configuration:
   - New time service configurations
   - Environment variable mappings
   - Default value updates

2. Container Deployment:
   - Dockerfile updates
   - Health check adjustments
   - Port exposure changes

3. Integration Setup:
   - Docker Compose updates
   - MCP proxy configuration
   - Service discovery changes
```

#### E. **Documentation Updates**
```
1. Technical Documentation:
   - Update relevant docs in docs/
   - Add new MCP tool documentation
   - Update CLAUDE.md if patterns change

2. Code Documentation:
   - Function and interface comments
   - Package documentation
   - MCP tool schema documentation

3. README Updates:
   - Configuration changes
   - New environment variables
   - Usage examples
```

#### F. **MCP Compliance Considerations**
```
1. Protocol Compliance:
   - Tool schema validation
   - Transport behavior verification
   - Error response formatting

2. Client Compatibility:
   - MCP client integration testing
   - Proxy compatibility verification
   - Protocol version considerations

3. Performance Impact:
   - Time operation efficiency
   - Memory usage optimization
   - Concurrent request handling
```

### Phase 3: User Validation

Present the complete plan with:

1. **Clear Summary**: High-level description of changes
2. **Implementation Steps**: Ordered list of changes to make
3. **MCP Impact**: How changes affect MCP protocol compliance
4. **Testing Approach**: How changes will be validated
5. **Integration Impact**: What changes during deployment

**Wait for user approval before proceeding.**

### Phase 4: Implementation

Once approved, follow this order:

1. **Time Service First**: Implement core time operations and interfaces
2. **Handlers**: Update MCP tool handlers
3. **Transports**: Modify SSE and Streamable transports if needed
4. **Features**: Update tool and prompt definitions
5. **Tests**: Write comprehensive tests with mocks
6. **Configuration**: Update config files and documentation
7. **Validation**: Run make commands to verify everything works

### Phase 5: Verification

After implementation:

1. **Code Quality Checks**
   ```bash
   make fmt lint          # Format and lint code
   make test              # Run all tests
   make mocks             # Regenerate mocks
   make build             # Verify build
   ```

2. **MCP Compliance Verification**
   ```bash
   make run               # Start server
   # Test MCP tools manually
   curl http://localhost:8080/health
   curl http://localhost:8080/sse
   curl http://localhost:8080/streamable
   ```

3. **Integration Verification**
   ```bash
   docker-compose up mcp-server-time  # Test containerized
   # Test with MCP proxy if available
   ```

4. **Documentation Review**
   - Verify all docs are updated
   - Check for broken links
   - Validate code examples and schemas

## Architecture Compliance Checklist

### ✅ **Time Service Layer**
- [ ] Interface-based design
- [ ] No external dependencies for core logic
- [ ] Proper error handling with context
- [ ] Timezone validation and support
- [ ] Format validation and conversion

### ✅ **Handler Layer**
- [ ] MCP tool schema compliance
- [ ] Input validation and sanitization
- [ ] Proper error response formatting
- [ ] Structured logging with context
- [ ] Metrics collection
- [ ] No business logic in handlers

### ✅ **Transport Layer**
- [ ] SSE transport compliance
- [ ] Streamable transport compliance
- [ ] Health check endpoints
- [ ] Proper content-type headers
- [ ] Error response formatting

### ✅ **Testing**
- [ ] Unit tests for all time operations
- [ ] Handler tests with mocked services
- [ ] Transport tests with HTTP utilities
- [ ] Mock generation for interfaces
- [ ] Test coverage adequate

### ✅ **Configuration**
- [ ] Environment variable overrides
- [ ] Container-friendly configuration
- [ ] Sensitive data handling
- [ ] Default values provided
- [ ] Documentation updated

## Common Patterns to Follow

### 1. **MCP Tool Definition**
```go
var GetTimeTool = &mcp.Tool{
    Name:        "get_time",
    Description: "Get current time with optional timezone and format",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "timezone": map[string]any{
                "type":        "string",
                "description": "Timezone name (IANA format)",
                "default":     "UTC",
            },
        },
    },
}
```

### 2. **Error Handling**
```go
// Domain errors
var ErrInvalidTimezone = errors.New("invalid timezone")

// Wrapped errors with context
return fmt.Errorf("failed to get current time for timezone %s: %w", timezone, err)

// MCP error responses
return &mcp.CallToolResult{
    IsError: true,
    Content: []mcp.Content{
        &mcp.TextContent{
            Type: "text",
            Text: fmt.Sprintf("Invalid timezone: %s", timezone),
        },
    },
}
```

### 3. **Interface Design**
```go
// TimeService defines core operations
type TimeService interface {
    GetCurrentTime(timezone string) (time.Time, error)
    FormatTime(t time.Time, format string) (string, error)
    ParseTime(timeStr, format string) (time.Time, error)
    GetTimezoneInfo(timezone string) (*TimezoneInfo, error)
}

// Constructor with dependency injection
func NewTimeService(config TimeConfig, logger *zap.Logger) TimeService {
    return &timeService{
        defaultTimezone: config.DefaultTimezone,
        logger:         logger,
    }
}
```

### 4. **Configuration Management**
```go
// Viper with environment override
type Config struct {
    Server ServerConfig `mapstructure:"server"`
    Time   TimeConfig   `mapstructure:"time"`
}

// Environment variable mapping
viper.SetEnvPrefix("TIME")
viper.AutomaticEnv()
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
```

### 5. **Structured Logging**
```go
// Contextual logging with Zap
logger.Info("Processing time tool request",
    zap.String("tool", "get_time"),
    zap.String("timezone", timezone),
    zap.String("format", format),
    zap.String("request_id", requestID))
```

## Quality Gates

Before considering work complete:

1. **All tests pass**: `make test`
2. **Code is formatted**: `make fmt`
3. **Linting passes**: `make lint`
4. **Mocks are current**: `make mocks`
5. **Build succeeds**: `make build`
6. **Documentation updated**: All relevant docs reflect changes
7. **Configuration validated**: YAML examples work
8. **MCP compliance verified**: Tools work with MCP clients

## Anti-Patterns to Avoid

### ❌ **MCP Protocol Violations**
- Non-standard tool schemas
- Incorrect error response formats
- Missing required fields in responses
- Transport protocol deviations

### ❌ **Architecture Violations**
- Business logic in handlers
- External dependencies in core service
- Concrete types instead of interfaces
- Global state or singletons

### ❌ **Time Handling Issues**
- Ignoring timezone requirements
- Incorrect format handling
- Missing validation for time inputs
- Poor error messages for time operations

### ❌ **Code Quality Issues**
- Missing error handling
- Unstructured logging
- No input validation
- Missing tests for time operations

### ❌ **Configuration Problems**
- Hardcoded timezone values
- No environment variable support
- Missing documentation
- No default values

### ❌ **Development Process Issues**
- Making changes without planning
- Skipping user validation
- Not updating documentation
- Ignoring failing tests

## MCP-Specific Guidelines

### Tool Implementation
1. **Always validate input schemas** according to MCP standards
2. **Return proper content types** (text, image, etc.)
3. **Handle errors gracefully** with user-friendly messages
4. **Document tool capabilities** clearly in schemas

### Transport Implementation
1. **Support both SSE and Streamable** for maximum compatibility
2. **Implement proper health checks** for monitoring
3. **Handle concurrent requests** efficiently
4. **Maintain protocol compliance** in all responses

### Testing MCP Tools
1. **Test tool schemas** for validation
2. **Test error scenarios** and responses
3. **Test with real MCP clients** when possible
4. **Verify transport behavior** under load

## Getting Help

If unclear about any aspect:

1. **Check MCP Go SDK examples** in the official repository
2. **Review existing tool implementations** in the codebase
3. **Reference MCP protocol documentation** for standards
4. **Ask for clarification** before proceeding
5. **Follow the planning process** - when in doubt, plan more

Remember: **Always plan first, validate with user, then implement.**
