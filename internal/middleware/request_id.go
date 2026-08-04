package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

// RequestID adds a unique request ID to every incoming HTTP request.
// If the client already sends X-Request-ID, the same value is reused.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestId := ctx.GetHeader(RequestIDHeader)
		if requestId == "" {
			requestId = uuid.NewString()
		}
		ctx.Set(RequestIDKey, requestId)
		ctx.Header(RequestIDHeader, requestId)
		ctx.Next()
	}
}
