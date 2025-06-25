package routes

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
)

func DefineAssignmentRoutes(e *echo.Echo, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewAssignmentRepo(config.DB)
	assignmentController := controller.AssignmentController{Repo: repo}

	e.POST("/assignment", assignmentController.CreateAssignment(), roleMiddlewareProvider("org:student"))
	e.POST("/assignment/verify/:assignment_id", assignmentController.VerifyAssignment(), roleMiddlewareProvider("org:teacher", "org:admin"))
	e.GET("/assignments/student", assignmentController.GetAssignmentsByStudent(), roleMiddlewareProvider("org:student"))
	e.GET("/assignments/teacher/:course_id", assignmentController.GetStudentsByCourse(), roleMiddlewareProvider("org:teacher"))
}
