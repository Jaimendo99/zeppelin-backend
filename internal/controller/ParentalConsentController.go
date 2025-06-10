package controller

import (
	"net/http"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type ParentalConsentController struct {
	Repo domain.ParentalConsentRepo
}

// GetConsentByToken obtiene el consentimiento parental por token
func (c *ParentalConsentController) GetConsentByToken() echo.HandlerFunc {
	return func(e echo.Context) error {
		token := e.QueryParam("token")
		if token == "" {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusBadRequest, "token requerido"), nil)
		}

		consent, err := c.Repo.GetConsentByToken(token)
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusNotFound, "consentimiento no encontrado"), nil)
		}

		return ReturnReadResponse(e, nil, consent)
	}
}

// UpdateConsentStatus actualiza el estado del consentimiento parental
func (c *ParentalConsentController) UpdateConsentStatus() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input struct {
			Token  string `json:"token" validate:"required"`
			Status string `json:"status" validate:"required,oneof=ACCEPTED REJECTED"`
		}

		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		// Obtener información del request
		ip := e.RealIP()
		userAgent := e.Request().Header.Get("User-Agent")

		err := c.Repo.UpdateConsentStatus(input.Token, input.Status, ip, userAgent)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al actualizar el consentimiento"), nil)
		}

		var message string
		switch input.Status {
		case "ACCEPTED":
			message = "Consentimiento aprobado exitosamente"
		case "REJECTED":
			message = "Consentimiento rechazado"
		default:
			message = "Estado del consentimiento actualizado"
		}

		return ReturnWriteResponse(e, nil, map[string]string{
			"message": message,
			"status":  input.Status,
		})
	}
}

// GetConsentByUserID obtiene el consentimiento parental del usuario autenticado
func (c *ParentalConsentController) GetConsentByUserID() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		consent, err := c.Repo.GetConsentByUserID(userID)
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusNotFound, "consentimiento no encontrado para este usuario"), nil)
		}

		return ReturnReadResponse(e, nil, consent)
	}
}
