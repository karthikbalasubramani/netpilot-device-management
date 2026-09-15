package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/auth"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

// registerProtectedV1Routes registers API v1 routes that require a valid
// access token.
func (server *Server) registerProtectedV1Routes(
	protectedV1 *gin.RouterGroup,
	tokenVerifier auth.AccessTokenVerifier,
	userService *user.Service,
) {
	protectedAuth := protectedV1.Group("/auth")

	protectedAuth.GET(
		"/verify",
		server.verifyAuthentication,
	)

	protectedAuth.GET(
		"/admin/verify",
		middleware.RequiredRoles(
			user.RoleAdmin,
		),
		verifyAdminAuthorization,
	)

	adminUsers := protectedV1.Group("/admin/users")

	adminUsers.Use(
		middleware.RequiredRoles(
			user.RoleAdmin,
		),
	)

	adminUsers.PATCH(
		"/:user_id/role",
		updateUserRoleHandler(
			userService,
		),
	)
}

// verifyAuthentication confirms that the request passed JWT authentication.
//
// User identity is intentionally not returned yet. Authenticated-user context
// will be added in the next ticket.
func (server *Server) verifyAuthentication(
	ctx *gin.Context,
) {
	authenticatedUser, ok := middleware.GetAuthenticatedUser(ctx)

	if !ok {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success":    false,
				"message":    "Authenticated user context is unavailable",
				"error_code": "AUTHENTICATION_CONTEXT_MISSING",
				"request_id": ctx.GetString(
					middleware.RequestIDKey,
				),
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"message": "Access token is valid",
			"data": gin.H{
				"user_id": authenticatedUser.UserID,
				"role":    authenticatedUser.Role,
			},
		},
	)
}

func verifyAdminAuthorization(
	ctx *gin.Context,
) {
	authenticatedUser, ok :=
		middleware.GetAuthenticatedUser(ctx)

	if !ok {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success":    false,
				"message":    "Authenticated user context is unavailable",
				"error_code": "AUTHENTICATION_CONTEXT_MISSING",
				"request_id": ctx.GetString(
					middleware.RequestIDKey,
				),
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"message": "Admin authorization verified",
			"data": gin.H{
				"user_id": authenticatedUser.UserID,
				"role":    authenticatedUser.Role,
			},
		},
	)
}
