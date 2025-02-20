package services

import (
	"context"
	"log"
	"net/smtp"
	"strings"
	"zeppelin/internal/domain"

	"firebase.google.com/go/v4/messaging"
)

type EmailNotification struct {
	smtpConfig domain.SmtpConfig
}

func NewEmailNotification(s domain.SmtpConfig) *EmailNotification {
	return &EmailNotification{smtpConfig: s}
}

func (e *EmailNotification) SendNotification(notification domain.NotificationData) error {
	log.Println("Sending email notification")
	receivers := strings.Join(notification.Address, ",")
	msg := []byte("To: " + receivers + "\r\n" +
		"Subject: Notification\r\n" +
		"\r\n" +
		notification.Message + "\r\n")

	err := smtp.SendMail(
		e.smtpConfig.Host+":"+e.smtpConfig.Port,
		e.smtpConfig.Auth,
		e.smtpConfig.Username,
		notification.Address,
		msg,
	)

	if err != nil {
		log.Println("Error sending email notification")
		return err
	}
	log.Println("Email Notification sent")

	return nil
}

type PushNotification struct {
	client *messaging.Client
}

func NewPushNotification(client messaging.Client) *PushNotification {
	return &PushNotification{client: &client}
}

func (p *PushNotification) SendNotification(notification domain.NotificationData) error {
	log.Println("Sending push notification")
	var messages []*messaging.Message
	for _, token := range notification.Address {
		message := &messaging.Message{
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Title: notification.Title,
					Body:  notification.Message,
				},
			},
			Token: token,
		}
		messages = append(messages, message)
	}

	_, err := p.client.SendEach(context.Background(), messages)
	if err != nil {
		log.Println("Error sending push notification")
		return err
	}
	log.Println("Push Notification sent")
	return nil
}
