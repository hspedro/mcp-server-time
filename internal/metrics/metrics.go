package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the Prometheus metrics for the MCP Time Server
type Metrics struct {
	// MCP tool request metrics
	ToolRequestDuration prometheus.HistogramVec

	// Time operation metrics
	TimeOperationDuration prometheus.HistogramVec

	// Transport metrics
	TransportRequestsTotal prometheus.CounterVec

	// Error metrics
	ErrorsTotal prometheus.CounterVec
}

// New creates a new Metrics instance with all metrics registered
func New() *Metrics {
	return &Metrics{
		ToolRequestDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mcp_time_tool_request_duration_seconds",
				Help:    "Duration of MCP tool requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"tool", "status"},
		),

		TimeOperationDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mcp_time_operation_duration_seconds",
				Help:    "Duration of time operations in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"operation", "status"},
		),

		TransportRequestsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcp_time_transport_requests_total",
				Help: "Total number of transport requests",
			},
			[]string{"transport", "method", "status"},
		),

		ErrorsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcp_time_errors_total",
				Help: "Total number of errors by category",
			},
			[]string{"category", "error_type"},
		),
	}
}

// RecordToolRequestDuration records the duration of a tool request
func (m *Metrics) RecordToolRequestDuration(tool, status string, duration float64) {
	m.ToolRequestDuration.WithLabelValues(tool, status).Observe(duration)
}

// RecordTimeOperationDuration records the duration of a time operation
func (m *Metrics) RecordTimeOperationDuration(operation, status string, duration float64) {
	m.TimeOperationDuration.WithLabelValues(operation, status).Observe(duration)
}

// RecordTransportRequest records a transport request
func (m *Metrics) RecordTransportRequest(transport, method, status string) {
	m.TransportRequestsTotal.WithLabelValues(transport, method, status).Inc()
}

// RecordError records an error by category and type
func (m *Metrics) RecordError(category, errorType string) {
	m.ErrorsTotal.WithLabelValues(category, errorType).Inc()
}

// Status constants for metrics
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusTimeout = "timeout"
	StatusInvalid = "invalid"
)

// Tool operation constants
const (
	OperationGetTime         = "get_time"
	OperationFormatTime      = "format_time"
	OperationParseTime       = "parse_time"
	OperationTimezoneInfo    = "timezone_info"
	OperationConvertTimezone = "convert_timezone"
)

// Transport constants
const (
	TransportSSE        = "sse"
	TransportStreamable = "streamable"
)

// Error category constants
const (
	ErrorCategoryValidation = "validation"
	ErrorCategoryTime       = "time"
	ErrorCategoryTransport  = "transport"
	ErrorCategoryInternal   = "internal"
)

// Error type constants
const (
	ErrorTypeInvalidTimezone = "invalid_timezone"
	ErrorTypeInvalidFormat   = "invalid_format"
	ErrorTypeParseFailure    = "parse_failure"
	ErrorTypeConnectionLost  = "connection_lost"
	ErrorTypeInvalidRequest  = "invalid_request"
)
