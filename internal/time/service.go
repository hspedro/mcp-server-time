package time

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -source=service.go -destination=mocks/service_mock.go

// TimeService defines the interface for time operations
type TimeService interface {
	// GetCurrentTime returns the current time in the specified timezone and format
	GetCurrentTime(input GetTimeInput) (GetTimeResult, error)

	// FormatTime formats a timestamp using the specified format and timezone
	FormatTime(input FormatTimeInput) (FormatTimeResult, error)

	// ParseTime parses a time string and returns timestamp information
	ParseTime(input ParseTimeInput) (ParseTimeResult, error)

	// GetTimezoneInfo returns information about a timezone
	GetTimezoneInfo(input TimezoneInfoInput) (TimezoneInfo, error)

	// ConvertTimezone converts a time from one timezone to another (kept for internal use)
	ConvertTimezone(t time.Time, fromTZ, toTZ string) (time.Time, error)

	// IsFormatSupported checks if a format is supported
	IsFormatSupported(format string) bool

	// GetSupportedFormats returns a list of supported formats
	GetSupportedFormats() []string
}

// timeService implements the TimeService interface
type timeService struct {
	defaultTimezone  string
	defaultFormat    string
	supportedFormats []string
	logger           *zap.Logger
}

// NewTimeService creates a new time service instance
func NewTimeService(defaultTimezone, defaultFormat string, supportedFormats []string, logger *zap.Logger) TimeService {
	return &timeService{
		defaultTimezone:  defaultTimezone,
		defaultFormat:    defaultFormat,
		supportedFormats: supportedFormats,
		logger:           logger,
	}
}

// GetCurrentTime returns the current time with result information
func (s *timeService) GetCurrentTime(input GetTimeInput) (GetTimeResult, error) {
	timezone := input.Timezone
	format := input.Format

	if timezone == "" {
		timezone = s.defaultTimezone
	}
	if format == "" {
		format = s.defaultFormat
	}

	currentTime, err := s.getCurrentTimeInternal(timezone)
	if err != nil {
		return GetTimeResult{}, err
	}

	formatted, err := s.formatTimeInternal(currentTime, format)
	if err != nil {
		return GetTimeResult{}, err
	}

	return GetTimeResult{
		FormattedTime: formatted,
		Timezone:      timezone,
		Format:        format,
		UnixTimestamp: currentTime.Unix(),
	}, nil
}

// getCurrentTimeInternal returns the current time in the specified timezone (internal method)
func (s *timeService) getCurrentTimeInternal(timezone string) (time.Time, error) {
	if timezone == "" {
		timezone = s.defaultTimezone
	}

	s.logger.Debug("Getting current time",
		zap.String("timezone", timezone),
		zap.String("default_timezone", s.defaultTimezone))

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		s.logger.Error("Failed to load timezone location",
			zap.String("timezone", timezone),
			zap.Error(err))
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	currentTime := time.Now().In(loc)
	s.logger.Debug("Successfully retrieved current time",
		zap.String("timezone", timezone),
		zap.Time("time", currentTime))

	return currentTime, nil
}

// FormatTime formats a timestamp with result information
func (s *timeService) FormatTime(input FormatTimeInput) (FormatTimeResult, error) {
	format := input.Format
	timezone := input.Timezone

	if timezone == "" {
		timezone = s.defaultTimezone
	}

	// Parse the timestamp
	var t time.Time
	var err error

	switch v := input.Timestamp.(type) {
	case string:
		// Try to parse as Unix timestamp first, then as RFC3339
		if unixTime, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil {
			t = time.Unix(unixTime, 0)
		} else {
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return FormatTimeResult{}, fmt.Errorf("failed to parse timestamp string: %w", err)
			}
		}
	case int:
		t = time.Unix(int64(v), 0)
	case int64:
		t = time.Unix(v, 0)
	case float64:
		t = time.Unix(int64(v), 0)
	case time.Time:
		t = v
	default:
		return FormatTimeResult{}, fmt.Errorf("unsupported timestamp type: %T", input.Timestamp)
	}

	// Convert to target timezone
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return FormatTimeResult{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
		}
		t = t.In(loc)
	}

	formatted, err := s.formatTimeInternal(t, format)
	if err != nil {
		return FormatTimeResult{}, err
	}

	return FormatTimeResult{
		FormattedTime: formatted,
		Timezone:      t.Location().String(),
		Format:        format,
		UnixTimestamp: t.Unix(),
	}, nil
}

