package controller

import (
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type UserFcmTokenController struct {
	Repo domain.UserFcmTokenRepo
}

func (c *UserFcmTokenController) CreateUserFcmToken() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		var input domain.UserFcmTokenInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		token := domain.UserFcmTokenDb{
			UserID:        userID,
			FirebaseToken: input.FirebaseToken,
			DeviceType:    input.DeviceType,
			DeviceInfo:    input.DeviceInfo,
		}

		err := c.Repo.CreateUserFcmToken(token)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Token registrado con éxito"})
	}
}

func (c *UserFcmTokenController) GetUserFcmTokens() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		tokens, err := c.Repo.GetUserFcmTokensByUserID(userID)
		return ReturnReadResponse(e, err, tokens)
	}
}

func (c *UserFcmTokenController) DeleteUserFcmToken() echo.HandlerFunc {
	return func(e echo.Context) error {
		var req domain.UserFcmTokenDeleteInput
		if err := ValidateAndBind(e, &req); err != nil {
			return err
		}

		err := c.Repo.DeleteUserFcmTokenByToken(req.FirebaseToken)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Token eliminado con éxito"})
	}
}

func (c *UserFcmTokenController) UpdateDeviceInfo() echo.HandlerFunc {
	return func(e echo.Context) error {
		var req domain.UserFcmTokenUpdateDeviceInput
		if err := ValidateAndBind(e, &req); err != nil {
			return err
		}

		err := c.Repo.UpdateDeviceInfo(req.FirebaseToken, req.DeviceInfo)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Información del dispositivo actualizada con éxito"})
	}
}

func (c *UserFcmTokenController) UpdateWebToken() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		type Request struct {
			FirebaseToken string `json:"firebase_token" validate:"required"`
		}
		var req Request
		if err := ValidateAndBind(e, &req); err != nil {
			return err
		}

		err := c.Repo.UpdateFirebaseToken(userID, "WEB", req.FirebaseToken)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Firebase token WEB actualizado con éxito"})
	}
}

func (c *UserFcmTokenController) UpdateMobileToken() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		type Request struct {
			FirebaseToken string `json:"firebase_token" validate:"required"`
		}
		var req Request
		if err := ValidateAndBind(e, &req); err != nil {
			return err
		}

		err := c.Repo.UpdateFirebaseToken(userID, "MOBILE", req.FirebaseToken)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Firebase token MOBILE actualizado con éxito"})
	}
}
