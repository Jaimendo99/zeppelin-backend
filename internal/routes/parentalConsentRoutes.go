package routes

import (
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineParentalConsentRoutes(e *echo.Echo) {
	repo := data.NewParentalConsentRepo(config.DB)
	consentController := controller.ParentalConsentController{
		Repo: repo,
	}

	authService, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error inicializando AuthService: ", err)
		return
	}

	e.GET("/parental-consent", consentController.GetConsentByToken())
	e.PUT("/parental-consent/status", consentController.UpdateConsentStatus())
	e.GET("/parental-consent/user", consentController.GetConsentByUserID(), middleware.RoleMiddleware(authService, "org:student"))
}
