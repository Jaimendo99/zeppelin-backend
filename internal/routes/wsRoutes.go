package routes

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/data"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"
)

func DefineWebSocketRoutes(e *echo.Echo, authSvc *services.AuthService) {
	sessionRepo := data.NewSessionRepo(config.DB)
	cm := controller.NewConnectionManager(sessionRepo)

	e.GET("/ws",
		cm.WebSocketHandler(), // <-- Call the method on the instance
		middleware.WsAuthMiddleware(authSvc, "org:student", "org:admin", "org:teacher"),
	)
}
