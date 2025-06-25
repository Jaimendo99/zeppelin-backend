package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefinePomodoroRoutes(e *echo.Echo, authService *services.AuthService, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewUserPomodoroRepo(config.DB)
	pomodoroController := controller.PomodoroController{Repo: repo}

	e.GET("/user/pomodoro", pomodoroController.GetPomodoroByUserID(), middleware.RoleMiddleware(authService, "org:student"))
	e.PUT("/user/pomodoro", pomodoroController.UpdatePomodoroByUserID(), middleware.RoleMiddleware(authService, "org:student"))
}
