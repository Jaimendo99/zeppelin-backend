package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/db"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineRepresentativeRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := db.NewRepresentativeRepo(config.DB)

	recontroller := controller.RepresentativeController{Repo: repo}

	e.GET("/representative/:representative_id", recontroller.GetRepresentative(), m...)
	e.POST("/representative", recontroller.CreateRepresentative(), m...)
	e.GET("/representatives", recontroller.GetAllRepresentatives(), m...)
	e.PUT("/representative/:representative_id", recontroller.UpdateRepresentative(), m...)
}

func DefineNotificationRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	service := services.NotificationPrinter{}
	repo := db.NewNotificationMq(config.ProducerChannel, service)
	controller := controller.NewNotificationController(repo)
	e.POST("/notification", controller.SendNotification(), m...)
}
