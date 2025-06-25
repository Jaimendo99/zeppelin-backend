package domain_test

import (
	"errors"
	"testing"
	"zeppelin/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestRepresentativeGetTableName(t *testing.T) {
	name := domain.RepresentativeDb{}.TableName()
	assert.Equal(t, "representatives", name)
}

func TestRepresentativeInputGetTableName(t *testing.T) {
	name := domain.RepresentativeInput{}.TableName()
	assert.Equal(t, "representatives", name)
}

type MockAuthService struct{}

func (m *MockAuthService) VerifyToken(token string) (*domain.AuthResponse, error) {
	if token == "valid-token" {
		return &domain.AuthResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}, nil
	}
	return nil, errors.New("token inválido o sesión no encontrada")
}

func TestVerifyToken_Valid(t *testing.T) {
	authService := &MockAuthService{}

	resp, err := authService.VerifyToken("valid-token")
	assert.NoError(t, err, "No se esperaba error con un token válido")
	assert.NotNil(t, resp, "La respuesta no debe ser nil")
	assert.Equal(t, "test-access-token", resp.AccessToken, "El AccessToken debe coincidir")
	assert.Equal(t, "Bearer", resp.TokenType, "El TokenType debe ser 'Bearer'")
	assert.Equal(t, 3600, resp.ExpiresIn, "ExpiresIn debe ser 3600")
}

func TestVerifyToken_Invalid(t *testing.T) {
	authService := &MockAuthService{}

	resp, err := authService.VerifyToken("invalid-token")
	assert.Error(t, err, "Se esperaba error con un token inválido")
	assert.Nil(t, resp, "La respuesta debe ser nil en caso de error")
	assert.Equal(t, "token inválido o sesión no encontrada", err.Error())
}
