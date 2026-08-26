package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/auth"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

const (
	authorizationHeader = "Authorization"
	bearerScheme        = "Bearer"

	authenticationRequiredMessage = "A valid access token is required"
	authenticationRequiredCode    = "UNAUTHORIZED"
	AuthenticatedUserKey          = "authenticated_user"
)

// Authentication validates the JWT access token supplied in the
// Authorization header before allowing a protected request to continue.
func Authentication(
	tokenVerifier auth.AccessTokenVerifier,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetString(
			RequestIDKey,
		)
		authorizationValue := strings.TrimSpace(
			ctx.GetHeader(authorizationHeader),
		)
		tokenString, ok := extractBearerToken(
			authorizationValue,
		)
		if !ok {
			writeUnauthorizedResponse(
				ctx,
				requestID,
			)

			return
		}

		claims, err := tokenVerifier.Verify(tokenString)

		if err != nil {
			logger.Warn(
				"Access token validation failed",
				"request_id", requestID,
				"error", err,
			)

			writeUnauthorizedResponse(
				ctx,
				requestID,
			)

			return
		}
		authenticatedUser := auth.AuthenticatedUser{
			UserID:  claims.Subject,
			Role:    claims.Role,
			TokenID: claims.ID,
		}

		ctx.Set(
			AuthenticatedUserKey,
			authenticatedUser,
		)
		ctx.Next()
	}
}

// GetAuthenticatedUser retrieves the verified authenticated-user identity
// placed into Gin context by Authentication middleware.
//
// The boolean is false when the authentication context does not exist
// or contains an unexpected type.
func GetAuthenticatedUser(ctx *gin.Context) (auth.AuthenticatedUser, bool) {
	value, exists := ctx.Get(AuthenticatedUserKey)
	if !exists {
		return auth.AuthenticatedUser{}, false
	}

	authenticatedUser, ok := value.(auth.AuthenticatedUser)
	if !ok {
		return auth.AuthenticatedUser{}, false
	}

	return authenticatedUser, true
}

func extractBearerToken(
	authorizationValue string,
) (string, bool) {
	parts := strings.Fields(
		authorizationValue,
	)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(
		parts[0],
		bearerScheme,
	) {
		return "", false
	}
	tokenString := strings.TrimSpace(
		parts[1],
	)
	if tokenString == "" {
		return "", false
	}
	return tokenString, true
}

func writeUnauthorizedResponse(
	ctx *gin.Context,
	requestID string,
) {
	ctx.Header(
		"WWW-Authenticate",
		bearerScheme,
	)
	response.WriteHTTPError(
		ctx,
		http.StatusUnauthorized,
		authenticationRequiredMessage,
		authenticationRequiredCode,
		requestID,
	)
	ctx.Abort()
}
