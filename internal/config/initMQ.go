package config

import (
	// "fmt" // Only if needed elsewhere
	"github.com/rabbitmq/amqp091-go"
)

type AmqpConnection interface {
	Channel() (*amqp091.Channel, error)
	Close() error
}

var (
	MQConn          AmqpConnection
	ConsumerChannel *amqp091.Channel
	ProducerChannel *amqp091.Channel
)

var amqpDial func(url string) (AmqpConnection, error)

func init() {
	amqpDial = func(url string) (AmqpConnection, error) {
		conn, err := amqp091.Dial(url)
		if err != nil {
			var nilConn AmqpConnection = nil
			return nilConn, err
		}
		return conn, nil
	}
}

var AmqpDial = &amqpDial

func InitMQ(connectionString string) error {
	conn, err := amqpDial(connectionString)
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
		// Consider cleanup if needed
		return err
	}

	return nil
}
func ResetMQState() {
	MQConn = nil // Now resets an interface variable
	ConsumerChannel = nil
	ProducerChannel = nil
	// Reset dialer (same as before)
	amqpDial = func(url string) (AmqpConnection, error) {
		conn, err := amqp091.Dial(url)
		if err != nil {
			var nilConn AmqpConnection = nil
			return nilConn, err
		}
		return conn, nil
	}
}
