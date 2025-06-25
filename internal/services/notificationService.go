package services

import (
	"context"
	"log"
	"net/smtp"
	"strings"
	"zeppelin/internal/domain"

	"firebase.google.com/go/v4/messaging"
)

var SmtpSendMail = smtp.SendMail //nolint:gochecknoglobals

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
		"Subject: Notification\r\n" + // Consider making Subject dynamic if needed
		"\r\n" +
		notification.Message + "\r\n")

	addr := e.smtpConfig.Host + ":" + e.smtpConfig.Port
	err := SmtpSendMail(
		addr,
		e.smtpConfig.Auth,
		e.smtpConfig.Username,
		notification.Address,
		msg,
	)

	if err != nil {
		log.Printf("Error sending email notification: %v", err) // Log the error
		return err
	}
	log.Println("Email Notification sent")

	return nil
}

type PushNotification struct {
	client domain.FirebaseMessenger
}

func NewPushNotification(client domain.FirebaseMessenger) *PushNotification {
	return &PushNotification{client: client}
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

	// Handle empty messages case to avoid unnecessary API call
	if len(messages) == 0 {
		log.Println("No push notification tokens provided, skipping send.")
		return nil // Or return an error if appropriate
	}

	// Call the method via the interface
	_, err := p.client.SendEach(context.Background(), messages)
	if err != nil {
		log.Printf("Error sending push notification: %v", err) // Log the error
		return err
	}
	log.Println("Push Notification sent")
	return nil
}
