package services

import (
	"context"
	"errors"
	"net/smtp"
	"testing"
	"zeppelin/internal/domain"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/assert"
)

func TestRepresetativeInputToDb_FullFields(t *testing.T) {
	input := domain.RepresentativeInput{
		Name:        "Mateo",
		Lastname:    "Mejia",
		Email:       "jaimendo26@gmail.com",
		PhoneNumber: "+129129122",
	}

	dbModel := RepresetativeInputToDb(&input)

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

	dbModel := RepresetativeInputToDb(&input)

	assert.Equal(t, input.Name, dbModel.Name, "Name should match")
	assert.Equal(t, input.Lastname, dbModel.Lastname, "Lastname should match")

	assert.False(t, dbModel.Email.Valid, "Email.Valid should be false for empty input")
	assert.Equal(t, "", dbModel.Email.String, "Email should be empty")

	assert.False(t, dbModel.PhoneNumber.Valid, "PhoneNumber.Valid should be false for empty input")
	assert.Equal(t, "", dbModel.PhoneNumber.String, "PhoneNumber should be empty")
}

func TestParamToId_Valid(t *testing.T) {
	id, err := ParamToId("123")
	assert.NoError(t, err, "Expected no error for valid numeric string")
	assert.Equal(t, 123, id, "Expected id to be 123")
}

func TestParamToId_Invalid(t *testing.T) {
	id, err := ParamToId("abc")
	assert.Error(t, err, "Expected error for non-numeric string")
	assert.Equal(t, -1, id, "Expected id to be -1 on error")
}

// TestEmailNotification_Success verifies that the email is "sent" correctly.
func TestEmailNotification_Success(t *testing.T) {
	// Override the smtpSendMail function.
	originalSendMail := smtpSendMail
	defer func() { smtpSendMail = originalSendMail }()

	var called bool
	var receivedAddr, receivedUsername string
	var receivedTo []string

	smtpSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		called = true
		receivedAddr = addr
		receivedUsername = from
		receivedTo = to
		return nil
	}

	// Create a dummy SMTP config.
	smtpConfig := domain.SmtpConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Auth:     nil, // For testing, nil is fine.
		Username: "user@example.com",
	}

	emailNotif := NewEmailNotification(smtpConfig)
	notification := domain.NotificationData{
		NotificationId: "notif1",
		Title:          "Email Title",
		Message:        "Email message",
		Address:        []string{"recipient@example.com"},
	}

	err := emailNotif.SendNotification(notification)
	assert.NoError(t, err)
	assert.True(t, called, "smtpSendMail should be called")
	// Verify that the address and from fields are correctly composed.
	assert.Equal(t, "smtp.example.com:587", receivedAddr)
	assert.Equal(t, "user@example.com", receivedUsername)
	assert.Equal(t, []string{"recipient@example.com"}, receivedTo)
	// Optionally, check the message content.
}

func TestEmailNotification_Error(t *testing.T) {
	originalSendMail := smtpSendMail
	defer func() { smtpSendMail = originalSendMail }()

	smtpSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("failed to send")
	}

	smtpConfig := domain.SmtpConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Auth:     nil,
		Username: "user@example.com",
	}
	emailNotif := NewEmailNotification(smtpConfig)
	notification := domain.NotificationData{
		NotificationId: "notif1",
		Title:          "Email Title",
		Message:        "Email message",
		Address:        []string{"recipient@example.com"},
	}
	err := emailNotif.SendNotification(notification)
	assert.Error(t, err)
}

type FakeMessagingClient struct {
	SendEachFunc func(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error)
}

func (f *FakeMessagingClient) SendEach(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error) {
	if f.SendEachFunc != nil {
		return f.SendEachFunc(ctx, messages)
	}
	return &messaging.BatchResponse{}, nil
}

func TestPushNotification_Success(t *testing.T) {
	fakeClient := &FakeMessagingClient{
		SendEachFunc: func(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error) {
			assert.Len(t, messages, 2)
			return &messaging.BatchResponse{SuccessCount: len(messages)}, nil
		},
	}
	pushNotif := NewPushNotification(fakeClient)
	notification := domain.NotificationData{
		NotificationId: "notif1",
		Title:          "Push Title",
		Message:        "Push message",
		Address:        []string{"token1", "token2"},
	}

	err := pushNotif.SendNotification(notification)
	assert.NoError(t, err)
}

func TestPushNotification_Error(t *testing.T) {
	fakeClient := &FakeMessagingClient{
		SendEachFunc: func(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error) {
			return nil, errors.New("failed to send push")
		},
	}
	pushNotif := NewPushNotification(fakeClient)
	notification := domain.NotificationData{
		NotificationId: "notif1",
		Title:          "Push Title",
		Message:        "Push message",
		Address:        []string{"token1"},
	}
	err := pushNotif.SendNotification(notification)
	assert.Error(t, err)
}
