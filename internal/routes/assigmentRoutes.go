package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineAssignmentRoutes(e *echo.Echo) {
	repo := data.NewAssignmentRepo(config.DB)
	assignmentController := controller.AssignmentController{Repo: repo}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	e.POST("/assignment", assignmentController.CreateAssignment(), middleware.RoleMiddleware(authService, "org:student"))
	e.POST("/assignment/verify/:assignment_id", assignmentController.VerifyAssignment(), middleware.RoleMiddleware(authService, "org:teacher", "org:admin"))
	e.GET("/assignments/student", assignmentController.GetAssignmentsByStudent(), middleware.RoleMiddleware(authService, "org:student"))
	e.GET("/assignments/teacher/:course_id", assignmentController.GetStudentsByCourse(), middleware.RoleMiddleware(authService, "org:teacher"))
}
