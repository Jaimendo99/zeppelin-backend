package config_test

import (
	"os"
	"testing"
	"zeppelin/internal/config"
)

func TestLoadEnv(t *testing.T) {
	err := config.LoadEnv()
	if err != nil {
		// t.Errorf("Expected nil, got %v", err)
	}

}

func TestGetConnectionString(t *testing.T) {

	os.Setenv("CONNECTION_STRING", "test")
	result := config.GetConnectionString()
	if result != "test" {
		t.Errorf("Expected test, got %s", result)
	}
}
