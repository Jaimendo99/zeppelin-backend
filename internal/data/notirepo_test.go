package data_test

import (
	"encoding/json"
	"errors"
	"testing"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

type FakeAMQPChannel struct {
	QueueDeclareFunc func(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
	PublishFunc      func(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	ConsumeFunc      func(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
}

func (f *FakeAMQPChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	return f.QueueDeclareFunc(name, durable, autoDelete, exclusive, noWait, args)
}
func (f *FakeAMQPChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return f.PublishFunc(exchange, key, mandatory, immediate, msg)
}
func (f *FakeAMQPChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	return f.ConsumeFunc(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func TestSendToQueue_Success(t *testing.T) {
	fakeChannel := &FakeAMQPChannel{}

	// Fake QueueDeclare: return a queue with the same name.
	fakeChannel.QueueDeclareFunc = func(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
		return amqp091.Queue{Name: name}, nil
	}
	var publishedMsg amqp091.Publishing
	fakeChannel.PublishFunc = func(exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
		publishedMsg = msg
		return nil
	}

	rabbit := data.NewRabbitMQImpl(fakeChannel)

	notification := domain.NotificationQueue{
		NotificationId: "id1",
		Title:          "Title",
		Message:        "Message",
		ReceiversId:    []string{"receiver1"},
	}
	err := rabbit.SendToQueue(notification, "testQueue")
	assert.NoError(t, err)

	// Check that the published message contains the proper JSON
	var parsedNotification domain.NotificationQueue
	err = json.Unmarshal(publishedMsg.Body, &parsedNotification)
	assert.NoError(t, err)
	assert.Equal(t, notification, parsedNotification)
}

func TestConsumeFromQueue_Success(t *testing.T) {
	fakeChannel := &FakeAMQPChannel{}

	fakeChannel.QueueDeclareFunc = func(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
		return amqp091.Queue{Name: name}, nil
	}
	fakeDeliveries := make(chan amqp091.Delivery, 1)
	fakeDelivery := amqp091.Delivery{Body: []byte("test message")}
	fakeDeliveries <- fakeDelivery
	close(fakeDeliveries)

	fakeChannel.ConsumeFunc = func(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
		return fakeDeliveries, nil
	}

	rabbit := data.NewRabbitMQImpl(fakeChannel)

	deliveries, err := rabbit.ConsumeFromQueue("testQueue")
	assert.NoError(t, err)

	d, ok := <-deliveries
	assert.True(t, ok)
	assert.Equal(t, fakeDelivery.Body, d.Body)
}

type FakeQueue struct {
	SendToQueueFunc      func(notification domain.NotificationQueue, queueName string) error
	ConsumeFromQueueFunc func(queueName string) (<-chan amqp091.Delivery, error)
}

func (f *FakeQueue) SendToQueue(notification domain.NotificationQueue, queueName string) error {
	if f.SendToQueueFunc != nil {
		return f.SendToQueueFunc(notification, queueName)
	}
	return nil
}

func (f *FakeQueue) ConsumeFromQueue(queueName string) (<-chan amqp091.Delivery, error) {
	if f.ConsumeFromQueueFunc != nil {
		return f.ConsumeFromQueueFunc(queueName)
	}
	return nil, errors.New("ConsumeFromQueueFunc not defined")
}

type FakeNotificationService struct {
	ReceivedData domain.NotificationData
	Err          error
}

func (f *FakeNotificationService) SendNotification(data domain.NotificationData) error {
	f.ReceivedData = data
	return f.Err
}

func TestNotificationRepo_SendToQueue(t *testing.T) {
	fakeQueue := &FakeQueue{
		SendToQueueFunc: func(notification domain.NotificationQueue, queueName string) error {
			assert.Equal(t, "testQueue", queueName)
			return nil
		},
	}

	repo := data.NewNotificationRepo(nil, fakeQueue, nil)

	notification := domain.NotificationQueue{
		NotificationId: "notif1",
		Title:          "Test Title",
		Message:        "Test Message",
		ReceiversId:    []string{"receiver1"},
	}
	err := repo.SendToQueue(notification, "testQueue")
	assert.NoError(t, err)
}

func TestNotificationRepo_ConsumeFromQueue_Success(t *testing.T) {
	fakeDeliveries := make(chan amqp091.Delivery, 1)
	notification := domain.NotificationQueue{
		NotificationId: "notif1",
		Title:          "Test Title",
		Message:        "Test Message",
		ReceiversId:    []string{"receiver1", "receiver2"},
	}
	body, err := json.Marshal(notification)
	assert.NoError(t, err)

	fakeDeliveries <- amqp091.Delivery{Body: body}
	close(fakeDeliveries)

	fakeQueue := &FakeQueue{
		ConsumeFromQueueFunc: func(queueName string) (<-chan amqp091.Delivery, error) {
			assert.Equal(t, "testQueue", queueName)
			return fakeDeliveries, nil
		},
	}

	fakeNotifService1 := &FakeNotificationService{}
	fakeNotifService2 := &FakeNotificationService{}

	repo := data.NewNotificationRepo(nil, fakeQueue, []domain.NotificationService{fakeNotifService1, fakeNotifService2})

	err = repo.ConsumeFromQueue("testQueue")
	assert.NoError(t, err)

	expectedEmailData := domain.NotificationData{
		NotificationId: "notif1",
		Title:          "Test Title",
		Message:        "Test Message",
		Address:        []string{"jaimendo99@gmail.com", "jaimendo99@gmail.com"},
	}
	expectedFCMData := domain.NotificationData{
		NotificationId: "notif1",
		Title:          "Test Title",
		Message:        "Test Message",
		Address:        []string{"fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU", "fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU"},
	}

	assert.Equal(t, expectedEmailData, fakeNotifService1.ReceivedData)
	assert.Equal(t, expectedFCMData, fakeNotifService2.ReceivedData)
}
