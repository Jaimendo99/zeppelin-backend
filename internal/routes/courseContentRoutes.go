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
		Repo:          repo,
		RepoAssigment: assignmentRepo,
		RepoCourse:    courseRepo,
	}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	e.GET("/course-content", controller.GetCourseContentTeacher(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/course-content/student", controller.GetCourseContentForStudent(), middleware.RoleMiddleware(authService, "org:student"))
	e.POST("/course-content/section/video", controller.AddVideoSection(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.POST("/course-content/section/quiz", controller.AddQuizSection(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.POST("/course-content/section/text", controller.AddTextSection(), middleware.RoleMiddleware(authService, "org:teacher"))

	e.PUT("/course-content/video", controller.UpdateVideoContent(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/quiz", controller.UpdateQuizContent(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/text", controller.UpdateTextContent(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/status", controller.UpdateContentStatus(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.PUT("/course-content/module-title", controller.UpdateModuleTitle(), middleware.RoleMiddleware(authService, "org:teacher"))

}
