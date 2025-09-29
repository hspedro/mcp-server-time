package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Time    TimeConfig    `mapstructure:"time"`
	Logging LogConfig     `mapstructure:"logging"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Name                    string        `mapstructure:"name"`
	Version                 string        `mapstructure:"version"`
	Host                    string        `mapstructure:"host"`
	Port                    int           `mapstructure:"port"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
	ConnectionStaleTimeout  time.Duration `mapstructure:"connection_stale_timeout"`
}

// TimeConfig contains time service configuration
type TimeConfig struct {
	DefaultTimezone  string   `mapstructure:"default_timezone"`
	DefaultFormat    string   `mapstructure:"default_format"`
	SupportedFormats []string `mapstructure:"supported_formats"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MetricsConfig contains Prometheus metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set environment variable prefix and replacement
	viper.SetEnvPrefix("MCP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	// Read config file if available
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.name", "mcp-server-time")
	viper.SetDefault("server.version", "1.0.0")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.graceful_shutdown_timeout", "1s")
	viper.SetDefault("server.connection_stale_timeout", "2m")

	// Time service defaults
	viper.SetDefault("time.default_timezone", "UTC")
	viper.SetDefault("time.default_format", "RFC3339")
	viper.SetDefault("time.supported_formats", []string{
		"RFC3339",
		"RFC3339Nano",
		"Unix",
		"UnixMilli",
		"UnixMicro",
		"UnixNano",
		"Layout",
	})

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", 9080)
	viper.SetDefault("metrics.path", "/metrics")
}

// validate checks configuration for required values and consistency
func validate(config *Config) error {
	// Validate server configuration
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got: %d", config.Server.Port)
	}

	if config.Server.Host == "" {
		return fmt.Errorf("server.host cannot be empty")
	}

	// Validate time configuration
	if config.Time.DefaultTimezone == "" {
		return fmt.Errorf("time.default_timezone cannot be empty")
	}

	// Validate timezone by attempting to load it
	if _, err := time.LoadLocation(config.Time.DefaultTimezone); err != nil {
		return fmt.Errorf("invalid default timezone %s: %w", config.Time.DefaultTimezone, err)
	}

	if config.Time.DefaultFormat == "" {
		return fmt.Errorf("time.default_format cannot be empty")
	}

	// Validate supported formats are not empty
	if len(config.Time.SupportedFormats) == 0 {
		return fmt.Errorf("time.supported_formats cannot be empty")
	}

	// Validate logging configuration
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid logging.level: %s (must be one of: debug, info, warn, error, fatal)", config.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "console": true,
	}
	if !validLogFormats[config.Logging.Format] {
		return fmt.Errorf("invalid logging.format: %s (must be one of: json, console)", config.Logging.Format)
	}

	// Validate metrics configuration
	if config.Metrics.Enabled {
		if config.Metrics.Port <= 0 || config.Metrics.Port > 65535 {
			return fmt.Errorf("metrics.port must be between 1 and 65535, got: %d", config.Metrics.Port)
		}

		if config.Server.Port == config.Metrics.Port {
			return fmt.Errorf("metrics.port (%d) cannot be the same as server.port (%d)", config.Metrics.Port, config.Server.Port)
		}

		if config.Metrics.Path == "" {
			return fmt.Errorf("metrics.path cannot be empty when metrics are enabled")
		}

		if !strings.HasPrefix(config.Metrics.Path, "/") {
			return fmt.Errorf("metrics.path must start with '/', got: %s", config.Metrics.Path)
		}
	}

	return nil
}

// IsFormatSupported checks if a given format is in the supported formats list
func (c *TimeConfig) IsFormatSupported(format string) bool {
	for _, supported := range c.SupportedFormats {
		if supported == format {
			return true
		}
	}
	return false
}

// GetValidFormats returns a copy of the supported formats slice
func (c *TimeConfig) GetValidFormats() []string {
	formats := make([]string, len(c.SupportedFormats))
	copy(formats, c.SupportedFormats)
	return formats
}
