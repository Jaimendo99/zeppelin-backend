package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"zeppelin/internal/config"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a temporary .env file
func createTempEnvFile(t *testing.T, dir, content string) string {
	t.Helper() // Marks this as a test helper
	tmpFilePath := filepath.Join(dir, ".env")
	err := os.WriteFile(tmpFilePath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create temporary .env file: %v", err)
	}
	return tmpFilePath
}

// Test case for when the .env file exists and loads successfully
func TestLoadEnv_Success(t *testing.T) {
	// 1. Create a temporary directory for the test
	tempDir := t.TempDir() // Automatically cleaned up after the test

	// 2. Create a dummy .env file inside the temporary directory
	envContent := "TEST_VAR=test_value\nANOTHER_VAR=123"
	_ = createTempEnvFile(t, tempDir, envContent) // We don't need the path here

	// 3. Store the original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// 4. Change working directory to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	// 5. IMPORTANT: Ensure we change back to the original directory afterwards
	defer func() {
		err := os.Chdir(originalWd)
		if err != nil {
			// Use t.Errorf for errors in defer as t.Fatalf will stop execution early
			t.Errorf("Failed to change back to original working directory: %v", err)
		}
		// Clean up environment variables set by the temp .env
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("ANOTHER_VAR")
	}()

	// 6. Call the function under test
	loadErr := config.LoadEnv()

	// 7. Assert that no error occurred (this hits the `return nil` line)
	assert.NoError(t, loadErr, "LoadEnv should not return an error when .env exists")

	// 8. Optional: Verify that environment variables were actually loaded
	assert.Equal(t, "test_value", os.Getenv("TEST_VAR"), "TEST_VAR should be loaded from .env")
	assert.Equal(t, "123", os.Getenv("ANOTHER_VAR"), "ANOTHER_VAR should be loaded from .env")
}

// Test case for when the .env file does not exist
func TestLoadEnv_FileNotFound(t *testing.T) {
	// 1. Create a temporary directory (it will be empty)
	tempDir := t.TempDir()

	// 2. Store the original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// 3. Change working directory to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	// 4. IMPORTANT: Ensure we change back to the original directory afterwards
	defer func() {
		err := os.Chdir(originalWd)
		if err != nil {
			t.Errorf("Failed to change back to original working directory: %v", err)
		}
	}()

	// 5. Call the function under test (it should fail as .env doesn't exist here)
	loadErr := config.LoadEnv()

	// 6. Assert that an error *did* occur
	assert.Error(t, loadErr, "LoadEnv should return an error when .env does not exist")

	// 7. Optional: Check if the error message is what we expect
	//    Note: godotenv might return a specific error type or message.
	//          os.IsNotExist(err) might work depending on godotenv's implementation.
	//          Checking the wrapped message is often safer.
	assert.Contains(t, loadErr.Error(), "error loading .env file", "Error message should indicate loading failure")
}

func TestGetConnectionString(t *testing.T) {
	original := os.Getenv("CONNECTION_STRING")
	defer os.Setenv("CONNECTION_STRING", original) // Restore variable after test

	expectedValue := "test-connection-string"
	os.Setenv("CONNECTION_STRING", expectedValue) // Set the test value

	result := config.GetConnectionString()
	assert.Equal(t, expectedValue, result, "GetConnectionString should return the set value")
}

func TestGetConnectionString_Empty(t *testing.T) {
	original := os.Getenv("CONNECTION_STRING")
	defer os.Setenv("CONNECTION_STRING", original) // Restore variable after test

	os.Unsetenv("CONNECTION_STRING") // Remove the variable temporarily
	result := config.GetConnectionString()
	assert.Equal(t, "", result, "GetConnectionString should return an empty string if not defined")
}

// --- Tests for GetMQConnectionString ---

func TestGetMQConnectionString(t *testing.T) {
	original := os.Getenv("MQ_CONN_STRING")
	defer os.Setenv("MQ_CONN_STRING", original)

	expectedValue := "test-mq-connection-string"
	os.Setenv("MQ_CONN_STRING", expectedValue)

	result := config.GetMQConnectionString()
	assert.Equal(t, expectedValue, result, "GetMQConnectionString should return the set value")
}

func TestGetMQConnectionString_Empty(t *testing.T) {
	original := os.Getenv("MQ_CONN_STRING")
	defer os.Setenv("MQ_CONN_STRING", original)

	os.Unsetenv("MQ_CONN_STRING")
	result := config.GetMQConnectionString()
	assert.Equal(t, "", result, "GetMQConnectionString should return an empty string if not defined")
}

// --- Tests for GetSmtpPassword ---

func TestGetSmtpPassword(t *testing.T) {
	original := os.Getenv("SMTP_PASSWORD")
	defer os.Setenv("SMTP_PASSWORD", original)

	expectedValue := "test-smtp-password"
	os.Setenv("SMTP_PASSWORD", expectedValue)

	result := config.GetSmtpPassword()
	assert.Equal(t, expectedValue, result, "GetSmtpPassword should return the set value")
}

func TestGetSmtpPassword_Empty(t *testing.T) {
	original := os.Getenv("SMTP_PASSWORD")
	defer os.Setenv("SMTP_PASSWORD", original)

	os.Unsetenv("SMTP_PASSWORD")
	result := config.GetSmtpPassword()
	assert.Equal(t, "", result, "GetSmtpPassword should return an empty string if not defined")
}

// --- Tests for GetFirebaseConn ---

func TestGetFirebaseConn(t *testing.T) {
	original := os.Getenv("FIREBASE_CONN")
	defer os.Setenv("FIREBASE_CONN", original)

	expectedValue := "test-firebase-conn"
	os.Setenv("FIREBASE_CONN", expectedValue)

	result := config.GetFirebaseConn()
	assert.Equal(t, expectedValue, result, "GetFirebaseConn should return the set value")
}

func TestGetFirebaseConn_Empty(t *testing.T) {
	original := os.Getenv("FIREBASE_CONN")
	defer os.Setenv("FIREBASE_CONN", original)

	os.Unsetenv("FIREBASE_CONN")
	result := config.GetFirebaseConn()
	assert.Equal(t, "", result, "GetFirebaseConn should return an empty string if not defined")
}

// --- Tests for GetClerkConfig ---

func TestGetClerkConfig(t *testing.T) {
	original := os.Getenv("CLERK_API_KEY")
	defer os.Setenv("CLERK_API_KEY", original)

	expectedValue := "test-clerk-api-key"
	os.Setenv("CLERK_API_KEY", expectedValue)

	result := config.GetClerkConfig()
	assert.Equal(t, expectedValue, result, "GetClerkConfig should return the set value")
}

// Keep your existing empty test for Clerk Config
func TestGetClerkConfig_Empty(t *testing.T) {
	original := os.Getenv("CLERK_API_KEY")
	defer os.Setenv("CLERK_API_KEY", original) // Restore variable after test

	os.Unsetenv("CLERK_API_KEY") // Remove the variable temporarily
	result := config.GetClerkConfig()
	assert.Equal(t, "", result, "GetClerkConfig should return an empty string if not defined")
}
