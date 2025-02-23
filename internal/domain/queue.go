package domain

import "github.com/rabbitmq/amqp091-go"

type Queue interface {
	SendToQueue(notification NotificationQueue, queueName string) error
	ConsumeFromQueue(queueName string) (<-chan amqp091.Delivery, error)
}
