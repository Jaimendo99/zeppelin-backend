package routes

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/controller"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"
)

func DefineWebSocketRoutes1(e *echo.Echo, authSvc *services.AuthService) {
	cm := controller.NewConnectionManager()
	e.GET("/ws",
		cm.WebSocketHandler(), // <-- Call the method on the instance
		middleware.WsAuthMiddleware(authSvc, "org:student", "org:admin", "org:teacher"),
	)
}
