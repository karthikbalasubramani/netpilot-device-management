package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/health"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

// Server holds the HTTP router, HTTP server instance, and application
// configuration.
type Server struct {
	config     *config.Config
	router     *gin.Engine
	httpServer *http.Server
}

// NewHTTPServer creates and configures the NetPilot HTTP server.
func NewHTTPServer(
	cfg *config.Config,
	httpServerConfig config.HTTPServerConfig,
) (*Server, error) {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a Gin router without Gin's default logging and recovery
	// middleware.
	router := gin.New()

	// NetPilot currently receives requests directly without a trusted reverse
	// proxy. Disable proxy-header trust so clients cannot spoof their IP
	// address through X-Forwarded-For or X-Real-IP.
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf(
			"disable trusted proxy headers: %w",
			err,
		)
	}

	router.Use(
		middleware.RequestID(),
		middleware.RequestLogger(),
		middleware.SecurityHeaders(),
		middleware.Recovery(),
	)

	// Register standard handlers for HTTP 404 and 405 responses.
	registerErrorHandlers(router)

	server := &Server{
		config: cfg,
		router: router,
	}

	server.registerRoutes()

	server.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.AppPort),
		Handler:           router,
		ReadHeaderTimeout: httpServerConfig.ReadHeaderTimeout,
		ReadTimeout:       httpServerConfig.ReadTimeout,
		WriteTimeout:      httpServerConfig.WriteTimeout,
		IdleTimeout:       httpServerConfig.IdleTimeout,
		MaxHeaderBytes:    httpServerConfig.MaxHeaderBytes,
	}

	return server, nil
}

// StartHTTPServer starts the HTTP server on the configured application port.
func (server *Server) StartHTTPServer() error {
	logger.Info(
		"Starting HTTP server",
		"port", server.config.AppPort,
	)

	return server.httpServer.ListenAndServe()
}

// ShutdownHTTPServer gracefully shuts down the HTTP server.
func (server *Server) ShutdownHTTPServer(ctx context.Context) error {
	if server.httpServer == nil {
		return nil
	}

	if err := server.httpServer.Shutdown(ctx); err != nil {
		logger.Error(
			"Failed to shut down HTTP server",
			"error", err,
		)

		return fmt.Errorf(
			"shut down HTTP server: %w",
			err,
		)
	}

	return nil
}

// registerRoutes registers all HTTP routes for the application.
func (server *Server) registerRoutes() {
	server.router.GET("/health", server.healthCheck)
	server.router.GET(
		"/debug/panic",
		func(ctx *gin.Context) {
			panic("manual recovery middleware test")
		},
	)
	logger.Debug("Health routes registered successfully")
}

// healthCheck returns the current application and system health status.
func (server *Server) healthCheck(ctx *gin.Context) {
	systemState, err := health.GetSystemInfoHealth(
		server.config.CPUThresholdPercent,
		server.config.DiskPath,
	)
	if err != nil {
		logger.Warn(
			"Health check degraded",
			"error", err,
		)

		response.ErrorResponse(
			ctx,
			http.StatusServiceUnavailable,
			"Health Check Degraded",
			err,
		)
		return
	}

	response.SuccessResponse(
		ctx,
		http.StatusOK,
		"Health Check Successful",
		gin.H{
			"status":        "ok",
			"service":       server.config.AppName,
			"environment":   server.config.AppEnv,
			"server_status": "Running",
			"system_state":  systemState,
		},
	)
}
