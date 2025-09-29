package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewTimeService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	supportedFormats := []string{"RFC3339", "Unix"}

	service := NewTimeService("UTC", "RFC3339", supportedFormats, logger)

	assert.NotNil(t, service)
	assert.Equal(t, supportedFormats, service.GetSupportedFormats())
}

func TestTimeService_GetCurrentTime(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewTimeService("UTC", "RFC3339", []string{"RFC3339"}, logger)

	tests := []struct {
		name    string
		input   GetTimeInput
		wantErr bool
	}{
		{
			name:    "default timezone (empty string)",
			input:   GetTimeInput{Timezone: "", Format: "RFC3339"},
			wantErr: false,
		},
		{
			name:    "UTC timezone",
			input:   GetTimeInput{Timezone: "UTC", Format: "RFC3339"},
			wantErr: false,
		},
		{
			name:    "New York timezone",
			input:   GetTimeInput{Timezone: "America/New_York", Format: "RFC3339"},
			wantErr: false,
		},
		{
			name:    "invalid timezone",
			input:   GetTimeInput{Timezone: "Invalid/Timezone", Format: "RFC3339"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetCurrentTime(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid timezone")
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result.FormattedTime)
			assert.NotZero(t, result.UnixTimestamp)

			// Verify the timezone is correct
			expectedTimezone := tt.input.Timezone
			if expectedTimezone == "" {
				expectedTimezone = "UTC"
			}
			assert.Equal(t, expectedTimezone, result.Timezone)
		})
	}
}

func TestTimeService_FormatTime(t *testing.T) {
	logger := zaptest.NewLogger(t)
	supportedFormats := []string{"RFC3339", "Unix", "UnixMilli", "2006-01-02 15:04:05"}
	service := NewTimeService("UTC", "RFC3339", supportedFormats, logger)

	testTime := time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name     string
		input    FormatTimeInput
		expected string
		wantErr  bool
	}{
		{
			name:     "default format (empty string)",
			input:    FormatTimeInput{Timestamp: testTime, Format: "", Timezone: "UTC"},
			expected: "2023-12-25T15:30:45Z",
			wantErr:  false,
		},
		{
			name:     "RFC3339 format",
			input:    FormatTimeInput{Timestamp: testTime, Format: "RFC3339", Timezone: "UTC"},
			expected: "2023-12-25T15:30:45Z",
			wantErr:  false,
		},
		{
			name:     "Unix format",
			input:    FormatTimeInput{Timestamp: testTime, Format: "Unix", Timezone: "UTC"},
			expected: "1703518245",
			wantErr:  false,
		},
		{
			name:     "UnixMilli format",
			input:    FormatTimeInput{Timestamp: testTime, Format: "UnixMilli", Timezone: "UTC"},
			expected: "1703518245123",
			wantErr:  false,
		},
		{
			name:     "custom layout",
			input:    FormatTimeInput{Timestamp: testTime, Format: "2006-01-02 15:04:05", Timezone: "UTC"},
			expected: "2023-12-25 15:30:45",
			wantErr:  false,
		},
		{
			name:    "unsupported format",
			input:   FormatTimeInput{Timestamp: testTime, Format: "UnsupportedFormat", Timezone: "UTC"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FormatTime(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.FormattedTime)
		})
	}
}

func TestTimeService_ParseTime(t *testing.T) {
	logger := zaptest.NewLogger(t)
	supportedFormats := []string{"RFC3339", "Unix", "UnixMilli"}
	service := NewTimeService("UTC", "RFC3339", supportedFormats, logger)

	tests := []struct {
		name     string
		input    ParseTimeInput
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "RFC3339 format",
			input:    ParseTimeInput{TimeString: "2023-12-25T15:30:45Z", Format: "RFC3339"},
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Unix format",
			input:    ParseTimeInput{TimeString: "1703518245", Format: "Unix"},
			expected: time.Unix(1703518245, 0),
			wantErr:  false,
		},
		{
			name:     "UnixMilli format",
			input:    ParseTimeInput{TimeString: "1703518245123", Format: "UnixMilli"},
			expected: time.UnixMilli(1703518245123),
			wantErr:  false,
		},
		{
			name:     "custom layout",
			input:    ParseTimeInput{TimeString: "2023-12-25 15:30:45", Format: "2006-01-02 15:04:05"},
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "invalid time string for RFC3339",
			input:   ParseTimeInput{TimeString: "invalid-time", Format: "RFC3339"},
			wantErr: true,
		},
		{
			name:    "invalid Unix timestamp",
			input:   ParseTimeInput{TimeString: "invalid-unix", Format: "Unix"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ParseTime(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to parse time string")
				return
			}

			require.NoError(t, err)
			// Compare unix timestamps since they're easier to work with
			assert.Equal(t, tt.expected.Unix(), result.UnixTimestamp, "Expected unix %d, got %d", tt.expected.Unix(), result.UnixTimestamp)
		})
	}
}

