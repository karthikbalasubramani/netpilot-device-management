package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
}

// Func to Construct and send common Success response
func SuccessResponse(ctx *gin.Context, statusCode int, message string, data any) {
	ctx.JSON(
		statusCode, Body{
			Success: true,
			Message: message,
			Data:    data,
		})
}

// Func to Construct and send common Error response
func ErrorResponse(ctx *gin.Context, statusCode int, message string, err error) {
	errMessage := http.StatusText(statusCode)
	if err != nil {
		errMessage = err.Error()
	}
	ctx.JSON(
		statusCode, Body{
			Success: false,
			Message: message,
			Error:   errMessage,
		})
}
