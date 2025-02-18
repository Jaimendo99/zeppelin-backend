package services

import (
	"log"
	"zeppelin/internal/domain"
)

type NotificationPrinter struct {
}

func (p NotificationPrinter) SendNotification(notification domain.NotificationQueue) error {
	log.Print("Notification sent", notification)
	return nil
}
