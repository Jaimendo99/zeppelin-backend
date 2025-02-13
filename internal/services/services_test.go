package services_test

import (
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

	dbModel := services.RepresetativeInputToDb(&input)

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

	dbModel := services.RepresetativeInputToDb(&input)

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