// formatTimeInternal formats a time value using the specified format (internal method)
func (s *timeService) formatTimeInternal(t time.Time, format string) (string, error) {
	if format == "" {
		format = s.defaultFormat
	}

	s.logger.Debug("Formatting time",
		zap.Time("time", t),
		zap.String("format", format))

	if !s.IsFormatSupported(format) {
		return "", fmt.Errorf("unsupported format: %s (supported: %v)", format, s.supportedFormats)
	}

	var result string
	var err error

	switch FormatType(format) {
	case FormatRFC3339:
		result = t.Format(time.RFC3339)
	case FormatRFC3339Nano:
		result = t.Format(time.RFC3339Nano)
	case FormatUnix:
		result = strconv.FormatInt(t.Unix(), 10)
	case FormatUnixMilli:
		result = strconv.FormatInt(t.UnixMilli(), 10)
	case FormatUnixMicro:
		result = strconv.FormatInt(t.UnixMicro(), 10)
	case FormatUnixNano:
		result = strconv.FormatInt(t.UnixNano(), 10)
	case FormatLayout:
		// For layout format, we expect the format to be a Go time layout
		result = t.Format(format)
	default:
		// Try as a Go time layout
		result = t.Format(format)
	}

	s.logger.Debug("Successfully formatted time",
		zap.String("format", format),
		zap.String("result", result))

	return result, err
}

// ParseTime parses a time string and returns result information
func (s *timeService) ParseTime(input ParseTimeInput) (ParseTimeResult, error) {
	timeStr := input.TimeString
	format := input.Format
	timezone := input.Timezone

	if format == "" {
		format = s.defaultFormat
	}

	parsedTime, err := s.parseTimeInternal(timeStr, format)
	if err != nil {
		return ParseTimeResult{}, err
	}

	// Apply timezone if specified
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return ParseTimeResult{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
		}
		// If the parsed time has no timezone info, assume it's in the specified timezone
		if parsedTime.Location() == time.UTC {
			parsedTime = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
				parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), parsedTime.Nanosecond(), loc)
		} else {
			parsedTime = parsedTime.In(loc)
		}
	}

	return ParseTimeResult{
		UnixTimestamp: parsedTime.Unix(),
		RFC3339:       parsedTime.Format(time.RFC3339),
		Timezone:      parsedTime.Location().String(),
		IsDST:         s.isDST(parsedTime, parsedTime.Location()),
	}, nil
}

// parseTimeInternal parses a time string using the specified format (internal method)
func (s *timeService) parseTimeInternal(timeStr, format string) (time.Time, error) {
	if format == "" {
		format = s.defaultFormat
	}

	s.logger.Debug("Parsing time string",
		zap.String("time_string", timeStr),
		zap.String("format", format))

	var parsedTime time.Time
	var err error

	switch FormatType(format) {
	case FormatRFC3339:
		parsedTime, err = time.Parse(time.RFC3339, timeStr)
	case FormatRFC3339Nano:
		parsedTime, err = time.Parse(time.RFC3339Nano, timeStr)
	case FormatUnix:
		var unixTime int64
		unixTime, err = strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			parsedTime = time.Unix(unixTime, 0)
		}
	case FormatUnixMilli:
		var milliTime int64
		milliTime, err = strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			parsedTime = time.UnixMilli(milliTime)
		}
	case FormatUnixMicro:
		var microTime int64
		microTime, err = strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			parsedTime = time.UnixMicro(microTime)
		}
	case FormatUnixNano:
		var nanoTime int64
		nanoTime, err = strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			parsedTime = time.Unix(0, nanoTime)
		}
	default:
		// Try as Go time layout
		parsedTime, err = time.Parse(format, timeStr)
	}

	if err != nil {
		s.logger.Error("Failed to parse time string",
			zap.String("time_string", timeStr),
			zap.String("format", format),
			zap.Error(err))
		return time.Time{}, fmt.Errorf("failed to parse time string %s with format %s: %w", timeStr, format, err)
	}

	s.logger.Debug("Successfully parsed time string",
		zap.String("time_string", timeStr),
		zap.String("format", format),
		zap.Time("parsed_time", parsedTime))

	return parsedTime, nil
}

// GetTimezoneInfo returns information about a timezone
func (s *timeService) GetTimezoneInfo(input TimezoneInfoInput) (TimezoneInfo, error) {
	timezone := input.Timezone
	if timezone == "" {
		timezone = s.defaultTimezone
	}

	// Use provided reference time or current time
	refTime := time.Now()
	if !input.ReferenceTime.IsZero() {
		refTime = input.ReferenceTime
	}

	info, err := s.getTimezoneInfoInternal(timezone, &refTime)
	if err != nil {
		return TimezoneInfo{}, err
	}

	// Return as value instead of pointer to match interface
	return *info, nil
}

