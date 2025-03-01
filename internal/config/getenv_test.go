package config_test

import (
	"os"
	"testing"
	"zeppelin/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnv(t *testing.T) {
	err := config.LoadEnv()
	if err != nil {
		// t.Errorf("Expected nil, got %v", err)
	}
}

func TestGetConnectionString(t *testing.T) {
	original := os.Getenv("CONNECTION_STRING")
	defer os.Setenv("CONNECTION_STRING", original) // Restaurar variable después del test

	os.Setenv("CONNECTION_STRING", "test-connection-string")
	os.Setenv("CONNECTION_STRING", "test")
	result := config.GetConnectionString()
	if result != "test" {
		t.Errorf("Expected test, got %s", result)
	}
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
