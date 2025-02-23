package domain

import (
	"context"
	"net/smtp"

	"firebase.google.com/go/v4/messaging"
	"github.com/rabbitmq/amqp091-go"
)

type NotificationQueue struct {
	NotificationId string   `json:"notification_id" validate:"required"`
	Title          string   `json:"title" validate:"required"`
	Message        string   `json:"message" validate:"required"`
	ReceiversId    []string `json:"receiver_id" validate:"required,min=1,dive,required"`
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

type AMQPChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
}

type MessagingClient interface {
	SendEach(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error)
}
