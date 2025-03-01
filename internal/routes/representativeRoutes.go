package routes

import (
	"zeppelin/internal/controller"
	"zeppelin/internal/db"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineRepresentativeRoutes(e *echo.Echo, m ...echo.MiddlewareFunc) {
	repo := db.NewRepresentativeRepo(db.DB)

	recontroller := controller.RepresentativeController{Repo: repo}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	e.GET("/representative/:representative_id", recontroller.GetRepresentative(), middleware.AuthMiddleware(authService))
	e.POST("/representative", recontroller.CreateRepresentative(), middleware.AuthMiddleware(authService))
	e.GET("/representatives", recontroller.GetAllRepresentatives(), middleware.AuthMiddleware(authService))
	e.PUT("/representative/:representative_id", recontroller.UpdateRepresentative(), middleware.AuthMiddleware(authService))
}
