package domain

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type Queue interface {
	SendToQueue(notification NotificationQueue, queueName string) error
	ConsumeFromQueue(queueName string) (<-chan amqp091.Delivery, error)
}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) SendToQueue(notification NotificationQueue, queueName string) error {
	args := m.Called(notification, queueName)
	return args.Error(0)
}

func (m *MockQueue) ConsumeFromQueue(queueName string) (<-chan amqp091.Delivery, error) {
	args := m.Called(queueName)
	// Need to handle potential nil channel return
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan amqp091.Delivery), args.Error(1)
}
