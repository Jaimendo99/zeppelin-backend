package services_test

import (
	"context"
	"errors"
	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
	"net/smtp"
	"strings"
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

// Mock implementation for smtp.SendMail
type mockSmtpSender struct {
	called    bool
	addr      string
	auth      smtp.Auth
	from      string
	to        []string
	msg       []byte
	returnErr error
}

func (m *mockSmtpSender) send(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.called = true
	m.addr = addr
	m.auth = a // Note: comparing smtp.Auth implementations can be tricky
	m.from = from
	m.to = to
	m.msg = msg
	return m.returnErr
}

func TestEmailNotification_SendNotification(t *testing.T) {
	// Store original function and restore it after the test
	originalSmtpSendMail := services.SmtpSendMail
	t.Cleanup(func() {
		services.SmtpSendMail = originalSmtpSendMail
	})

	// --- Test Setup ---
	testConfig := domain.SmtpConfig{
		Host:     "smtp.test.com",
		Port:     "587",
		Username: "user@test.com",
		Auth:     smtp.PlainAuth("", "user@test.com", "password", "smtp.test.com"), // Example Auth
	}
	emailService := services.NewEmailNotification(testConfig)

	testNotification := domain.NotificationData{
		Address: []string{"receiver1@example.com", "receiver2@example.com"},
		Title:   "Test Title", // Title isn't used in email subject currently
		Message: "This is the test message body.",
	}

	// --- Test Case: Success ---
	t.Run("Success", func(t *testing.T) {
		mockSender := &mockSmtpSender{}
		services.SmtpSendMail = mockSender.send // Replace with mock

		err := emailService.SendNotification(testNotification)

		assert.NoError(t, err)
		assert.True(t, mockSender.called, "smtp.SendMail should have been called")

		// Assert arguments passed to the mock
		expectedAddr := "smtp.test.com:587"
		assert.Equal(t, expectedAddr, mockSender.addr)
		assert.Equal(t, testConfig.Username, mockSender.from)
		assert.Equal(t, testNotification.Address, mockSender.to)
		// You might want more specific checks for Auth if needed
		assert.ObjectsAreEqual(testConfig.Auth, mockSender.auth)

		// Check message content (basic check)
		expectedMsgPart := "Subject: Notification\r\n\r\n" + testNotification.Message
		assert.Contains(t, string(mockSender.msg), expectedMsgPart)
		expectedToHeader := "To: " + strings.Join(testNotification.Address, ",")
		assert.Contains(t, string(mockSender.msg), expectedToHeader)
	})

	// --- Test Case: Failure ---
	t.Run("Failure", func(t *testing.T) {
		mockSender := &mockSmtpSender{
			returnErr: errors.New("SMTP server connection failed"),
		}
		services.SmtpSendMail = mockSender.send // Replace with mock

		err := emailService.SendNotification(testNotification)

		assert.Error(t, err)
		assert.Equal(t, "SMTP server connection failed", err.Error())
		assert.True(t, mockSender.called, "smtp.SendMail should have been called")
		// Argument checks are still relevant even on failure
		expectedAddr := "smtp.test.com:587"
		assert.Equal(t, expectedAddr, mockSender.addr)
		assert.Equal(t, testConfig.Username, mockSender.from)
		assert.Equal(t, testNotification.Address, mockSender.to)
	})
}

// --- Mock FirebaseMessenger (Keep as is) ---
type MockFirebaseMessenger struct {
	mock.Mock
}

func (m *MockFirebaseMessenger) SendEach(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error) {
	args := m.Called(ctx, messages)
	resp := args.Get(0)
	if resp == nil && args.Error(1) == nil {
		return nil, args.Error(1)
	}
	batchResp, _ := resp.(*messaging.BatchResponse)
	return batchResp, args.Error(1)
}

// --- Tests ---

func TestPushNotification_SendNotification(t *testing.T) {
	// --- Common Test Data (can stay in outer scope) ---
	testNotification := domain.NotificationData{
		Address: []string{"token1", "token2"},
		Title:   "Push Title",
		Message: "Push message body.",
	}

	// Helper function or direct construction to build expected messages
	buildExpectedMessages := func(n domain.NotificationData) []*messaging.Message {
		var expected []*messaging.Message
		for _, token := range n.Address {
			expected = append(expected, &messaging.Message{
				Android: &messaging.AndroidConfig{
					Priority: "high",
					Notification: &messaging.AndroidNotification{
						Title: n.Title,
						Body:  n.Message,
					},
				},
				Token: token,
			})
		}
		return expected
	}
	expectedMessages := buildExpectedMessages(testNotification)

	// --- Test Case: Success ---
	t.Run("Success", func(t *testing.T) {
		mockClient := new(MockFirebaseMessenger)
		pushService := services.NewPushNotification(mockClient)

		// --- FIX: Use mock.Anything for the context ---
		mockClient.On("SendEach", mock.Anything, expectedMessages).Return(nil, nil).Once()
		// --- End Fix ---

		err := pushService.SendNotification(testNotification)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	// --- Test Case: Failure ---
	t.Run("Failure", func(t *testing.T) {
		mockClient := new(MockFirebaseMessenger)
		pushService := services.NewPushNotification(mockClient)

		expectedError := errors.New("FCM service unavailable")
		// --- FIX: Use mock.Anything for the context ---
		mockClient.On("SendEach", mock.Anything, expectedMessages).Return(nil, expectedError).Once()
		// --- End Fix ---

		err := pushService.SendNotification(testNotification)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockClient.AssertExpectations(t)
	})

	// --- Test Case: No Tokens (remains the same) ---
	t.Run("NoTokens", func(t *testing.T) {
		mockClient := new(MockFirebaseMessenger)
		pushService := services.NewPushNotification(mockClient)

		noTokenNotification := domain.NotificationData{
			Address: []string{},
			Title:   "Push Title",
			Message: "Push message body.",
		}

		err := pushService.SendNotification(noTokenNotification)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}
