package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/db"

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
