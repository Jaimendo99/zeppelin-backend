package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineTeacherRoutes(e *echo.Echo, authService *services.AuthService, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewUserRepo(config.DB)

	userController := controller.UserController{AuthService: authService, UserRepo: repo}

	e.POST("/teacher/register", userController.RegisterUser("org:teacher"), roleMiddlewareProvider("org:admin"))
	e.GET("/teachers", userController.GetAllTeachers(), roleMiddlewareProvider("org:admin", "org:teacher", "org:student"))
}
