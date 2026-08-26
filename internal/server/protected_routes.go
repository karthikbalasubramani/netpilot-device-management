package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

// registerProtectedV1Routes registers API v1 routes that require a valid
// access token.
func (server *Server) registerProtectedV1Routes(
	protectedV1 *gin.RouterGroup,
) {
	protectedV1.GET(
		"/auth/verify",
		server.verifyAuthentication,
	)
}

// verifyAuthentication confirms that the request passed JWT authentication.
//
// User identity is intentionally not returned yet. Authenticated-user context
// will be added in the next ticket.
func (server *Server) verifyAuthentication(
	ctx *gin.Context,
) {
	response.SuccessResponse(
		ctx,
		http.StatusOK,
		"Authentication successful",
		gin.H{
			"status": "authenticated",
		},
	)
}
