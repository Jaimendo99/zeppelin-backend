package data

import (
	"encoding/json"
	"log"
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type NotificationRepo struct {
	database            *gorm.DB
	queue               domain.Queue
	notificationService []domain.NotificationService
}

func NewNotificationRepo(db *gorm.DB, q domain.Queue, s []domain.NotificationService) *NotificationRepo {
	return &NotificationRepo{database: db, queue: q, notificationService: s}
}

func (n *NotificationRepo) SendToQueue(notification domain.NotificationQueue, queueName string) error {
	return n.queue.SendToQueue(notification, queueName)

}
func (n *NotificationRepo) ConsumeFromQueue(queueName string) error {
	log.Println("Consuming from queue")
	msgs, err := n.queue.ConsumeFromQueue(queueName)
	if err != nil {
		return err
	}
	for d := range msgs {
		log.Println("Message received")
		notification := domain.NotificationQueue{}
		err := json.Unmarshal(d.Body, &notification)
		log.Println(notification)
		if err != nil {
			log.Println("Error unmarshalling")
			return err
		}

		emailAddresses := []string{}
		fcmTokens := []string{}
		for _, id := range notification.ReceiverId {
			log.Println("Getting receiver address: " + id)
			receiver := n.GetReceiverAddress(id)
			emailAddresses = append(emailAddresses, receiver.Email)
			fcmTokens = append(fcmTokens, receiver.FCMToken)
		}

		notificationDataEmail := domain.NotificationData{
			NotificationId: notification.NotificationId,
			Title:          notification.Title,
			Message:        notification.Message,
			Address:        emailAddresses,
		}
		notificationDataFCM := domain.NotificationData{
			NotificationId: notification.NotificationId,
			Title:          notification.Title,
			Message:        notification.Message,
			Address:        fcmTokens,
		}

		datas := []domain.NotificationData{notificationDataEmail, notificationDataFCM}
		log.Println("Sending notification")
		for i, s := range n.notificationService {
			err := s.SendNotification(datas[i])
			if err != nil {
				log.Println("Error sending notification")
				return err
			}
		}
	}
	return nil
}

func (n NotificationRepo) GetReceiverAddress(id string) ReceiverAddr {
	// TODO: Implement logic to get receiver address from database
	if id == "1" {
		return ReceiverAddr{
			UserId:   1,
			FCMToken: "fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU",
			Email:    "jaimendo26@gmail.com",
		}
	} else {
		return ReceiverAddr{
			UserId:   2,
			FCMToken: "fycHDn1xTNuSMrZd3kCmWx:APA91bGzHCiPqj-GZZXu_JWiMc8kjROf1jIxbc1kJP_YM6rYnZOucVcMFcOe23wftKBJMPceRk_kwZbNz4Vrp7jV_OtDh9vHk8TmRBmFLMw20Hl5RMlCbCU",
			Email:    "jaimendo99@gmail.com",
		}
	}
}

type ReceiverAddr struct {
	UserId   int    `json:"user_id"`
	FCMToken string `json:"fcm_token"`
	Email    string `json:"email"`
}
