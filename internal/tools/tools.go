package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/hspedro/mcp-server-time/internal/metrics"
	timeservice "github.com/hspedro/mcp-server-time/internal/time"
)

// RegisterTimeTools registers all time-related tools with the MCP server
func RegisterTimeTools(server *mcp.Server, timeService timeservice.TimeService, metrics *metrics.Metrics, logger *zap.Logger) {
	registerGetTimeTool(server, timeService, metrics, logger)
	registerFormatTimeTool(server, timeService, metrics, logger)
	registerParseTimeTool(server, timeService, metrics, logger)
	registerTimezoneInfoTool(server, timeService, metrics, logger)
}

// registerGetTimeTool registers the get_time tool
func registerGetTimeTool(server *mcp.Server, timeService timeservice.TimeService, metrics *metrics.Metrics, logger *zap.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_time",
		Description: "Get the current time in a specified timezone and format",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input timeservice.GetTimeInput) (*mcp.CallToolResult, timeservice.GetTimeResult, error) {
		startTime := time.Now()

		result, err := timeService.GetCurrentTime(input)
		if err != nil {
			recordError(metrics, "get_time", "get_current_time", startTime, logger, err)
			return nil, timeservice.GetTimeResult{}, err
		}

		recordSuccess(metrics, "get_time", "get_current_time", startTime)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Current time: %s\nTimezone: %s\nFormat: %s",
						result.FormattedTime, result.Timezone, result.Format),
				},
			},
		}, result, nil
	})
}

// registerFormatTimeTool registers the format_time tool
func registerFormatTimeTool(server *mcp.Server, timeService timeservice.TimeService, metrics *metrics.Metrics, logger *zap.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "format_time",
		Description: "Format a timestamp into a specified format and timezone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input timeservice.FormatTimeInput) (*mcp.CallToolResult, timeservice.FormatTimeResult, error) {
		startTime := time.Now()

		result, err := timeService.FormatTime(input)
		if err != nil {
			recordError(metrics, "format_time", "format_time", startTime, logger, err)
			return nil, timeservice.FormatTimeResult{}, err
		}

		recordSuccess(metrics, "format_time", "format_time", startTime)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Formatted time: %s\nOriginal: %s\nTimezone: %s\nFormat: %s",
						result.FormattedTime, input.Timestamp, result.Timezone, result.Format),
				},
			},
		}, result, nil
	})
}

// registerParseTimeTool registers the parse_time tool
func registerParseTimeTool(server *mcp.Server, timeService timeservice.TimeService, metrics *metrics.Metrics, logger *zap.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "parse_time",
		Description: "Parse a time string and return timestamp information",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input timeservice.ParseTimeInput) (*mcp.CallToolResult, timeservice.ParseTimeResult, error) {
		startTime := time.Now()

		result, err := timeService.ParseTime(input)
		if err != nil {
			recordError(metrics, "parse_time", "parse_time", startTime, logger, err)
			return nil, timeservice.ParseTimeResult{}, err
		}

		recordSuccess(metrics, "parse_time", "parse_time", startTime)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Parsed time:\n- Unix timestamp: %d\n- RFC3339: %s\n- Timezone: %s\n- Is DST: %t",
						result.UnixTimestamp, result.RFC3339, result.Timezone, result.IsDST),
				},
			},
		}, result, nil
	})
}

// registerTimezoneInfoTool registers the timezone_info tool
func registerTimezoneInfoTool(server *mcp.Server, timeService timeservice.TimeService, metrics *metrics.Metrics, logger *zap.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "timezone_info",
		Description: "Get detailed information about a timezone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input timeservice.TimezoneInfoInput) (*mcp.CallToolResult, timeservice.TimezoneInfo, error) {
		startTime := time.Now()

		result, err := timeService.GetTimezoneInfo(input)
		if err != nil {
			recordError(metrics, "timezone_info", "get_timezone_info", startTime, logger, err)
			return nil, timeservice.TimezoneInfo{}, err
		}

		recordSuccess(metrics, "timezone_info", "get_timezone_info", startTime)

		dstInfo := "No DST transitions"
		if result.DST != nil {
			dstInfo = fmt.Sprintf("DST: %s to %s (saving %s)",
				result.DST.Start.Format(time.RFC3339),
				result.DST.End.Format(time.RFC3339),
				result.DST.Saving.String())
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Timezone: %s\nAbbreviation: %s\nOffset: %s\nCurrent DST: %t\n%s",
						result.Name, result.Abbreviation, result.Offset, result.IsDST, dstInfo),
				},
			},
		}, result, nil
	})
}

// recordError is a helper function to record error metrics and log
func recordError(metrics *metrics.Metrics, toolName, operationName string, startTime time.Time, logger *zap.Logger, err error) {
	duration := time.Since(startTime).Seconds()
	metrics.RecordToolRequestDuration(toolName, "error", duration)
	metrics.RecordTimeOperationDuration(operationName, "error", duration)
	logger.Error(fmt.Sprintf("%s failed", toolName), zap.Error(err))
}

// recordSuccess is a helper function to record success metrics
func recordSuccess(metrics *metrics.Metrics, toolName, operationName string, startTime time.Time) {
	duration := time.Since(startTime).Seconds()
	metrics.RecordToolRequestDuration(toolName, "success", duration)
	metrics.RecordTimeOperationDuration(operationName, "success", duration)
}