func TestTimeService_GetTimezoneInfo(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewTimeService("UTC", "RFC3339", []string{"RFC3339"}, logger)

	tests := []struct {
		name     string
		input    TimezoneInfoInput
		wantErr  bool
		validate func(t *testing.T, info TimezoneInfo)
	}{
		{
			name:    "UTC timezone",
			input:   TimezoneInfoInput{Timezone: "UTC"},
			wantErr: false,
			validate: func(t *testing.T, info TimezoneInfo) {
				assert.Equal(t, "UTC", info.Name)
				assert.Equal(t, "UTC", info.Abbreviation)
				assert.Equal(t, "+00:00", info.Offset)
				assert.Equal(t, 0, info.OffsetSeconds)
			},
		},
		{
			name:    "America/New_York timezone",
			input:   TimezoneInfoInput{Timezone: "America/New_York"},
			wantErr: false,
			validate: func(t *testing.T, info TimezoneInfo) {
				assert.Equal(t, "America/New_York", info.Name)
				assert.Contains(t, []string{"EST", "EDT"}, info.Abbreviation)
			},
		},
		{
			name:    "invalid timezone",
			input:   TimezoneInfoInput{Timezone: "Invalid/Timezone"},
			wantErr: true,
		},
		{
			name:    "empty timezone (should use default)",
			input:   TimezoneInfoInput{Timezone: ""},
			wantErr: false,
			validate: func(t *testing.T, info TimezoneInfo) {
				assert.Equal(t, "UTC", info.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetTimezoneInfo(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid timezone")
				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestTimeService_ConvertTimezone(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewTimeService("UTC", "RFC3339", []string{"RFC3339"}, logger)

	// Create a time in UTC
	utcTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		fromTZ   string
		toTZ     string
		wantErr  bool
		validate func(t *testing.T, original, converted time.Time)
	}{
		{
			name:    "UTC to America/New_York",
			fromTZ:  "UTC",
			toTZ:    "America/New_York",
			wantErr: false,
			validate: func(t *testing.T, original, converted time.Time) {
				// The time instant should be the same (Unix timestamp)
				assert.Equal(t, original.Unix(), converted.Unix())
				// But the location should be different
				assert.NotEqual(t, original.Location().String(), converted.Location().String())
			},
		},
		{
			name:    "UTC to Europe/London",
			fromTZ:  "UTC",
			toTZ:    "Europe/London",
			wantErr: false,
			validate: func(t *testing.T, original, converted time.Time) {
				assert.Equal(t, original.Unix(), converted.Unix())
			},
		},
		{
			name:    "invalid destination timezone",
			fromTZ:  "UTC",
			toTZ:    "Invalid/Timezone",
			wantErr: true,
		},
		{
			name:    "invalid source timezone",
			fromTZ:  "Invalid/Timezone",
			toTZ:    "UTC",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ConvertTimezone(utcTime, tt.fromTZ, tt.toTZ)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid")
				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, utcTime, result)
			}
		})
	}
}

func TestTimeService_IsFormatSupported(t *testing.T) {
	logger := zaptest.NewLogger(t)
	supportedFormats := []string{"RFC3339", "Unix", "UnixMilli"}
	service := NewTimeService("UTC", "RFC3339", supportedFormats, logger)

	tests := []struct {
		format   string
		expected bool
	}{
		{"RFC3339", true},
		{"Unix", true},
		{"UnixMilli", true},
		{"UnsupportedFormat", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := service.IsFormatSupported(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeService_GetSupportedFormats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	supportedFormats := []string{"RFC3339", "Unix", "UnixMilli"}
	service := NewTimeService("UTC", "RFC3339", supportedFormats, logger)

	result := service.GetSupportedFormats()

	assert.Equal(t, supportedFormats, result)
	// Ensure it returns a copy, not the original slice
	result[0] = "Modified"
	assert.Equal(t, "RFC3339", service.GetSupportedFormats()[0])
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format   string
		expected bool
	}{
		{"RFC3339", true},
		{"RFC3339Nano", true},
		{"Unix", true},
		{"UnixMilli", true},
		{"UnixMicro", true},
		{"UnixNano", true},
		{"Layout", true},
		{"InvalidFormat", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := IsValidFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFormatLayout(t *testing.T) {
	tests := []struct {
		format   FormatType
		expected string
	}{
		{FormatRFC3339, time.RFC3339},
		{FormatRFC3339Nano, time.RFC3339Nano},
		{FormatLayout, time.RFC3339}, // default fallback
		{FormatUnix, time.RFC3339},   // default fallback
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			result := GetFormatLayout(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatOffset(t *testing.T) {
	tests := []struct {
		offsetSeconds int
		expected      string
	}{
		{0, "+00:00"},
		{3600, "+01:00"},
		{-3600, "-01:00"},
		{5400, "+01:30"},
		{-5400, "-01:30"},
		{43200, "+12:00"},
		{-43200, "-12:00"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatOffset(tt.offsetSeconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}
