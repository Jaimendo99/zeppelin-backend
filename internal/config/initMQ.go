package config

import "github.com/rabbitmq/amqp091-go"

var (
	MQConn          *amqp091.Connection
	ConsumerChannel *amqp091.Channel
	ProducerChannel *amqp091.Channel
)

func InitMQ(connectionString string) error {
	conn, err := amqp091.Dial(connectionString)
	if err != nil {
		return err
	}
	MQConn = conn

	ConsumerChannel, err = conn.Channel()
	if err != nil {
		return err
	}

	ProducerChannel, err = conn.Channel()
	if err != nil {
		return err
	}

	return nil
}
