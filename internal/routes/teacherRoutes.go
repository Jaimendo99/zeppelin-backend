package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineTeacherRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := data.NewUserRepo(config.DB)

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	userController := controller.UserController{AuthService: authService, UserRepo: repo}

	e.POST("/teacher/register", userController.RegisterUser("org:teacher"), middleware.RoleMiddleware(authService, "org:admin"))
	e.GET("/teacher/me", userController.GetUser(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/teachers", userController.GetAllTeachers(), middleware.RoleMiddleware(authService, "org:admin", "org:teacher", "org:student"))
}
