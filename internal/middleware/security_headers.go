package middleware

import "github.com/gin-gonic/gin"

const (
	headerContentTypeOptions           = "X-Content-Type-Options"
	headerFrameOptions                 = "X-Frame-Options"
	headerReferrerPolicy               = "Referrer-Policy"
	headerPermissionsPolicy            = "Permission-Policy"
	headerPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies"
	headerXSSProtection                = "X-XSS-Protection"
)

// SecurityHeaders adds common security-related HTTP headers to every response.
//
// CORS headers are intentionally not included here. Cross-origin access will
// be configured separately when NetPilot requires browser-based clients.
func SecurityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headers := ctx.Writer.Header()

		headers.Set(headerContentTypeOptions, "nosniff")
		headers.Set(headerFrameOptions, "DENY")
		headers.Set(headerReferrerPolicy, "no-referrer")
		headers.Set(headerPermissionsPolicy, "camera=(), microphone=(), geolocation=()")
		headers.Set(headerPermittedCrossDomainPolicies, "none")
		headers.Set(headerXSSProtection, "0")

		// Remove headers that could reveal server implementation details.
		headers.Del("Server")
		headers.Del("X-Powered-By")

		ctx.Next()
	}
}
