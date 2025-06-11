package routes

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"
)

func DefineStudentRoutes(e *echo.Echo) {
	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	roleMiddlewareProvider := func(roles ...string) echo.MiddlewareFunc {
		return middleware.RoleMiddleware(authService, roles...)
	}

	userRepo := data.NewUserRepo(config.DB)
	consentRepo := data.NewParentalConsentRepo(config.DB)
	repRepo := data.NewRepresentativeRepo(config.DB)

	resendService, err := config.InitResend()
	if err != nil {
		e.Logger.Fatal("Error inicializando ResendService: ", err)
		return
	}

	userController := controller.UserController{
		AuthService:   authService,
		UserRepo:      userRepo,
		ConsentRepo:   consentRepo,
		RepRepo:       repRepo,
		SendEmailFunc: resendService.SendParentalConsentEmail,
	}

	e.POST("/student/register", userController.RegisterUser("org:student"), roleMiddlewareProvider("org:admin"))
	e.GET("/student/me", userController.GetUser(), roleMiddlewareProvider("org:student"))
	e.GET("/students", userController.GetAllStudents(), roleMiddlewareProvider("org:admin", "org:teacher"))
}
