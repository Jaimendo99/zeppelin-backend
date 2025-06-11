package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineCourseRoutes(e *echo.Echo, authService *services.AuthService, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewCourseRepo(config.DB)
	courseController := controller.CourseController{Repo: repo}

	e.POST("/course", courseController.CreateCourse(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/courses/teacher", courseController.GetCoursesByTeacher(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/courses/student", courseController.GetCoursesByStudent(), middleware.RoleMiddleware(authService, "org:student"))
	e.GET("v2/courses/student", courseController.GetCoursesByStudent2(), middleware.RoleMiddleware(authService, "org:student"))
	e.GET("/course/:course_id", courseController.GetCourse(), middleware.RoleMiddleware(authService, "org:student"))
}
