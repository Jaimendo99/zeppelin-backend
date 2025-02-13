package routes

import (
	"zeppelin/internal/controller"
	"zeppelin/internal/db"

	"github.com/labstack/echo/v4"
)

func DefineRepresentativeRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := db.NewRepresentativeRepo(db.DB)

	recontroller := controller.RepresentativeController{Repo: repo}

	e.GET("/representative/:representative_id", recontroller.GetRepresentative())
	e.POST("/representative", recontroller.CreateRepresentative())
	e.GET("/representatives", recontroller.GetAllRepresentatives())
	e.PUT("/representative/:representative_id", recontroller.UpdateRepresentative())
}
