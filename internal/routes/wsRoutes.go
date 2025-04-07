package routes

import (
	"zeppelin/internal/controller"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func DefineWebSocketRoutes(e *echo.Echo) {
	authService, _ := services.NewAuthService()
	e.GET("/ws", controller.WebSocketHandler(authService))
}