// getTimezoneInfoInternal returns information about a timezone (internal method)
func (s *timeService) getTimezoneInfoInternal(timezone string, referenceTime *time.Time) (*TimezoneInfo, error) {
	if timezone == "" {
		timezone = s.defaultTimezone
	}

	s.logger.Debug("Getting timezone info",
		zap.String("timezone", timezone))

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		s.logger.Error("Failed to load timezone location for info",
			zap.String("timezone", timezone),
			zap.Error(err))
		return nil, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// Use provided reference time or current time
	refTime := time.Now()
	if referenceTime != nil {
		refTime = *referenceTime
	}

	// Get time in the specified timezone
	timeInZone := refTime.In(loc)

	// Get timezone abbreviation and offset
	zoneName, offset := timeInZone.Zone()

	// Check if DST is active
	isDST := s.isDST(timeInZone, loc)

	// Calculate DST transition info
	dstTransition := s.getNextDSTTransition(timeInZone, loc)

	info := &TimezoneInfo{
		Name:          timezone,
		Abbreviation:  zoneName,
		Offset:        formatOffset(offset),
		OffsetSeconds: offset,
		IsDST:         isDST,
		DSTTransition: dstTransition,
	}

	s.logger.Debug("Successfully retrieved timezone info",
		zap.String("timezone", timezone),
		zap.String("abbreviation", zoneName),
		zap.Int("offset_seconds", offset),
		zap.Bool("is_dst", isDST))

	return info, nil
}

// ConvertTimezone converts a time from one timezone to another
func (s *timeService) ConvertTimezone(t time.Time, fromTZ, toTZ string) (time.Time, error) {
	s.logger.Debug("Converting timezone",
		zap.Time("time", t),
		zap.String("from_timezone", fromTZ),
		zap.String("to_timezone", toTZ))

	toLoc, err := time.LoadLocation(toTZ)
	if err != nil {
		s.logger.Error("Failed to load destination timezone",
			zap.String("to_timezone", toTZ),
			zap.Error(err))
		return time.Time{}, fmt.Errorf("invalid destination timezone %s: %w", toTZ, err)
	}

	// If the time doesn't have location info and fromTZ is specified, set it
	if fromTZ != "" && t.Location() == time.UTC {
		fromLoc, err := time.LoadLocation(fromTZ)
		if err != nil {
			s.logger.Error("Failed to load source timezone",
				zap.String("from_timezone", fromTZ),
				zap.Error(err))
			return time.Time{}, fmt.Errorf("invalid source timezone %s: %w", fromTZ, err)
		}
		// Interpret the time as being in the source timezone
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), fromLoc)
	}

	convertedTime := t.In(toLoc)

	s.logger.Debug("Successfully converted timezone",
		zap.String("from_timezone", fromTZ),
		zap.String("to_timezone", toTZ),
		zap.Time("original_time", t),
		zap.Time("converted_time", convertedTime))

	return convertedTime, nil
}

// IsFormatSupported checks if a format is supported
func (s *timeService) IsFormatSupported(format string) bool {
	for _, supported := range s.supportedFormats {
		if supported == format {
			return true
		}
	}
	return false
}

// GetSupportedFormats returns a list of supported formats
func (s *timeService) GetSupportedFormats() []string {
	formats := make([]string, len(s.supportedFormats))
	copy(formats, s.supportedFormats)
	return formats
}

// Helper functions

// isDST checks if the given time is in daylight saving time
func (s *timeService) isDST(t time.Time, loc *time.Location) bool {
	// Get the time zone info for the given time
	_, offset1 := t.Zone()

	// Check the same date but 6 months earlier/later to see if offset changes
	sixMonthsLater := t.AddDate(0, 6, 0)
	_, offset2 := sixMonthsLater.In(loc).Zone()

	// If the current offset is greater than the offset 6 months away,
	// we're likely in DST (this is a heuristic and may not be perfect for all zones)
	return offset1 > offset2
}

// getNextDSTTransition finds the next DST transition
func (s *timeService) getNextDSTTransition(t time.Time, loc *time.Location) *DSTTransitionInfo {
	// This is a simplified implementation
	// In a production system, you might want to use a more sophisticated approach
	// or a library that has complete DST transition data

	current := t
	_, currentOffset := current.Zone()

	// Look forward up to 1 year to find a transition
	for i := 0; i < 365; i++ {
		next := current.AddDate(0, 0, 1)
		nextInZone := next.In(loc)
		_, nextOffset := nextInZone.Zone()

		if nextOffset != currentOffset {
			transitionType := "enter_dst"
			if nextOffset < currentOffset {
				transitionType = "exit_dst"
			}

			return &DSTTransitionInfo{
				NextTransition: nextInZone,
				TransitionType: transitionType,
				OffsetChange:   nextOffset - currentOffset,
			}
		}
		current = next
	}

	return nil // No transition found within a year
}

// formatOffset formats a timezone offset in seconds to a human-readable string
func formatOffset(offsetSeconds int) string {
	if offsetSeconds == 0 {
		return "+00:00"
	}

	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60

	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}
