package routes

import (
	"zeppelin/internal/controller"
	"zeppelin/internal/db"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineRepresentativeRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := db.NewRepresentativeRepo(db.DB)
	reservice := services.NewRepresentativeService(repo)
	recontroller := controller.NewRepresentativeController(reservice)

	e.GET("/representative/:representative_id", recontroller.GetRepresentative())
	e.POST("/representative", recontroller.CreateRepresentative())
	e.GET("/representatives", recontroller.GetAllRepresentatives())
	e.PUT("/representative/:representative_id", recontroller.UpdateRepresentative())
}
