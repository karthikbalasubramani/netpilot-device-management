package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/health"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

const (
	serviceNotReadyMessage = "The service is not ready to accept requests"
	serviceNotReadyCode    = "SERVICE_NOT_READY"
)

// registerHealthRoutes registers infrastructure health endpoints.
func (server *Server) registerHealthRoutes() {
	server.router.GET("/health", server.healthCheck)
	server.router.GET("/health/live", server.livenessProbe)
	server.router.GET("/health/ready", server.readinessProbe)
	server.router.GET(
		"/debug/panic",
		func(ctx *gin.Context) {
			panic("Manual recovery middleware test")
		},
	)
	logger.Debug(
		"Health routes registered successfully",
		"routes",
		[]string{
			"/health",
			"/health/live",
			"/health/ready",
			"/debug/panic",
		},
	)
}

// livenessProbe confirms that the NetPilot process can handle HTTP requests.
//
// External dependencies such as MongoDB are intentionally not checked here.
func (server *Server) livenessProbe(ctx *gin.Context) {
	response.SuccessResponse(
		ctx,
		http.StatusOK,
		"Liveness Check Successful",
		gin.H{
			"status":      "alive",
			"service":     server.config.AppName,
			"environment": server.config.AppEnv,
		},
	)
}

// readinessProbe confirms that NetPilot can reach the dependencies required
// to process normal application requests.
func (server *Server) readinessProbe(ctx *gin.Context) {
	readinessContext, cancel := context.WithTimeout(
		ctx.Request.Context(),
		server.healthProbeConfig.ReadinessTimeout,
	)
	defer cancel()

	if err := server.checkReadiness(readinessContext); err != nil {
		requestID := ctx.GetString(middleware.RequestIDKey)

		logger.Warn(
			"Readiness check failed",
			"request_id", requestID,
			"dependency", "mongodb",
			"timeout",
			server.healthProbeConfig.ReadinessTimeout.String(),
			"error", err,
		)

		response.WriteHTTPError(
			ctx,
			http.StatusServiceUnavailable,
			serviceNotReadyMessage,
			serviceNotReadyCode,
			requestID,
		)

		return
	}

	response.SuccessResponse(
		ctx,
		http.StatusOK,
		"Readiness Check Successful",
		gin.H{
			"status":      "ready",
			"service":     server.config.AppName,
			"environment": server.config.AppEnv,
			"dependencies": gin.H{
				"mongodb": "available",
			},
		},
	)
}

// healthCheck returns the detailed application and system health status.
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
