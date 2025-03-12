package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineStudentRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := data.NewUserRepo(config.DB)

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	userController := controller.UserController{AuthService: authService, UserRepo: repo}

	e.POST("/student/register", userController.RegisterUser("org:student"), middleware.RoleMiddleware(authService, "org:admin"))
	e.GET("/student/me", userController.GetUser(), middleware.RoleMiddleware(authService, "org:student"))
	e.GET("/students", userController.GetAllStudents(), middleware.RoleMiddleware(authService, "org:admin", "org:teacher"))
}
