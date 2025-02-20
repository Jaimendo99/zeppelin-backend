package services

import (
	"log"
	"net/smtp"
	"strings"
	"zeppelin/internal/domain"
)

type EmailNotification struct {
	smtpConfig domain.SmtpConfig
}

func NewEmailNotification(s domain.SmtpConfig) *EmailNotification {
	return &EmailNotification{smtpConfig: s}
}

func (e *EmailNotification) SendNotification(notification domain.NotificationQueue) error {
	log.Println("Sending email notification")
	receivers := strings.Join(notification.Receiver, ",")
	msg := []byte("To: " + receivers + "\r\n" +
		"Subject: Notification\r\n" +
		"\r\n" +
		notification.Message + "\r\n")

	err := smtp.SendMail(
		e.smtpConfig.Host+":"+e.smtpConfig.Port,
		e.smtpConfig.Auth,
		e.smtpConfig.Username,
		notification.Receiver, msg)

	if err != nil {
		log.Println("Error sending email notification")
		return err
	}
	log.Println("Notification sent")

	return nil
}
