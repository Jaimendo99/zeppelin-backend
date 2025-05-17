package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineQuizAnswerRoutes(e *echo.Echo, authService *services.AuthService, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewQuizRepository(config.DB)
	assignmentRepo := data.NewAssignmentRepo(config.DB)
	courseContentRepo := data.NewCourseContentRepo(config.DB, controller.GenerateUID)

	Controller := controller.QuizController{
		QuizRepo:          repo,
		AssignmentRepo:    assignmentRepo,
		CourseContentRepo: courseContentRepo,
	}

	e.POST("/quiz/submit", Controller.SubmitQuiz(), middleware.RoleMiddleware(authService, "org:student"))
	e.GET("/quiz/answers", Controller.GetQuizAnswersByStudent(), middleware.RoleMiddleware(authService, "org:student"))
}
