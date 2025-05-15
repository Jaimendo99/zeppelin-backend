package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineCourseContentRoutes(e *echo.Echo) {
	repo := data.NewCourseContentRepo(config.DB, controller.GenerateUID)
	assignmentRepo := data.NewAssignmentRepo(config.DB)
	courseRepo := data.NewCourseRepo(config.DB)
	controller := controller.CourseContentController{
		Repo:                 repo,
		RepoAssigment:        assignmentRepo,
		RepoCourse:           courseRepo,
		GeneratePresignedURL: config.GeneratePresignedURL,
	}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	// GET routes
	e.GET("/course-content", controller.GetCourseContentTeacher(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/course-content/student", controller.GetCourseContentForStudent(), middleware.RoleMiddleware(authService, "org:student"))

	// POST routes
	e.POST("/course-content/module", controller.AddModule(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.POST("/course-content/section", controller.AddSection(), middleware.RoleMiddleware(authService, "org:teacher"))

	// PUT routes
	e.PUT("/course-content", controller.UpdateContent(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/status", controller.UpdateContentStatus(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/module-title", controller.UpdateModuleTitle(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/in_progress", controller.UpdateUserContentStatus(2), middleware.RoleMiddleware(authService, "org:student"))
	e.PUT("/course-content/completed", controller.UpdateUserContentStatus(3), middleware.RoleMiddleware(authService, "org:student"))
}
