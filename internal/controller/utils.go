package controller

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func ReturnReadResponse(e echo.Context, err error, body any) error {
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			return echo.NewHTTPError(http.StatusBadRequest, struct {
				Message string `json:"message"`
			}{Message: "Invalid request"})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, struct {
				Message string `json:"message"`
			}{Message: "Record not found"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, struct {
			Message string `json:"message"`
		}{Message: "Internal server error"})
	}
	return e.JSON(http.StatusOK, body)
}

func ReturnWriteResponse(e echo.Context, err error, body any) error {
	if err != nil {
		switch err {
		case gorm.ErrDuplicatedKey:
			return e.JSON(http.StatusConflict, struct {
				Message string `json:"message"`
			}{Message: "Duplicated key"})
		case gorm.ErrInvalidData:
			return e.JSON(http.StatusBadRequest, struct {
				Message string `json:"message"`
			}{Message: "Invalid request"})
		default:
			return e.JSON(http.StatusInternalServerError, struct {
				Message string `json:"message"`
			}{Message: err.Error()})
		}
	}
	return e.JSON(http.StatusOK, struct{ Body any }{Body: body})
}

func ValidateAndBind[T any](e echo.Context, input *T) error {
	if err := e.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := e.Validate(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, GetValidationFieldError(err))
	}
	return nil
}

func GetValidationFieldError(err error) map[string]string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMap := make(map[string]string)
		for _, fieldErr := range validationErrors {
			fieldName := strings.ToLower(fieldErr.Field())
			switch fieldErr.Tag() {
			case "required":
				errorMap[fieldName] = "This field is required"
			case "email":
				errorMap[fieldName] = "Invalid email address"
			case "e164":
				errorMap[fieldName] = "Invalid phone number"
			default:
				errorMap[fieldName] = "Invalid value"
			}
		}
		return errorMap
	}
	return nil
}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}
