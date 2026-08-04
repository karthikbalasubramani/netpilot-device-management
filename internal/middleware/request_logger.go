package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
)

const unmatchedRoute = "unmatched"

func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()

		ctx.Next()
		duration := time.Since(startTime)
		statusCode := ctx.Writer.Status()
		route := ctx.FullPath()
		if route == "" {
			route = unmatchedRoute
		}

		logAttributes := []any{
			"request_id", ctx.GetString(RequestIDKey),
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"route", route,
			"status_code", statusCode,
			"duration_ms", durationMilliseconds(duration),
			"client_ip", ctx.ClientIP(),
			"user_agent", ctx.Request.UserAgent(),
			"response_size_bytes", ctx.Writer.Size(),
		}

		switch {
		case statusCode >= http.StatusInternalServerError:
			logger.APIError(
				"HTTP request completed with server error",
				logAttributes...,
			)

		case statusCode >= http.StatusBadRequest:
			logger.APIWarn(
				"HTTP request completed with client error",
				logAttributes...,
			)

		default:
			logger.APIInfo(
				"HTTP request completed",
				logAttributes...,
			)
		}
	}
}

// durationMilliseconds converts a duration into milliseconds while preserving
// fractions of a millisecond.
func durationMilliseconds(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000
}
