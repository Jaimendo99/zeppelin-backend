package config_test

import (
	"os"
	"testing"
	"zeppelin/internal/config"
)

func TestGetConnectionString(t *testing.T) {

	os.Setenv("CONNECTION_STRING", "test")
	result := config.GetConnectionString()
	if result != "test" {
		t.Errorf("Expected test, got %s", result)
	}
}
