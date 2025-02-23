package controller

import (
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type NotificationController struct {
	Repo domain.NotificationRepo
}

func NewNotificationController(repo domain.NotificationRepo) *NotificationController {
	return &NotificationController{Repo: repo}
}

func (c *NotificationController) SendNotification() func(e echo.Context) error {
	return func(e echo.Context) error {
		notification := domain.NotificationQueue{}
		if err := ValidateAndBind(e, &notification); err != nil {
			return err
		}
		err := c.Repo.SendToQueue(notification, "notification")
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Notification sent"})
	}
}
