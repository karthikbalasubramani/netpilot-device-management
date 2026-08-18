package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/auth"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/response"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

const (
	invalidRegistrationMessage = "Invalid user registration request"
	invalidRegistrationCode    = "INVALID_REGISTRATION_REQUEST"

	userAlreadyExistsMessage = "An account with the provided details already exists"
	userAlreadyExistsCode    = "USER_ALREADY_EXISTS"

	registrationFailedMessage = "User registration could not be completed"
	registrationFailedCode    = "USER_REGISTRATION_FAILED"

	invalidLoginRequestMessage = "Invalid login request"
	invalidLoginRequestCode    = "INVALID_LOGIN_REQUEST"

	invalidCredentialsMessage = "Invalid email or password"
	invalidCredentialsCode    = "INVALID_CREDENTIALS"

	accountUnavailableMessage = "User account is not available for login"
	accountUnavailableCode    = "ACCOUNT_UNAVAILABLE"

	loginFailedMessage = "Login could not be completed"
	loginFailedCode    = "LOGIN_FAILED"
)

// AuthService defines the authentication operations required by the HTTP
// server.
type AuthService interface {
	Register(
		ctx context.Context,
		request auth.RegisterRequest,
	) (*auth.RegisterResponse, error)

	Login(
		ctx context.Context,
		request auth.LoginRequest,
	) (*auth.LoginResponse, error)
}

// registerAuthRoutes registers version 1 authentication routes.
func (server *Server) registerAuthRoutes(
	apiV1 *gin.RouterGroup,
) {
	authRoutes := apiV1.Group("/auth")

	authRoutes.POST(
		"/register",
		server.registerUser,
	)

	authRoutes.POST(
		"/login",
		server.loginUser,
	)

	logger.Debug(
		"Authentication routes registered successfully",
		"base_path",
		ApiV1basepath+"/auth",
	)
}

// registerUser handles new NetPilot user registration.
func (server *Server) registerUser(
	ctx *gin.Context,
) {
	requestID := ctx.GetString(
		middleware.RequestIDKey,
	)

	var request auth.RegisterRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.Warn(
			"Invalid user registration request body",
			"request_id", requestID,
			"error", err,
		)

		response.WriteHTTPError(
			ctx,
			http.StatusBadRequest,
			invalidRegistrationMessage,
			invalidRegistrationCode,
			requestID,
		)

		return
	}

	registeredUser, err := server.authService.Register(
		ctx.Request.Context(),
		request,
	)
	if err != nil {
		switch {
		case isRegistrationValidationError(err):
			response.WriteHTTPError(
				ctx,
				http.StatusBadRequest,
				err.Error(),
				invalidRegistrationCode,
				requestID,
			)

		case errors.Is(
			err,
			user.ErrAlreadyExists,
		):
			response.WriteHTTPError(
				ctx,
				http.StatusConflict,
				userAlreadyExistsMessage,
				userAlreadyExistsCode,
				requestID,
			)

		default:
			logger.Error(
				"User registration failed",
				"request_id", requestID,
				"error", err,
			)

			response.WriteHTTPError(
				ctx,
				http.StatusInternalServerError,
				registrationFailedMessage,
				registrationFailedCode,
				requestID,
			)
		}

		return
	}

	response.SuccessResponse(
		ctx,
		http.StatusCreated,
		"User registered successfully",
		registeredUser,
	)
}

// loginUser handles NetPilot user authentication.
func (server *Server) loginUser(
	ctx *gin.Context,
) {
	requestID := ctx.GetString(
		middleware.RequestIDKey,
	)

	var request auth.LoginRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.Warn(
			"Invalid login request body",
			"request_id", requestID,
			"error", err,
		)

		response.WriteHTTPError(
			ctx,
			http.StatusBadRequest,
			invalidLoginRequestMessage,
			invalidLoginRequestCode,
			requestID,
		)

		return
	}

	loginResponse, err := server.authService.Login(
		ctx.Request.Context(),
		request,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			auth.ErrInvalidCredentials,
		):
			response.WriteHTTPError(
				ctx,
				http.StatusUnauthorized,
				invalidCredentialsMessage,
				invalidCredentialsCode,
				requestID,
			)

		case errors.Is(
			err,
			auth.ErrAccountUnavailable,
		):
			response.WriteHTTPError(
				ctx,
				http.StatusForbidden,
				accountUnavailableMessage,
				accountUnavailableCode,
				requestID,
			)

		default:
			logger.Error(
				"User login failed",
				"request_id", requestID,
				"error", err,
			)

			response.WriteHTTPError(
				ctx,
				http.StatusInternalServerError,
				loginFailedMessage,
				loginFailedCode,
				requestID,
			)
		}

		return
	}

	response.SuccessResponse(
		ctx,
		http.StatusOK,
		"Login successful",
		loginResponse,
	)
}

func isRegistrationValidationError(
	err error,
) bool {
	return errors.Is(err, auth.ErrInvalidName) ||
		errors.Is(err, auth.ErrInvalidEmail) ||
		errors.Is(err, auth.ErrPasswordTooShort) ||
		errors.Is(err, auth.ErrPasswordTooLong)
}
