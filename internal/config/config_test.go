package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
		wantErr  bool
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name:     "default configuration",
			setupEnv: func() {},
			wantErr:  false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "mcp-server-time", cfg.Server.Name)
				assert.Equal(t, "localhost", cfg.Server.Host)
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "UTC", cfg.Time.DefaultTimezone)
				assert.Equal(t, "RFC3339", cfg.Time.DefaultFormat)
				assert.Contains(t, cfg.Time.SupportedFormats, "RFC3339")
				assert.Equal(t, "info", cfg.Logging.Level)
				assert.True(t, cfg.Metrics.Enabled)
				assert.Equal(t, 9080, cfg.Metrics.Port)
			},
		},
		{
			name: "environment variable overrides",
			setupEnv: func() {
				os.Setenv("MCP_SERVER_PORT", "8081")
				os.Setenv("MCP_SERVER_HOST", "0.0.0.0")
				os.Setenv("MCP_TIME_DEFAULT_TIMEZONE", "America/New_York")
				os.Setenv("MCP_LOGGING_LEVEL", "debug")
				os.Setenv("MCP_METRICS_ENABLED", "false")
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 8081, cfg.Server.Port)
				assert.Equal(t, "0.0.0.0", cfg.Server.Host)
				assert.Equal(t, "America/New_York", cfg.Time.DefaultTimezone)
				assert.Equal(t, "debug", cfg.Logging.Level)
				assert.False(t, cfg.Metrics.Enabled)
			},
		},
		{
			name: "invalid port should fail validation",
			setupEnv: func() {
				os.Setenv("MCP_SERVER_PORT", "0")
			},
			wantErr: true,
		},
		{
			name: "invalid timezone should fail validation",
			setupEnv: func() {
				os.Setenv("MCP_SERVER_TIME_TIME_DEFAULT_TIMEZONE", "Invalid/Timezone")
			},
			wantErr: true,
		},
		{
			name: "invalid log level should fail validation",
			setupEnv: func() {
				os.Setenv("MCP_SERVER_TIME_LOGGING_LEVEL", "invalid")
			},
			wantErr: true,
		},
		{
			name: "same port for server and metrics should fail",
			setupEnv: func() {
				os.Setenv("MCP_SERVER_TIME_SERVER_PORT", "8080")
				os.Setenv("MCP_SERVER_TIME_METRICS_PORT", "8080")
				os.Setenv("MCP_SERVER_TIME_METRICS_ENABLED", "true")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear viper state
			viper.Reset()

			// Clear environment
			envVars := []string{
				"MCP_SERVER_TIME_SERVER_PORT", "MCP_SERVER_TIME_SERVER_HOST", "MCP_SERVER_TIME_TIME_DEFAULT_TIMEZONE",
				"MCP_SERVER_TIME_LOGGING_LEVEL", "MCP_SERVER_TIME_METRICS_ENABLED", "MCP_SERVER_TIME_METRICS_PORT",
			}
			for _, env := range envVars {
				os.Unsetenv(env)
			}

			// Setup test environment
			tt.setupEnv()

			// Defer cleanup
			defer func() {
				for _, env := range envVars {
					os.Unsetenv(env)
				}
				viper.Reset()
			}()

			cfg, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Server: ServerConfig{
					Name:                    "test-server",
					Host:                    "localhost",
					Port:                    8080,
					GracefulShutdownTimeout: 30 * time.Second,
				},
				Time: TimeConfig{
					DefaultTimezone:  "UTC",
					DefaultFormat:    "RFC3339",
					SupportedFormats: []string{"RFC3339", "Unix"},
				},
				Logging: LogConfig{
					Level:  "info",
					Format: "json",
				},
				Metrics: MetricsConfig{
					Enabled: true,
					Port:    9090,
					Path:    "/metrics",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid server port - zero",
			config: &Config{
				Server:  ServerConfig{Port: 0},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
			errMsg:  "server.port must be between 1 and 65535",
		},
		{
			name: "invalid server port - too high",
			config: &Config{
				Server:  ServerConfig{Port: 70000},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
			errMsg:  "server.port must be between 1 and 65535",
		},
		{
			name: "empty server host",
			config: &Config{
				Server:  ServerConfig{Host: "", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
			errMsg:  "server.host cannot be empty",
		},
		{
			name: "invalid timezone",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "Invalid/Zone", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
			errMsg:  "invalid default timezone",
		},
		{
			name: "empty default format",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
			errMsg:  "time.default_format cannot be empty",
		},
		{
			name: "empty supported formats",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{}},
				Logging: LogConfig{Level: "info", Format: "json"},
			},
			wantErr: true,
			errMsg:  "time.supported_formats cannot be empty",
		},
		{
			name: "invalid log level",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "invalid", Format: "json"},
			},
			wantErr: true,
			errMsg:  "invalid logging.level",
		},
		{
			name: "invalid log format",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "invalid"},
			},
			wantErr: true,
			errMsg:  "invalid logging.format",
		},
		{
			name: "same ports for server and metrics",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
				Metrics: MetricsConfig{Enabled: true, Port: 8080, Path: "/metrics"},
			},
			wantErr: true,
			errMsg:  "metrics.port (8080) cannot be the same as server.port (8080)",
		},
		{
			name: "invalid metrics path",
			config: &Config{
				Server:  ServerConfig{Host: "localhost", Port: 8080},
				Time:    TimeConfig{DefaultTimezone: "UTC", DefaultFormat: "RFC3339", SupportedFormats: []string{"RFC3339"}},
				Logging: LogConfig{Level: "info", Format: "json"},
				Metrics: MetricsConfig{Enabled: true, Port: 9090, Path: "metrics"},
			},
			wantErr: true,
			errMsg:  "metrics.path must start with '/'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimeConfig_IsFormatSupported(t *testing.T) {
	config := &TimeConfig{
		SupportedFormats: []string{"RFC3339", "Unix", "UnixMilli"},
	}

	tests := []struct {
		format   string
		expected bool
	}{
		{"RFC3339", true},
		{"Unix", true},
		{"UnixMilli", true},
		{"RFC822", false},
		{"", false},
		{"INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := config.IsFormatSupported(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeConfig_GetValidFormats(t *testing.T) {
	original := []string{"RFC3339", "Unix", "UnixMilli"}
	config := &TimeConfig{
		SupportedFormats: original,
	}

	result := config.GetValidFormats()

	// Should return a copy, not the original slice
	assert.Equal(t, original, result)
	assert.NotSame(t, &original, &result)

	// Modifying the result should not affect the original
	result[0] = "Modified"
	assert.Equal(t, "RFC3339", config.SupportedFormats[0])
}
