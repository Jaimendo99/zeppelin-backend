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

func TestGetMQConnectionString(t *testing.T) {
	os.Setenv("MQ_CONN_STRING", "test")
	result := config.GetMQConnectionString()
	if result != "test" {
		t.Errorf("Expected test, got %s", result)
	}
}

func TestGetSmtpPassword(t *testing.T) {
	os.Setenv("SMTP_PASSWORD", "test")
	result := config.GetSmtpPassword()
	if result != "test" {
		t.Errorf("Expected test, got %s", result)
	}
}

func TestGetFCMConnection(t *testing.T) {
	os.Setenv("FIREBASE_CONN", "test")
	result := config.GetFirebaseConn()
	if result != "test" {
		t.Errorf("Expected test, got %s", result)
	}
}
