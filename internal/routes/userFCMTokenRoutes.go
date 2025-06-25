package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineUserFcmTokenRoutes(e *echo.Echo, authService *services.AuthService, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewUserFcmTokenRepo(config.DB)
	userFcmTokenController := controller.UserFcmTokenController{Repo: repo}

	e.POST("/fcm/token", userFcmTokenController.CreateUserFcmToken(), middleware.RoleMiddleware(authService, "org:teacher", "org:student"))
	e.GET("/fcm/tokens", userFcmTokenController.GetUserFcmTokens(), middleware.RoleMiddleware(authService, "org:teacher", "org:student"))
	e.DELETE("/fcm/token", userFcmTokenController.DeleteUserFcmToken(), middleware.RoleMiddleware(authService, "org:teacher", "org:student"))
	e.PATCH("/fcm/token/device-info", userFcmTokenController.UpdateDeviceInfo(), middleware.RoleMiddleware(authService, "org:teacher", "org:student"))
	e.PATCH("/fcm/token/web", userFcmTokenController.UpdateWebToken(), middleware.RoleMiddleware(authService, "org:teacher", "org:student"))
	e.PATCH("/fcm/token/mobile", userFcmTokenController.UpdateMobileToken(), middleware.RoleMiddleware(authService, "org:teacher", "org:student"))
}
