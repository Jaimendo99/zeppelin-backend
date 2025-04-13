package domain

import (
	"context"
	"firebase.google.com/go/v4/messaging"
	"net/smtp"
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

type SmtpConfig struct {
	Host     string
	Port     string
	Username string
	Auth     smtp.Auth
}
