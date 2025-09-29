package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/hspedro/mcp-server-time/internal/config"
	"github.com/hspedro/mcp-server-time/internal/logger"
	"github.com/hspedro/mcp-server-time/internal/metrics"
	"github.com/hspedro/mcp-server-time/internal/server"
	timeservice "github.com/hspedro/mcp-server-time/internal/time"
	"github.com/hspedro/mcp-server-time/internal/tools"
)

// App represents the MCP Time Server application
type App struct {
	config     *config.Config
	logger     *zap.Logger
	httpServer *server.HTTPServer
}

// New creates a new App instance
func New(version, buildTime string) (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup logger
	appLogger, err := logger.New(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	appLogger.Info("Starting MCP Time Server",
		zap.String("version", version),
		zap.String("build_time", buildTime),
		zap.String("server_name", cfg.Server.Name),
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.Bool("metrics_enabled", cfg.Metrics.Enabled))

	// Initialize components
	metricsCollector := metrics.New()
	timeService := timeservice.NewTimeService(
		cfg.Time.DefaultTimezone,
		cfg.Time.DefaultFormat,
		cfg.Time.SupportedFormats,
		appLogger,
	)

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	// Register time tools
	tools.RegisterTimeTools(mcpServer, timeService, metricsCollector, appLogger)

	// Create HTTP server
	httpServer := server.NewHTTPServer(cfg, mcpServer, metricsCollector, appLogger)

	return &App{
		config:     cfg,
		logger:     appLogger,
		httpServer: httpServer,
	}, nil
}

// Run starts the application and handles graceful shutdown
func (a *App) Run() error {
	// Start HTTP server in background
	serverErr := make(chan error, 1)
	go func() {
		if err := a.httpServer.Start(); err != nil {
			a.logger.Error("Server failed", zap.Error(err))
			serverErr <- err
		}
	}()

	// Wait for either interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		a.logger.Info("Received shutdown signal")
	case err := <-serverErr:
		a.logger.Error("Server failed to start", zap.Error(err))
		return err
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.Server.GracefulShutdownTimeout)
	defer cancel()

	// Shutdown gracefully
	return a.httpServer.Shutdown(shutdownCtx)
}

// Close performs cleanup operations
func (a *App) Close() error {
	if a.logger != nil {
		return a.logger.Sync()
	}
	return nil
}
