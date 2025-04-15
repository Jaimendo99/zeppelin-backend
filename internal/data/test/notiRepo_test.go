package test_test

import (
	"encoding/json"
	"errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
)

func TestNotificationRepo_SendToQueue(t *testing.T) {
	queueName := "test-notifications"
	notification := domain.NotificationQueue{
		NotificationId: "nid-123",
		Title:          "Test Title",
		Message:        "Test Message",
		ReceiverId:     []string{"1", "2"},
	}

	t.Run("Success", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		// No notification services needed for this method
		repo := data.NewNotificationRepo(nil, mockQueue, nil) // Pass nil for DB and services

		mockQueue.On("SendToQueue", notification, queueName).Return(nil).Once()

		err := repo.SendToQueue(notification, queueName)

		assert.NoError(t, err)
		mockQueue.AssertExpectations(t)
	})

	t.Run("Queue Error", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		repo := data.NewNotificationRepo(nil, mockQueue, nil)

		expectedErr := errors.New("failed to send to queue")
		mockQueue.On("SendToQueue", notification, queueName).Return(expectedErr).Once()

		err := repo.SendToQueue(notification, queueName)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockQueue.AssertExpectations(t)
	})
}

func TestNotificationRepo_ConsumeFromQueue(t *testing.T) {
	queueName := "consume-queue"
	notificationInput := domain.NotificationQueue{
		NotificationId: "consume-nid-456",
		Title:          "Consumed Title",
		Message:        "Consumed Message",
		ReceiverId:     []string{"1", "99"}, // Use IDs matching GetReceiverAddress logic
	}
	notificationBody, _ := json.Marshal(notificationInput)

	// Expected data based on hardcoded GetReceiverAddress
	expectedEmailData := domain.NotificationData{
		NotificationId: notificationInput.NotificationId,
		Title:          notificationInput.Title,
		Message:        notificationInput.Message,
		Address:        []string{"jaimendo26@gmail.com", "jaimendo99@gmail.com"},
	}
	expectedFCMData := domain.NotificationData{
		NotificationId: notificationInput.NotificationId,
		Title:          notificationInput.Title,
		Message:        notificationInput.Message,
		Address:        []string{"fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU", "fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU"},
	}

	t.Run("Success", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		mockService1 := domain.NewMockNotificationService(t) // e.g., Email service
		mockService2 := domain.NewMockNotificationService(t) // e.g., FCM service
		services := []domain.NotificationService{mockService1, mockService2}
		repo := data.NewNotificationRepo(nil, mockQueue, services) // Pass nil for DB

		// Channel to simulate message delivery
		msgChan := make(chan amqp091.Delivery, 1) // Buffered to avoid blocking sender

		// Mock ConsumeFromQueue to return the channel
		mockQueue.On("ConsumeFromQueue", queueName).Return((<-chan amqp091.Delivery)(msgChan), nil).Once()

		// Mock the notification services
		mockService1.On("SendNotification", expectedEmailData).Return(nil).Once()
		mockService2.On("SendNotification", expectedFCMData).Return(nil).Once()

		// Goroutine to simulate message arrival and channel closing
		go func() {
			msgChan <- amqp091.Delivery{Body: notificationBody}
			// Close the channel to signal the end of messages for the range loop
			close(msgChan)
		}()

		// Call the method under test - this will block until the channel is closed
		err := repo.ConsumeFromQueue(queueName)

		assert.NoError(t, err)
		// Assert all mocks were called as expected
		mockQueue.AssertExpectations(t)
		mockService1.AssertExpectations(t)
		mockService2.AssertExpectations(t)
	})

	t.Run("Consume Error", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		repo := data.NewNotificationRepo(nil, mockQueue, nil) // Services not needed

		expectedErr := errors.New("failed to consume")
		// Mock ConsumeFromQueue to return an error
		mockQueue.On("ConsumeFromQueue", queueName).Return(nil, expectedErr).Once()

		err := repo.ConsumeFromQueue(queueName)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockQueue.AssertExpectations(t)
	})

	t.Run("Unmarshal Error", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		repo := data.NewNotificationRepo(nil, mockQueue, nil) // Services not needed

		msgChan := make(chan amqp091.Delivery, 1)
		mockQueue.On("ConsumeFromQueue", queueName).Return((<-chan amqp091.Delivery)(msgChan), nil).Once()

		go func() {
			msgChan <- amqp091.Delivery{Body: []byte("this is not json")}
			close(msgChan)
		}()

		err := repo.ConsumeFromQueue(queueName)

		assert.Error(t, err)
		// Check if the error is a JSON unmarshal error (more specific check)
		_, isJsonError := err.(*json.SyntaxError)
		assert.True(t, isJsonError, "Expected a JSON syntax error")
		mockQueue.AssertExpectations(t)
	})

	t.Run("Send Notification Error - Service 1", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		mockService1 := domain.NewMockNotificationService(t)
		mockService2 := domain.NewMockNotificationService(t) // Still need the mock even if not expected to be called
		services := []domain.NotificationService{mockService1, mockService2}
		repo := data.NewNotificationRepo(nil, mockQueue, services)

		msgChan := make(chan amqp091.Delivery, 1)
		mockQueue.On("ConsumeFromQueue", queueName).Return((<-chan amqp091.Delivery)(msgChan), nil).Once()

		expectedErr := errors.New("failed sending via service 1")
		// Mock Service 1 to return an error
		mockService1.On("SendNotification", expectedEmailData).Return(expectedErr).Once()
		// Service 2 should NOT be called if Service 1 errors out
		// mockService2.On("SendNotification", mock.Anything).Return(nil) // No expectation for service 2

		go func() {
			msgChan <- amqp091.Delivery{Body: notificationBody}
			close(msgChan)
		}()

		err := repo.ConsumeFromQueue(queueName)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockQueue.AssertExpectations(t)
		mockService1.AssertExpectations(t)
		mockService2.AssertNotCalled(t, "SendNotification", mock.Anything) // Verify service 2 wasn't called
	})

	t.Run("Send Notification Error - Service 2", func(t *testing.T) {
		mockQueue := domain.NewMockQueue(t)
		mockService1 := domain.NewMockNotificationService(t)
		mockService2 := domain.NewMockNotificationService(t)
		services := []domain.NotificationService{mockService1, mockService2}
		repo := data.NewNotificationRepo(nil, mockQueue, services)

		msgChan := make(chan amqp091.Delivery, 1)
		mockQueue.On("ConsumeFromQueue", queueName).Return((<-chan amqp091.Delivery)(msgChan), nil).Once()

		expectedErr := errors.New("failed sending via service 2")
		// Mock Service 1 to succeed
		mockService1.On("SendNotification", expectedEmailData).Return(nil).Once()
		// Mock Service 2 to return an error
		mockService2.On("SendNotification", expectedFCMData).Return(expectedErr).Once()

		go func() {
			msgChan <- amqp091.Delivery{Body: notificationBody}
			close(msgChan)
		}()

		err := repo.ConsumeFromQueue(queueName)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockQueue.AssertExpectations(t)
		mockService1.AssertExpectations(t)
		mockService2.AssertExpectations(t)
	})

}

func TestNotificationRepo_GetReceiverAddress(t *testing.T) {
	// Instantiate repo - dependencies don't matter for this specific test yet
	repo := data.NewNotificationRepo(nil, nil, nil)

	// TODO: Update this test when GetReceiverAddress uses the database.
	// This will require setting up the gormDb and sqlmock, expecting a SELECT query,
	// and returning mock rows corresponding to the user ID.

	t.Run("ID is 1", func(t *testing.T) {
		id := "1"
		expected := data.ReceiverAddr{
			UserId:   1,
			FCMToken: "fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU",
			Email:    "jaimendo26@gmail.com",
		}
		actual := repo.GetReceiverAddress(id)
		assert.Equal(t, expected, actual)
	})

	t.Run("ID is not 1", func(t *testing.T) {
		id := "any-other-id"
		expected := data.ReceiverAddr{
			UserId:   2,
			FCMToken: "fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU",
			Email:    "jaimendo99@gmail.com",
		}
		actual := repo.GetReceiverAddress(id)
		assert.Equal(t, expected, actual)
	})
}
