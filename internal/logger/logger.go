package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/hspedro/mcp-server-time/internal/config"
)

// New creates a new zap logger based on the provided configuration
func New(cfg config.LogConfig) (*zap.Logger, error) {
	level := parseLogLevel(cfg.Level)

	var logger *zap.Logger
	var err error

	if cfg.Format == "json" {
		logger, err = newProductionLogger(level)
	} else {
		logger, err = newDevelopmentLogger(level)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// parseLogLevel converts string log level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// newProductionLogger creates a production-ready logger with JSON output
func newProductionLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	return config.Build()
}

// newDevelopmentLogger creates a development logger with console output
func newDevelopmentLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	return config.Build()
}
