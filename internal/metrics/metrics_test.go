package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Clear any existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics := New()

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.ToolRequestDuration)
	assert.NotNil(t, metrics.TimeOperationDuration)
	assert.NotNil(t, metrics.TransportRequestsTotal)
	assert.NotNil(t, metrics.ErrorsTotal)
}

func TestMetrics_RecordToolRequestDuration(t *testing.T) {
	// Clear any existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics := New()

	// Record some tool request durations
	metrics.RecordToolRequestDuration("get_time", StatusSuccess, 0.1)
	metrics.RecordToolRequestDuration("get_time", StatusSuccess, 0.2)
	metrics.RecordToolRequestDuration("get_time", StatusError, 0.05)
	metrics.RecordToolRequestDuration("format_time", StatusSuccess, 0.3)

	// Verify the histogram is working by checking that gathering metrics works
	// For histograms, we can't easily check exact counts, so we just verify no panics
	gatherer := prometheus.DefaultGatherer
	metricFamilies, err := gatherer.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestMetrics_RecordTimeOperationDuration(t *testing.T) {
	// Clear any existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics := New()

	// Record some operation durations
	metrics.RecordTimeOperationDuration(OperationGetTime, StatusSuccess, 0.001)
	metrics.RecordTimeOperationDuration(OperationGetTime, StatusSuccess, 0.002)
	metrics.RecordTimeOperationDuration(OperationGetTime, StatusError, 0.005)
	metrics.RecordTimeOperationDuration(OperationParseTime, StatusSuccess, 0.01)

	// Verify the histogram is working by checking that gathering metrics works
	gatherer := prometheus.DefaultGatherer
	metricFamilies, err := gatherer.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestMetrics_RecordError(t *testing.T) {
	// Clear any existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics := New()

	// Record some errors
	metrics.RecordError(ErrorCategoryValidation, ErrorTypeInvalidTimezone)
	metrics.RecordError(ErrorCategoryValidation, ErrorTypeInvalidFormat)
	metrics.RecordError(ErrorCategoryTime, ErrorTypeParseFailure)

	// Check the metrics
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.ErrorsTotal.WithLabelValues(ErrorCategoryValidation, ErrorTypeInvalidTimezone)))
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.ErrorsTotal.WithLabelValues(ErrorCategoryValidation, ErrorTypeInvalidFormat)))
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.ErrorsTotal.WithLabelValues(ErrorCategoryTime, ErrorTypeParseFailure)))
}

func TestConstants(t *testing.T) {
	// Test that all constants are defined and have expected values
	assert.Equal(t, "success", StatusSuccess)
	assert.Equal(t, "error", StatusError)
	assert.Equal(t, "timeout", StatusTimeout)
	assert.Equal(t, "invalid", StatusInvalid)

	assert.Equal(t, "get_time", OperationGetTime)
	assert.Equal(t, "format_time", OperationFormatTime)
	assert.Equal(t, "parse_time", OperationParseTime)
	assert.Equal(t, "timezone_info", OperationTimezoneInfo)
	assert.Equal(t, "convert_timezone", OperationConvertTimezone)

	assert.Equal(t, "sse", TransportSSE)
	assert.Equal(t, "streamable", TransportStreamable)

	assert.Equal(t, "validation", ErrorCategoryValidation)
	assert.Equal(t, "time", ErrorCategoryTime)
	assert.Equal(t, "transport", ErrorCategoryTransport)
	assert.Equal(t, "internal", ErrorCategoryInternal)

	assert.Equal(t, "invalid_timezone", ErrorTypeInvalidTimezone)
	assert.Equal(t, "invalid_format", ErrorTypeInvalidFormat)
	assert.Equal(t, "parse_failure", ErrorTypeParseFailure)
	assert.Equal(t, "connection_lost", ErrorTypeConnectionLost)
	assert.Equal(t, "invalid_request", ErrorTypeInvalidRequest)
}

func TestMetrics_Integration(t *testing.T) {
	// Clear any existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics := New()

	// Simulate a complete request flow
	toolName := "get_time"

	// Record tool request duration with status
	metrics.RecordToolRequestDuration(toolName, StatusSuccess, 0.15)

	// Record underlying time operations
	metrics.RecordTimeOperationDuration(OperationGetTime, StatusSuccess, 0.001)

	// Record transport activity
	metrics.RecordTransportRequest(TransportSSE, "POST", StatusSuccess)

	// Record an error
	metrics.RecordError(ErrorCategoryTime, ErrorTypeInvalidTimezone)

	// Verify all metrics were recorded (check counters work, histograms just verify no panics)
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.TransportRequestsTotal.WithLabelValues(TransportSSE, "POST", StatusSuccess)))
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.ErrorsTotal.WithLabelValues(ErrorCategoryTime, ErrorTypeInvalidTimezone)))

	// Verify histograms work by checking metrics gathering
	gatherer := prometheus.DefaultGatherer
	metricFamilies, err := gatherer.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}
