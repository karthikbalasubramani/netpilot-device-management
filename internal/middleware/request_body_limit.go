package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
)

const (
	requestBodyTooLargeMessage = "The request body exceeds the allowed size"
	requestBodyTooLargeCode    = "REQUEST_BODY_TOO_LARGE"

	requestBodyReadFailedMessage = "The request body could not be read"
	requestBodyReadFailedCode    = "INVALID_REQUEST_BODY"
)

// RequestBodyLimit rejects HTTP requests whose body exceeds the configured
// maximum size.
//
// The body is buffered up to the configured limit and then restored so route
// handlers can read it normally.
func RequestBodyLimit(maxBodyBytes int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body == nil ||
			ctx.Request.Body == http.NoBody {
			ctx.Next()
			return
		}

		requestID := ctx.GetString(RequestIDKey)

		// Reject immediately when the client declares a body size that exceeds
		// the configured limit.
		if ctx.Request.ContentLength > maxBodyBytes {
			response.WriteHTTPError(
				ctx,
				http.StatusRequestEntityTooLarge,
				requestBodyTooLargeMessage,
				requestBodyTooLargeCode,
				requestID,
			)

			return
		}

		limitedBody := http.MaxBytesReader(
			ctx.Writer,
			ctx.Request.Body,
			maxBodyBytes,
		)

		body, readErr := io.ReadAll(limitedBody)
		closeErr := limitedBody.Close()

		if readErr != nil {
			var maxBytesError *http.MaxBytesError

			if errors.As(readErr, &maxBytesError) {
				response.WriteHTTPError(
					ctx,
					http.StatusRequestEntityTooLarge,
					requestBodyTooLargeMessage,
					requestBodyTooLargeCode,
					requestID,
				)

				return
			}

			response.WriteHTTPError(
				ctx,
				http.StatusBadRequest,
				requestBodyReadFailedMessage,
				requestBodyReadFailedCode,
				requestID,
			)

			return
		}

		if closeErr != nil {
			response.WriteHTTPError(
				ctx,
				http.StatusBadRequest,
				requestBodyReadFailedMessage,
				requestBodyReadFailedCode,
				requestID,
			)

			return
		}

		// Replace the consumed request body so downstream route handlers can
		// read it normally.
		ctx.Request.Body = io.NopCloser(
			bytes.NewReader(body),
		)
		ctx.Request.ContentLength = int64(len(body))

		ctx.Next()
	}
}
