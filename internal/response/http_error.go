package response

import "github.com/gin-gonic/gin"

// HTTPErrorPayload represents the standard NetPilot API error response.
type HTTPErrorPayload struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteHTTPError aborts the current request and writes a standard JSON error
// response.
func WriteHTTPError(
	ctx *gin.Context,
	statusCode int,
	message string,
	errorCode string,
	requestID string,
) {
	ctx.AbortWithStatusJSON(
		statusCode,
		HTTPErrorPayload{
			Success:   false,
			Message:   message,
			ErrorCode: errorCode,
			RequestID: requestID,
		},
	)
}
