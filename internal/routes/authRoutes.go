package routes

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

func DefineAuthRoutes(e *echo.Echo, clerkClient domain.ClerkInterface) {

	authController := controller.AuthController{Clerk: clerkClient}
	e.GET("tokenFromSession", authController.GetTokenFromSession())

}
