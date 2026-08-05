package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

const (
	routeNotFoundMessage = "The requested API endpoint was not found"
	routeNotFoundCode    = "ROUTE_NOT_FOUND"

	methodNotAllowedMessage = "The HTTP method is not allowed for this endpoint"
	methodNotAllowedCode    = "METHOD_NOT_ALLOWED"
)

// registerErrorHandlers configures standard responses for unknown routes and
// unsupported HTTP methods.
func registerErrorHandlers(router *gin.Engine) {
	router.HandleMethodNotAllowed = true

	router.NoRoute(handleRouteNotFound)
	router.NoMethod(handleMethodNotAllowed)
}

func handleRouteNotFound(ctx *gin.Context) {
	response.WriteHTTPError(
		ctx,
		http.StatusNotFound,
		routeNotFoundMessage,
		routeNotFoundCode,
		ctx.GetString(middleware.RequestIDKey),
	)
}

func handleMethodNotAllowed(ctx *gin.Context) {
	response.WriteHTTPError(
		ctx,
		http.StatusMethodNotAllowed,
		methodNotAllowedMessage,
		methodNotAllowedCode,
		ctx.GetString(middleware.RequestIDKey),
	)
}
