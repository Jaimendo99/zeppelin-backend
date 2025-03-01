package config_test

import (
	"os"
	"testing"
	"zeppelin/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestGetConnectionString(t *testing.T) {
	original := os.Getenv("CONNECTION_STRING")
	defer os.Setenv("CONNECTION_STRING", original) // Restaurar variable después del test

	os.Setenv("CONNECTION_STRING", "test-connection-string")
	result := config.GetConnectionString()
	assert.Equal(t, "test-connection-string", result, "GetConnectionString debe devolver el valor correcto")
}

func TestGetClerkConfig(t *testing.T) {
	original := os.Getenv("CLERK_API_KEY")
	defer os.Setenv("CLERK_API_KEY", original) // Restaurar variable después del test

	os.Setenv("CLERK_API_KEY", "test-clerk-api-key")
	result := config.GetClerkConfig()
	assert.Equal(t, "test-clerk-api-key", result, "GetClerkConfig debe devolver el valor correcto")
}

func TestGetClerkConfig_Empty(t *testing.T) {
	original := os.Getenv("CLERK_API_KEY")
	defer os.Setenv("CLERK_API_KEY", original) // Restaurar variable después del test

	os.Unsetenv("CLERK_API_KEY") // Eliminar la variable temporalmente
	result := config.GetClerkConfig()
	assert.Equal(t, "", result, "GetClerkConfig debe devolver una cadena vacía si no está definida")
}

func TestGetConnectionString_Empty(t *testing.T) {
	original := os.Getenv("CONNECTION_STRING")
	defer os.Setenv("CONNECTION_STRING", original) // Restaurar variable después del test

	os.Unsetenv("CONNECTION_STRING") // Eliminar la variable temporalmente
	result := config.GetConnectionString()
	assert.Equal(t, "", result, "GetConnectionString debe devolver una cadena vacía si no está definida")
}
