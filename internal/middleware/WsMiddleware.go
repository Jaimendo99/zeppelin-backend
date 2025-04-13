package middleware

import (
	"errors"
	"github.com/labstack/echo/v4"
	"zeppelin/internal/controller"
	"zeppelin/internal/services"
)

func WsAuthMiddleware(as services.AuthService, requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.QueryParam("token")
			if token == "" {
				return controller.ReturnWriteResponse(c, errors.New("token requerido"), nil)
			}

			claims, err := as.DecodeToken(token)
			if err != nil {
				return controller.ReturnWriteResponse(c, errors.New("token inválido o sesión no encontrada"), nil)
			}

			sessionClaims, err := as.VerifyToken(token)
			if err != nil || sessionClaims == nil {
				return controller.ReturnWriteResponse(c, errors.New("token inválido o sesión no encontrada"), nil)
			}

			role, err := extractRoleFromClaims(claims)
			if err != nil {
				return controller.ReturnWriteResponse(c, errors.New("no se pudo extraer el rol del usuario"), nil)
			}

			if len(requiredRoles) > 0 && !contains(requiredRoles, role) {
				return controller.ReturnWriteResponse(c, errors.New("acceso denegado: rol no autorizado"), nil)
			}

			c.Set("user_id", claims.Subject)
			c.Set("user_role", role)
			c.Set("user_session", claims)
			c.Set("user_session_claims", sessionClaims)

			return next(c)
		}
	}
}
