package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/hspedro/mcp-server-time/internal/config"
	"github.com/hspedro/mcp-server-time/internal/metrics"
)

// HTTPServer wraps HTTP server functionality
type HTTPServer struct {
	Server        *http.Server
	MetricsServer *http.Server
	logger        *zap.Logger
}

// NewHTTPServer creates a new HTTP server with MCP endpoints
func NewHTTPServer(cfg *config.Config, mcpServer *mcp.Server, metrics *metrics.Metrics, logger *zap.Logger) *HTTPServer {
	mux := setupMainHandler(cfg, mcpServer, metrics, logger)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
	}

	var metricsServer *http.Server
	if cfg.Metrics.Enabled && cfg.Metrics.Port != cfg.Server.Port {
		metricsServer = setupMetricsServer(cfg, logger)
	}

	return &HTTPServer{
		Server:        server,
		MetricsServer: metricsServer,
		logger:        logger,
	}
}

// setupMainHandler configures the main HTTP handler with all endpoints
func setupMainHandler(cfg *config.Config, mcpServer *mcp.Server, metrics *metrics.Metrics, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Create MCP transport handlers
	sseHandler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)

	streamableHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)

	// Register MCP endpoints with metrics
	mux.Handle("/sse", withMetrics(sseHandler, metrics, logger, "sse"))
	mux.Handle("/streamable", withMetrics(streamableHandler, metrics, logger, "streamable"))
	mux.Handle("/mcp", withMetrics(streamableHandler, metrics, logger, "streamable")) // Alias

	// Register health check
	mux.HandleFunc("/health", createHealthHandler(cfg))

	// Register metrics endpoint if enabled on same port
	if cfg.Metrics.Enabled && cfg.Metrics.Port == cfg.Server.Port {
		mux.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	return mux
}

// setupMetricsServer creates a separate metrics server if configured
func setupMetricsServer(cfg *config.Config, logger *zap.Logger) *http.Server {
	metricsMux := http.NewServeMux()
	metricsMux.Handle(cfg.Metrics.Path, promhttp.Handler())

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port),
		Handler: metricsMux,
	}
}

// createHealthHandler creates the health check endpoint handler
func createHealthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"%s","version":"%s","timestamp":"%s"}`,
			cfg.Server.Name, cfg.Server.Version, time.Now().UTC().Format(time.RFC3339))
	}
}

// Start starts both the main server and metrics server (if configured)
func (s *HTTPServer) Start() error {
	// Start metrics server in background if configured
	if s.MetricsServer != nil {
		go func() {
			s.logger.Info("Starting metrics server",
				zap.String("addr", s.MetricsServer.Addr))

			if err := s.MetricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error("Metrics server failed", zap.Error(err))
			}
		}()
	}

	// Start main server
	s.logger.Info("Starting MCP server",
		zap.String("addr", s.Server.Addr),
		zap.Strings("endpoints", []string{"/sse", "/streamable", "/mcp", "/health"}))

	return s.Server.ListenAndServe()
}

// Shutdown gracefully shuts down both servers
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down servers...")

	// Shutdown main server
	if err := s.Server.Shutdown(ctx); err != nil {
		s.logger.Error("Main server forced shutdown", zap.Error(err))
		return err
	}

	// Shutdown metrics server if running
	if s.MetricsServer != nil {
		if err := s.MetricsServer.Shutdown(ctx); err != nil {
			s.logger.Error("Metrics server forced shutdown", zap.Error(err))
			return err
		}
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

// withMetrics wraps an HTTP handler with metrics collection
func withMetrics(handler http.Handler, metrics *metrics.Metrics, logger *zap.Logger, transport string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		logger.Debug("MCP transport request",
			zap.String("transport", transport),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr))

		// Set CORS headers for all transports
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			metrics.RecordTransportRequest(transport, r.Method, "success")
			return
		}

		// Wrap response writer to capture status
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the actual handler
		handler.ServeHTTP(wrapped, r)

		// Record metrics
		status := "success"
		if wrapped.statusCode >= 400 {
			status = "error"
		}

		metrics.RecordTransportRequest(transport, r.Method, status)

		duration := time.Since(startTime)
		logger.Debug("MCP transport request completed",
			zap.String("transport", transport),
			zap.String("method", r.Method),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration))
	})
}

// responseWriterWrapper captures the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
