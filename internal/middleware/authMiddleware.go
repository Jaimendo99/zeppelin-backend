package middleware

import (
	"net/http"
	"strings"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token de autorización requerido")
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Formato de token inválido")
			}
			token := tokenParts[1]

			claims, err := authService.VerifyToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token inválido o sesión no encontrada")
			}

			c.Set("user", claims)

			return next(c)
		}
	}
}
