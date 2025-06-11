package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineRepresentativeRoutes(e *echo.Echo, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc) {
	repo := data.NewRepresentativeRepo(config.DB)

	recontroller := controller.RepresentativeController{Repo: repo}

	e.GET("/representative/:representative_id", recontroller.GetRepresentative(), roleMiddlewareProvider("org:admin", "org:teacher"))
	e.GET("/representatives", recontroller.GetAllRepresentatives())
	e.PUT("/representative/:representative_id", recontroller.UpdateRepresentative(), roleMiddlewareProvider("org:admin", "org:teacher"))
}

func DefineNotificationRoutes(e *echo.Echo, roleMiddlewareProvider func(roles ...string) echo.MiddlewareFunc, m ...echo.MiddlewareFunc) {

	smtServer := config.GetSmtpConfig()
	fcmClient := config.GetFCMClient()
	db := config.DB

	ns := []domain.NotificationService{
		services.NewEmailNotification(*smtServer),
		services.NewPushNotification(*fcmClient),
	}

	queueServer := data.NewRabbitMQImpl(config.ProducerChannel)

	repo := data.NewNotificationRepo(db, queueServer, ns)
	nc := controller.NewNotificationController(repo)

	middlewares := []echo.MiddlewareFunc{
		roleMiddlewareProvider("org:teacher", "org:admin"),
	}
	middlewares = append(middlewares, m...)
	e.POST("/notification", nc.SendNotification(), middlewares...)

	go func() {
		if err := repo.ConsumeFromQueue("notification"); err != nil {
			e.Logger.Error("error consuming queue: ", err)
		}
	}()
}
