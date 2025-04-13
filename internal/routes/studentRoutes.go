package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineStudentRoutes(e *echo.Echo, authService *services.AuthService, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewUserRepo(config.DB)

	userController := controller.UserController{AuthService: authService, UserRepo: repo}

	e.POST("/student/register", userController.RegisterUser("org:student"), roleMiddlewareProvider("org:admin"))
	e.GET("/student/me", userController.GetUser(), roleMiddlewareProvider("org:student"))
	e.GET("/students", userController.GetAllStudents(), roleMiddlewareProvider("org:admin", "org:teacher"))
}
