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
	courseRepo := data.NewCourseRepo(config.DB)
	userRepo := data.NewUserRepo(config.DB)

	Controller := controller.QuizController{
		QuizRepo:              repo,
		AssignmentRepo:        assignmentRepo,
		CourseContentRepo:     courseContentRepo,
		CourseRepo:            courseRepo,
		UserRepo:              userRepo,
		UploadStudentAnswers:  config.UploadJSONToR2,
		GetTeacherQuizContent: config.GetR2Object,
		GeneratePresignedURL:  config.GeneratePresignedURL,
	}

	e.POST("/quiz/submit", Controller.SubmitQuiz(), middleware.RoleMiddleware(authService, "org:student"))
	e.POST("/quiz/review-text-answer", Controller.ReviewTextAnswer(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/quiz/teacher/courses/:courseId", Controller.GetQuizzesByCourse(), middleware.RoleMiddleware(authService, "org:teacher"))
	e.GET("/quiz/student", Controller.GetQuizzesByStudent(), middleware.RoleMiddleware(authService, "org:student", "org:teacher"))
}
