package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/health"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
)

// Server holds the HTTP router, HTTP server instance, and application configuration.
type Server struct {
	config     *config.Config
	router     *gin.Engine
	httpServer *http.Server
}

// NewHTTPServer creates a new HTTP server instance, configures middleware,
// disables default trusted proxies, and registers application routes.
func NewHTTPServer(cfg *config.Config) *Server {
	// Run Gin in release mode when application environment is production.
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a new Gin router without default middleware.
	router := gin.New()

	// Add required middleware explicitly.
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Disable trusting all proxies by default.
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(fmt.Errorf("failed to set trusted proxies: %w", err))
	}

	server := &Server{
		config: cfg,
		router: router,
	}

	server.registerRoutes()

	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.AppPort),
		Handler: router,
	}

	return server
}

// StartHTTPServer starts the HTTP server on the configured application port.
func (s *Server) StartHTTPServer() error {
	logger.Info("starting HTTP server", "port", s.config.AppPort)
	return s.httpServer.ListenAndServe()
}

// ShutdownHTTPServer gracefully shuts down the HTTP server.
func (s *Server) ShutdownHTTPServer(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}

// registerRoutes registers all HTTP routes for the application.
func (s *Server) registerRoutes() {
	s.router.GET("/health", s.healthCheck)
}

// healthCheck returns the current application and system health status.
func (s *Server) healthCheck(ctx *gin.Context) {
	// Collect current system state such as CPU, memory, disk, and uptime.
	systemState, err := health.GetSystemInfoHealth(s.config.CPUThresholdPercent, s.config.DiskPath)
	if err != nil {
		logger.Warn("health check degraded", "error", err)

		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"status":      "degraded",
			"service":     s.config.AppName,
			"environment": s.config.AppEnv,
			"error":       err.Error(),
		})
		return
	}

	// Return successful health response when application and system checks pass.
	ctx.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"service":       s.config.AppName,
		"environment":   s.config.AppEnv,
		"server_status": "Running",
		"system_state":  systemState,
	})
}
