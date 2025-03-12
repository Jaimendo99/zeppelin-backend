package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineRepresentativeRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := data.NewRepresentativeRepo(config.DB)

	recontroller := controller.RepresentativeController{Repo: repo}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	e.GET("/representative/:representative_id", recontroller.GetRepresentative(), middleware.RoleMiddleware(authService, "org:admin", "org:teacher"))
	e.POST("/representative", recontroller.CreateRepresentative(), middleware.RoleMiddleware(authService, "org:admin", "org:teacher"))
	e.GET("/representatives", recontroller.GetAllRepresentatives(), middleware.RoleMiddleware(authService, "org:admin", "org:teacher"))
	e.PUT("/representative/:representative_id", recontroller.UpdateRepresentative(), middleware.RoleMiddleware(authService, "org:admin", "org:teacher"))
}

func DefineNotificationRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {

	smtServer := config.GetSmtpConfig()
	fcmClient := config.GetFCMClient()
	db := config.DB

	services := []domain.NotificationService{
		services.NewEmailNotification(*smtServer),
		services.NewPushNotification(*fcmClient),
	}

	queueServer := data.NewRabbitMQImpl(config.ProducerChannel)

	repo := data.NewNotificationRepo(db, queueServer, services)
	controller := controller.NewNotificationController(repo)

	e.POST("/notification", controller.SendNotification(), m...)

	go func() {
		if err := repo.ConsumeFromQueue("notification"); err != nil {
			e.Logger.Error("error consuming queue: ", err)
		}
	}()
}
