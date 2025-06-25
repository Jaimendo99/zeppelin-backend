package domain

import (
	"context"
	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
	"net/smtp"
	"testing"
)

type NotificationQueue struct {
	NotificationId string   `json:"notification_id"`
	Title          string   `json:"title" validate:"required"`
	Message        string   `json:"message" validate:"required"`
	ReceiverId     []string `json:"receiver_id" validate:"required,gt=0"` // Example: required, must have > 0 elements
}

type NotificationData struct {
	NotificationId string   `json:"notification_id"`
	Title          string   `json:"title"`
	Message        string   `json:"message"`
	Address        []string `json:"address"`
}

type NotificationService interface {
	SendNotification(notification NotificationData) error
}

type NotificationRepo interface {
	SendToQueue(notification NotificationQueue, queueName string) error
	ConsumeFromQueue(queueName string) error
}

// FirebaseMessenger defines the interface for sending FCM messages
type FirebaseMessenger interface {
	SendEach(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error)
	// Add other methods from messaging.Client if your service uses them
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(notification NotificationData) error {
	args := m.Called(notification)
	return args.Error(0)
}

func NewMockQueue(t *testing.T) *MockQueue {
	m := new(MockQueue)
	m.Mock.Test(t) // Associate with the test for better error reporting
	return m
}

func NewMockNotificationService(t *testing.T) *MockNotificationService {
	m := new(MockNotificationService)
	m.Mock.Test(t) // Associate with the test
	return m
}

type SmtpConfig struct {
	Host     string
	Port     string
	Username string
	Auth     smtp.Auth
}
