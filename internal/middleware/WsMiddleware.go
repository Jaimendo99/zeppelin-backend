package middleware

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

func WsAuthMiddleware(as domain.AuthServiceI, requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.QueryParam("token")
			if token == "" {
				return controller.ReturnWriteResponse(c, domain.ErrAuthTokenMissing, nil)
			}

			claims, err := as.DecodeToken(token)
			if err != nil {
				return controller.ReturnWriteResponse(c, domain.ErrAuthTokenInvalid, nil)
			}

			sessionClaims, err := as.VerifyToken(token)
			if err != nil || sessionClaims == nil {
				return controller.ReturnWriteResponse(c, domain.ErrAuthTokenInvalid, nil)
			}

			role, err := extractRoleFromClaims(claims)
			if err != nil {
				return controller.ReturnWriteResponse(c, domain.ErrRoleExtractionFailed, nil)
			}

			if len(requiredRoles) > 0 && !contains(requiredRoles, role) {
				return controller.ReturnWriteResponse(c, domain.ErrAuthorizationFailed, nil)
			}

			c.Set("user_id", claims.Subject)
			c.Set("user_role", role)
			c.Set("user_session", claims)
			c.Set("user_session_claims", sessionClaims)

			return next(c)
		}
	}
}
