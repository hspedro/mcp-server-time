package time

import (
	"time"
)

// TimezoneInfo contains information about a timezone
type TimezoneInfo struct {
	Name          string             `json:"name"`
	Abbreviation  string             `json:"abbreviation"`
	Offset        string             `json:"offset"`
	OffsetSeconds int                `json:"offset_seconds"`
	IsDST         bool               `json:"is_dst"`
	DST           *DSTInfo           `json:"dst,omitempty"`
	DSTTransition *DSTTransitionInfo `json:"dst_transition,omitempty"` // Keep for backward compatibility
}

// DSTInfo contains DST period information
type DSTInfo struct {
	Start  time.Time     `json:"start"`
	End    time.Time     `json:"end"`
	Saving time.Duration `json:"saving"`
}

// DSTTransitionInfo contains information about DST transitions
type DSTTransitionInfo struct {
	NextTransition time.Time `json:"next_transition"`
	TransitionType string    `json:"transition_type"` // "enter_dst" or "exit_dst"
	OffsetChange   int       `json:"offset_change"`   // seconds
}

// FormatType represents supported time format types
type FormatType string

const (
	FormatRFC3339     FormatType = "RFC3339"
	FormatRFC3339Nano FormatType = "RFC3339Nano"
	FormatUnix        FormatType = "Unix"
	FormatUnixMilli   FormatType = "UnixMilli"
	FormatUnixMicro   FormatType = "UnixMicro"
	FormatUnixNano    FormatType = "UnixNano"
	FormatLayout      FormatType = "Layout"
)

// IsValidFormat checks if a format type is supported
func IsValidFormat(format string) bool {
	switch FormatType(format) {
	case FormatRFC3339, FormatRFC3339Nano, FormatUnix, FormatUnixMilli, FormatUnixMicro, FormatUnixNano, FormatLayout:
		return true
	default:
		return false
	}
}

// GetFormatLayout returns the Go time layout for a given format type
func GetFormatLayout(format FormatType) string {
	switch format {
	case FormatRFC3339:
		return time.RFC3339
	case FormatRFC3339Nano:
		return time.RFC3339Nano
	default:
		return time.RFC3339 // default fallback
	}
}

// ParseTimeInput represents input for parsing time strings
type ParseTimeInput struct {
	TimeString string `json:"time_string"`
	Format     string `json:"format,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

// FormatTimeInput represents input for formatting time
type FormatTimeInput struct {
	Timestamp interface{} `json:"timestamp"` // can be string, int, or time.Time
	Format    string      `json:"format"`
	Timezone  string      `json:"timezone,omitempty"`
}

// GetTimeInput represents input for getting current time
type GetTimeInput struct {
	Timezone string `json:"timezone,omitempty"`
	Format   string `json:"format,omitempty"`
}

// TimezoneInfoInput represents input for timezone information
type TimezoneInfoInput struct {
	Timezone      string    `json:"timezone"`
	ReferenceTime time.Time `json:"reference_time,omitempty"`
}

// Result types for MCP tool responses

// GetTimeResult represents the result of getting current time
type GetTimeResult struct {
	FormattedTime string `json:"formatted_time"`
	Timezone      string `json:"timezone"`
	Format        string `json:"format"`
	UnixTimestamp int64  `json:"unix_timestamp"`
}

// FormatTimeResult represents the result of formatting time
type FormatTimeResult struct {
	FormattedTime string `json:"formatted_time"`
	Timezone      string `json:"timezone"`
	Format        string `json:"format"`
	UnixTimestamp int64  `json:"unix_timestamp"`
}

// ParseTimeResult represents the result of parsing time
type ParseTimeResult struct {
	UnixTimestamp int64  `json:"unix_timestamp"`
	RFC3339       string `json:"rfc3339"`
	Timezone      string `json:"timezone"`
	IsDST         bool   `json:"is_dst"`
}
