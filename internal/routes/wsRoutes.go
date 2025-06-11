package routes

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"
)

// DefineWebSocketRoutes registra la ruta /ws y pasa al ConnectionManager
// tanto el repo de sesiones como el de consentimientos parentales.
func DefineWebSocketRoutes(e *echo.Echo, authSvc *services.AuthService) {
	// Repositorios de datos
	sessionRepo := data.NewSessionRepo(config.DB)
	consentRepo := data.NewParentalConsentRepo(config.DB)

	// Creamos el ConnectionManager con ambos repos
	cm := controller.NewConnectionManager(sessionRepo, consentRepo)

	// Ruta WebSocket
	e.GET("/ws",
		cm.WebSocketHandler(),
		middleware.WsAuthMiddleware(authSvc, "org:student", "org:admin", "org:teacher"),
	)
}
