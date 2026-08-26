package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

// RequireRoles authorizes an authenticated request when the user's role
// matches one of the supplied roles.
//
// Authentication middleware must run before this middleware so that a
// trusted AuthenticatedUser is already available in Gin's context.
func RequiredRoles(allowedRoles ...user.Role) gin.HandlerFunc {
	roleSet := make(map[user.Role]struct{}, len(allowedRoles))

	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(ctx *gin.Context) {
		authenticatedUser, ok := GetAuthenticatedUser(ctx)
		if !ok {
			abortAuthorizationUnauthenticated(ctx)
			return
		}

		_, allowed := roleSet[authenticatedUser.Role]
		if !allowed {
			abortForbidden(ctx)
			return
		}
	}
}

func abortAuthorizationUnauthenticated(ctx *gin.Context) {
	ctx.Header("WWW-Authenticate", bearerScheme)

	ctx.AbortWithStatusJSON(
		http.StatusUnauthorized,
		gin.H{
			"success":    false,
			"message":    "Authorization is required",
			"error_code": "UNAUTHORIZED",
			"request_id": ctx.GetString(RequestIDKey),
		},
	)
}

func abortForbidden(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(
		http.StatusForbidden,
		gin.H{
			"success":    false,
			"message":    "You do not have permission to access this resource",
			"error_code": "FORBIDDEN",
			"request_id": ctx.GetString(RequestIDKey),
		},
	)
}
