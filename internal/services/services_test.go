package services_test

import (
	"errors"
	"testing"
	"zeppelin/internal/domain"
	"zeppelin/internal/services"

	"github.com/stretchr/testify/assert"
)

func TestRepresetativeInputToDb_FullFields(t *testing.T) {
	input := domain.RepresentativeInput{
		Name:        "Mateo",
		Lastname:    "Mejia",
		Email:       "jaimendo26@gmail.com",
		PhoneNumber: "+129129122",
	}

	dbModel := services.RepresentativesInputToDb(&input)

	assert.Equal(t, input.Name, dbModel.Name, "Name should match")
	assert.Equal(t, input.Lastname, dbModel.Lastname, "Lastname should match")

	assert.True(t, dbModel.Email.Valid, "Email.Valid should be true")
	assert.Equal(t, input.Email, dbModel.Email.String, "Email should match")

	assert.True(t, dbModel.PhoneNumber.Valid, "PhoneNumber.Valid should be true")
	assert.Equal(t, input.PhoneNumber, dbModel.PhoneNumber.String, "PhoneNumber should match")
}

func TestRepresetativeInputToDb_EmptyFields(t *testing.T) {
	input := domain.RepresentativeInput{
		Name:        "Mateo",
		Lastname:    "Mejia",
		Email:       "",
		PhoneNumber: "",
	}

	dbModel := services.RepresentativesInputToDb(&input)

	assert.Equal(t, input.Name, dbModel.Name, "Name should match")
	assert.Equal(t, input.Lastname, dbModel.Lastname, "Lastname should match")

	assert.False(t, dbModel.Email.Valid, "Email.Valid should be false for empty input")
	assert.Equal(t, "", dbModel.Email.String, "Email should be empty")

	assert.False(t, dbModel.PhoneNumber.Valid, "PhoneNumber.Valid should be false for empty input")
	assert.Equal(t, "", dbModel.PhoneNumber.String, "PhoneNumber should be empty")
}

func TestParamToId_Valid(t *testing.T) {
	id, err := services.ParamToId("123")
	assert.NoError(t, err, "Expected no error for valid numeric string")
	assert.Equal(t, 123, id, "Expected id to be 123")
}

func TestParamToId_Invalid(t *testing.T) {
	id, err := services.ParamToId("abc")
	assert.Error(t, err, "Expected error for non-numeric string")
	assert.Equal(t, -1, id, "Expected id to be -1 on error")
}

type MockAuthService struct{}

func (m *MockAuthService) VerifyToken(token string) (*domain.AuthResponse, error) {
	if token == "valid-token" {
		return &domain.AuthResponse{
			AccessToken: "user_1234",
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
	assert.Equal(t, "user_1234", resp.AccessToken, "El AccessToken debe coincidir")
}

func TestVerifyToken_Invalid(t *testing.T) {
	authService := &MockAuthService{}

	resp, err := authService.VerifyToken("invalid-token")
	assert.Error(t, err, "Se esperaba error con un token inválido")
	assert.Nil(t, resp, "La respuesta debe ser nil en caso de error")
	assert.Equal(t, "token inválido o sesión no encontrada", err.Error())
}
