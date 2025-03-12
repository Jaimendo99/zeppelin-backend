package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineCourseRoutes(e *echo.Echo) {
	repo := data.NewCourseRepo(config.DB)
	courseController := controller.CourseController{Repo: repo}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	e.POST("/course", courseController.CreateCourse(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/courses/teacher", courseController.GetCoursesByTeacher(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/courses/student", courseController.GetCoursesByStudent(), middleware.RoleMiddleware(authService, "org:student"))
}
