package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

type updateUserRoleRequest struct {
	Role string `json:"role"`
}

func updateUserRoleHandler(
	userService *user.Service,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authenticatedUser, ok :=
			middleware.GetAuthenticatedUser(ctx)

		if !ok {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"success": false,

					"message": "Authenticated user context is unavailable",

					"error_code": "AUTHENTICATION_CONTEXT_MISSING",

					"request_id": ctx.GetString(
						middleware.RequestIDKey,
					),
				},
			)

			return
		}

		targetUserID :=
			ctx.Param("user_id")

		var request updateUserRoleRequest

		if err := ctx.ShouldBindJSON(
			&request,
		); err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"success": false,

					"message": "Invalid user role request",

					"error_code": "INVALID_USER_ROLE_REQUEST",

					"request_id": ctx.GetString(
						middleware.RequestIDKey,
					),
				},
			)

			return
		}

		role, err :=
			user.ParseRole(
				request.Role,
			)
		if err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"success": false,

					"message": "Role must be admin, operator, or viewer",

					"error_code": "INVALID_USER_ROLE",

					"request_id": ctx.GetString(
						middleware.RequestIDKey,
					),
				},
			)

			return
		}

		updatedUser, err :=
			userService.UpdateRole(
				ctx.Request.Context(),

				authenticatedUser.UserID,

				targetUserID,

				role,
			)

		if err != nil {
			switch {
			case errors.Is(
				err,
				user.ErrSelfRoleChange,
			):
				ctx.JSON(
					http.StatusForbidden,
					gin.H{
						"success": false,

						"message": "Administrators cannot change their own role",

						"error_code": "SELF_ROLE_CHANGE_FORBIDDEN",

						"request_id": ctx.GetString(
							middleware.RequestIDKey,
						),
					},
				)

			case errors.Is(
				err,
				user.ErrNotFound,
			):
				ctx.JSON(
					http.StatusNotFound,
					gin.H{
						"success": false,

						"message": "User not found",

						"error_code": "USER_NOT_FOUND",

						"request_id": ctx.GetString(
							middleware.RequestIDKey,
						),
					},
				)

			case errors.Is(
				err,
				user.ErrInvalidRole,
			):
				ctx.JSON(
					http.StatusBadRequest,
					gin.H{
						"success": false,

						"message": "Invalid user role",

						"error_code": "INVALID_USER_ROLE",

						"request_id": ctx.GetString(
							middleware.RequestIDKey,
						),
					},
				)

			default:
				ctx.JSON(
					http.StatusInternalServerError,
					gin.H{
						"success": false,

						"message": "Failed to update user role",

						"error_code": "USER_ROLE_UPDATE_FAILED",

						"request_id": ctx.GetString(
							middleware.RequestIDKey,
						),
					},
				)
			}

			return
		}

		ctx.JSON(
			http.StatusOK,
			gin.H{
				"success": true,

				"message": "User role updated successfully",

				"data": gin.H{
					"user_id": updatedUser.UserID,

					"name": updatedUser.Name,

					"email": updatedUser.Email,

					"role": updatedUser.Role,

					"status": updatedUser.Status,

					"updated_at": updatedUser.UpdatedAt,
				},
			},
		)
	}
}
