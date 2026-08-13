package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

const apiVersionV1 = "v1"

// Register V1 routes for v1 version API's
func (server *Server) registerV1Routes(apiV1 *gin.RouterGroup) {
	apiV1.GET("", server.apiV1Info)

	server.registerAuthRoutes(apiV1)

	logger.Debug(
		"API v1 routes registered successfully",
		"base_path", ApiV1basepath,
	)
}

// API V1 Version routes
func (server *Server) apiV1Info(ctx *gin.Context) {
	response.SuccessResponse(
		ctx,
		http.StatusOK,
		"NetPilot API Version v1 is available",
		gin.H{
			"service":     server.config.AppName,
			"environment": server.config.AppEnv,
			"api_version": apiVersionV1,
			"status":      "available",
		},
	)
}
