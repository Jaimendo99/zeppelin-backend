package data

import (
	"encoding/json"
	"zeppelin/internal/domain"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQImpl struct {
	ch *amqp091.Channel
}

func NewRabbitMQImpl(ch *amqp091.Channel) *RabbitMQImpl {
	return &RabbitMQImpl{ch: ch}
}

func (r *RabbitMQImpl) SendToQueue(notification domain.NotificationQueue, queueName string) error {
	q, err := r.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return err
	}

	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return r.ch.Publish("", q.Name, false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQImpl) ConsumeFromQueue(queueName string) (<-chan amqp091.Delivery, error) {
	q, err := r.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := r.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
