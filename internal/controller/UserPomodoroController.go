package controller

import (
	"github.com/labstack/echo/v4"
	"zeppelin/internal/domain"
)

type PomodoroController struct {
	Repo domain.UserPomodoroRepo
}

func (c *PomodoroController) GetPomodoroByUserID() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)
		data, err := c.Repo.GetByUserID(userID)
		return ReturnReadResponse(e, err, data)
	}
}

func (c *PomodoroController) UpdatePomodoroByUserID() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)
		var input domain.UpdatePomodoroInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}
		err := c.Repo.UpdateByUserID(userID, input)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Configuración actualizada"})
	}
}
