package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/health"
)

// Server holds the HTTP router and application configuration.
type Server struct {
	config *config.Config
	router *gin.Engine
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

	return server
}

// StartHTTPServer starts the HTTP server on the configured application port.
func (s *Server) StartHTTPServer() error {
	address := fmt.Sprintf(":%s", s.config.AppPort)
	return s.router.Run(address)
}

// registerRoutes registers all HTTP routes for the application.
func (s *Server) registerRoutes() {
	s.router.GET("/health", s.healthCheck)
}

// healthCheck returns the current application and system health status.
func (s *Server) healthCheck(ctx *gin.Context) {
	// Collect current system state such as CPU, memory, disk, and uptime.
	systemState, err := health.GetSystemInfoHealth(s.config.CPUThresholdPercent)
	if err != nil {
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
