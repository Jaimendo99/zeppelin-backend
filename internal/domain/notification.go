package domain

type NotificationQueue struct {
	NotificacionId string `json:"notification_id"`
	Message        string `json:"message"`
	Receiver       string `json:"receiver"`
}

type NotificationService interface {
	SendNotification(notification NotificationQueue) error
}

type NotificationRepo interface {
	SendToQueue(notification NotificationQueue, queueName string) error
	ConsumeFromQueue(queueName string) error
}
