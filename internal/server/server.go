package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/auth"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
)

// ReadinessCheck verifies whether the dependencies required by NetPilot are
// available.
type ReadinessCheck func(ctx context.Context) error

// Server holds the HTTP router, HTTP server, application configuration, and
// health-probe dependencies.
type Server struct {
	config              *config.Config
	healthProbeConfig   config.HealthProbeConfig
	checkReadiness      ReadinessCheck
	authService         AuthService
	router              *gin.Engine
	httpServer          *http.Server
	accessTokenVerifier auth.AccessTokenVerifier
}

// NewHTTPServer creates and configures the NetPilot HTTP server.
func NewHTTPServer(
	cfg *config.Config,
	httpServerConfig config.HTTPServerConfig,
	healthProbeConfig config.HealthProbeConfig,
	readinessCheck ReadinessCheck,
	authService AuthService,
	accessTokenVerifier auth.AccessTokenVerifier,
) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf(
			"application configuration is required",
		)
	}

	if readinessCheck == nil {
		return nil, fmt.Errorf(
			"readiness check function is required",
		)
	}

	if authService == nil {
		return nil, fmt.Errorf(
			"authentication service is required",
		)
	}

	if accessTokenVerifier == nil {
		return nil, fmt.Errorf(
			"access token verifier is required",
		)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a Gin router without Gin's default logging and recovery
	// middleware.
	router := gin.New()

	// NetPilot currently receives requests directly without a trusted reverse
	// proxy. Disable proxy-header trust so clients cannot spoof their IP.
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
		middleware.RequestBodyLimit(
			httpServerConfig.MaxRequestBodyBytes,
		),
	)

	registerErrorHandlers(router)

	server := &Server{
		config:              cfg,
		healthProbeConfig:   healthProbeConfig,
		checkReadiness:      readinessCheck,
		authService:         authService,
		accessTokenVerifier: accessTokenVerifier,
		router:              router,
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
