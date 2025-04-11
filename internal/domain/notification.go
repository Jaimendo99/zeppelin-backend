package domain

import "net/smtp"

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

type SmtpConfig struct {
	Host     string
	Port     string
	Username string
	Auth     smtp.Auth
}
