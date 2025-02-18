package db

import (
	"encoding/json"
	"zeppelin/internal/domain"

	"github.com/rabbitmq/amqp091-go"
)

type NotificationRabbitMq struct {
	channel             *amqp091.Channel
	notificationService domain.NotificationService
}

func NewNotificationMq(ch *amqp091.Channel, service domain.NotificationService) *NotificationRabbitMq {
	return &NotificationRabbitMq{channel: ch, notificationService: service}
}

func (n *NotificationRabbitMq) SendToQueue(notification domain.NotificationQueue, queueName string) error {
	q, err := n.channel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return err
	}
	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return n.channel.Publish("", q.Name, false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
func (n *NotificationRabbitMq) ConsumeFromQueue(queueName string) error {
	// Declare the queue to ensure it exists.
	q, err := n.channel.QueueDeclare(
		queueName, // queue name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := n.channel.Consume(
		q.Name, // use the declared queue name
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	// Process messages
	for msg := range msgs {
		var notification domain.NotificationQueue
		if err := json.Unmarshal(msg.Body, &notification); err != nil {
			return err
		}
		if err := n.notificationService.SendNotification(notification); err != nil {
			return err
		}
	}
	return nil
}
