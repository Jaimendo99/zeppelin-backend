package middleware

import (
	"errors"
	"net/http"
	"strings"
	"zeppelin/internal/services"

	"github.com/clerkinc/clerk-sdk-go/clerk"

	"github.com/labstack/echo/v4"
)

func ValidateTokenAndRole(token string, authService *services.AuthService, requiredRoles ...string) (*clerk.TokenClaims, error) {
	if token == "" {
		return nil, errors.New("token requerido")
	}

	claims, err := authService.DecodeToken(token)
	if err != nil {
		return nil, errors.New("token inválido o sesión no encontrada")
	}
	sessionClaims, err := authService.Client.VerifyToken(token)
	if err != nil || sessionClaims == nil {
		return nil, errors.New("token inválido o sesión no encontrada")
	}

	role, err := extractRoleFromClaims(claims)
	if err != nil {
		return nil, errors.New("no se pudo extraer el rol del usuario")
	}

	if len(requiredRoles) > 0 && !contains(requiredRoles, role) {
		return nil, errors.New("acceso denegado: rol no autorizado")
	}

	return claims, nil
}

func RoleMiddleware(authService *services.AuthService, requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token de autorización requerido")
			}

			headerToken := strings.TrimSpace(c.Request().Header.Get("Authorization"))
			token := strings.TrimPrefix(headerToken, "Bearer ")

			claims, err := authService.DecodeToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token inválido o sesión no encontrada")
			}

			sessionClaims, err := authService.Client.VerifyToken(token)
			if err != nil || sessionClaims == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token inválido o sesión no encontrada")
			}

			role, err := extractRoleFromClaims(claims)
			if err != nil {
				return echo.NewHTTPError(http.StatusForbidden, "No se pudo extraer el rol del usuario")
			}

			c.Set("user_role", role)
			c.Set("user_id", claims.Subject)

			if len(requiredRoles) > 0 {
				if contains(requiredRoles, role) {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusForbidden, "Acceso denegado: rol no autorizado")
			}
			return next(c)
		}
	}
}

func contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func extractRoleFromClaims(claims *clerk.TokenClaims) (string, error) {
	if claims.Extra == nil {
		return "", errors.New("no hay información adicional en los claims")
	}

	if role, ok := claims.Extra["role"].(string); ok {
		return role, nil
	}

	return "", errors.New("el rol no está definido en los claims")
}
