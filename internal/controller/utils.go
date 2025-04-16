package controller

import (
	"errors"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"zeppelin/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
		if errors.Is(err, echo.ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, struct {
				Message string `json:"message"`
			}{Message: "Unauthorized"})
		}
		if err.Error() == "database error" {
			return echo.NewHTTPError(http.StatusInternalServerError, struct {
				Message string `json:"message"`
			}{Message: "database error"})
		}
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return echo.NewHTTPError(httpErr.Code, struct {
				Message interface{} `json:"message"`
			}{Message: httpErr.Message})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, struct {
			Message string `json:"message"`
		}{Message: "Internal server error"})
	}
	if body == nil || (reflect.ValueOf(body).Kind() == reflect.Ptr && reflect.ValueOf(body).IsNil()) {
		return echo.NewHTTPError(http.StatusNotFound, struct {
			Message string `json:"message"`
		}{Message: "Resource not found"})
	}
	return e.JSON(http.StatusOK, body)
}

type ErrorWithLocation struct {
	Err      error
	File     string
	Line     int
	Function string
}

func ReturnWriteResponse(e echo.Context, err error, body any) error {
	if err != nil {
		var file, function string
		var line int

		pc, filename, lineno, ok := runtime.Caller(1)
		if ok {
			file = filename
			line = lineno
			funcName := runtime.FuncForPC(pc).Name()
			function = funcName
		}
		shortFile := strings.Split(file, "/")[len(strings.Split(file, "/"))-1]

		e.Logger().Debugf("%s:%v | Error: %v at %s", shortFile, line, err, function)
		var numError *strconv.NumError
		var httpError *echo.HTTPError
		switch {
		case errors.Is(err, domain.ErrAuthorizationFailed):
			return e.JSON(http.StatusForbidden, errorResponse("Authorization failed"))
		case errors.Is(err, domain.ErrRequiredParamsMissing):
			return e.JSON(http.StatusBadRequest, errorResponse("Required parameters are missing"))
		case errors.Is(err, domain.ErrAuthTokenMissing):
			return e.JSON(http.StatusUnauthorized, errorResponse("Authorization token is missing"))
		case errors.Is(err, domain.ErrAuthTokenInvalid):
			return e.JSON(http.StatusUnauthorized, errorResponse("Authorization token is invalid"))
		case errors.Is(err, domain.ErrRoleExtractionFailed):
			return e.JSON(http.StatusUnauthorized, errorResponse("Token Invalid: Role extraction failed"))
		case errors.Is(err, domain.ErrResourceNotFound):
			return e.JSON(http.StatusNotFound, errorResponse("Resource not found"))
		case errors.Is(err, domain.ErrValidationFailed):
			return e.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return e.JSON(http.StatusConflict, errorResponse("Duplicated key"))
		case errors.Is(err, gorm.ErrInvalidData):
			return e.JSON(http.StatusBadRequest, errorResponse("Invalid request"))
		case errors.Is(err, numError):
			return e.JSON(http.StatusBadRequest, errorResponse("Invalid number format"))
		case errors.Is(err, gorm.ErrRecordNotFound):
			return e.JSON(http.StatusNotFound, errorResponse("Record not found"))
		case errors.Is(err, echo.ErrUnauthorized):
			return e.JSON(http.StatusUnauthorized, errorResponse("Unauthorized"))
		case errors.Is(err, httpError):
			return e.JSON(httpError.Code, errorResponse(httpError.Message.(string)))
		default:
			return e.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		}
	}

	e.Logger().Debugf("Response body: %v", body)
	return e.JSON(http.StatusOK, struct{ Body any }{Body: body})
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func errorResponse(message string) ErrorResponse {
	return ErrorResponse{Message: message}
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
