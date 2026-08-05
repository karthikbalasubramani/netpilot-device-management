package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
)

const (
	internalServerErrorMessage = "An unexpected internal server error occurred"
	internalServerErrorCode    = "INTERNAL_SERVER_ERROR"
)

// recoveryErrorResponse is returned when NetPilot recovers from an unexpected
// panic while processing an HTTP request.
type recoveryErrorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code"`
	RequestID string `json:"request_id,omitempty"`
}

// Recovery catches unexpected panics raised while processing HTTP requests.
// Panic details and stack traces are written only to the application logger.
// Clients receive a safe and generic error response.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(
		func(ctx *gin.Context, recovered any) {
			requestID := ctx.GetString(RequestIDKey)
			logger.Error(
				"Recovered from HTTP request panic",
				"request_id", requestID,
				"method", ctx.Request.Method,
				"path", ctx.Request.URL.Path,
				"client_ip", ctx.ClientIP(),
				"panic", fmt.Sprint(recovered),
				"stack_trace", string(debug.Stack()),
			)
			if ctx.Writer.Written() {
				ctx.Abort()
				return
			}

			ctx.AbortWithStatusJSON(
				http.StatusInternalServerError,
				recoveryErrorResponse{
					Success:   false,
					Message:   internalServerErrorMessage,
					ErrorCode: internalServerErrorCode,
					RequestID: requestID,
				},
			)
		},
	)
}
