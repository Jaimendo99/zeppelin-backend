package middleware

import (
	"errors"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"strings"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

func ValidateTokenAndRole(token string, authService domain.AuthServiceI, requiredRoles ...string) (*clerk.TokenClaims, error) {

	if token == "" {
		return nil, errors.New("token requerido")
	}

	claims, err := authService.DecodeToken(token)
	if err != nil {
		return nil, errors.New("token inválido o sesión no encontrada")
	}
	sessionClaims, err := authService.VerifyToken(token)
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

func RoleMiddleware(authService domain.AuthServiceI, requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				_ = controller.ReturnWriteResponse(c, domain.ErrAuthTokenMissing, nil)
				return nil
			}

			headerToken := strings.TrimSpace(c.Request().Header.Get("Authorization"))
			token := strings.TrimPrefix(headerToken, "Bearer ")

			claims, err := authService.DecodeToken(token)
			if err != nil {
				return controller.ReturnWriteResponse(c, domain.ErrAuthTokenInvalid, nil)
			}

			sessionClaims, err := authService.VerifyToken(token)
			if err != nil || sessionClaims == nil {
				return controller.ReturnWriteResponse(c, domain.ErrAuthTokenInvalid, nil)
			}

			role, err := extractRoleFromClaims(claims)
			if err != nil {
				return controller.ReturnWriteResponse(c, domain.ErrRoleExtractionFailed, nil)
			}

			c.Set("user_role", role)
			c.Set("user_id", claims.Subject)

			if len(requiredRoles) > 0 {
				if contains(requiredRoles, role) {
					return next(c)
				}
				return controller.ReturnWriteResponse(c, domain.ErrAuthorizationFailed, nil)
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

	return "", domain.ErrRoleExtractionFailed
}
