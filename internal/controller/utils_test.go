package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestReturnReadResponse_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	body := map[string]string{"foo": "bar"}
	err := ReturnReadResponse(c, nil, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "bar", resp["foo"])
}

func TestReturnReadResponse_InvalidData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ReturnReadResponse(c, gorm.ErrInvalidData, nil)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)

	// Convert the error message (a struct) to JSON and then to a map for testing.
	var er struct {
		Message string `json:"message"`
	}
	b, marshalErr := json.Marshal(httpErr.Message)
	assert.NoError(t, marshalErr)
	unmarshalErr := json.Unmarshal(b, &er)
	assert.NoError(t, unmarshalErr)
	assert.Equal(t, "Invalid request", er.Message)
}

func TestReturnReadResponse_RecordNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ReturnReadResponse(c, gorm.ErrRecordNotFound, nil)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)

	var er struct {
		Message string `json:"message"`
	}
	b, _ := json.Marshal(httpErr.Message)
	_ = json.Unmarshal(b, &er)
	assert.Equal(t, "Record not found", er.Message)
}

func TestReturnReadResponse_GenericError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	genericErr := errors.New("something went wrong")
	err := ReturnReadResponse(c, genericErr, nil)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)

	var er struct {
		Message string `json:"message"`
	}
	b, _ := json.Marshal(httpErr.Message)
	_ = json.Unmarshal(b, &er)
	assert.Equal(t, "Internal server error", er.Message)
}

// ---
// Tests for ReturnWriteResponse

func TestReturnWriteResponse_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	body := map[string]string{"result": "ok"}
	err := ReturnWriteResponse(c, nil, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// The response is wrapped in a struct: { "body": ... }
	var resp struct {
		Body any `json:"body"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Since JSON unmarshalling into "any" yields map[string]interface{} for objects,
	// do a type assertion.
	if m, ok := resp.Body.(map[string]interface{}); ok {
		assert.Equal(t, "ok", m["result"])
	}
}

func TestReturnWriteResponse_InvalidData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = ReturnWriteResponse(c, gorm.ErrInvalidData, nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var er struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &er)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request", er.Message)
}

func TestReturnWriteResponse_DuplicatedKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = ReturnWriteResponse(c, gorm.ErrDuplicatedKey, nil)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var er struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &er)
	assert.Equal(t, "Duplicated key", er.Message)
}

func TestReturnWriteResponse_GenericError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	genericErr := errors.New("generic error")
	_ = ReturnWriteResponse(c, genericErr, nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var er struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &er)
	assert.Equal(t, "generic error", er.Message)
}

// ---
// Tests for ValidateAndBind

// Define a test struct with a required field.
type TestRequest struct {
	Field string `json:"field" validate:"required"`
}

func TestValidateAndBind_Success(t *testing.T) {
	e := echo.New()
	// Provide a valid JSON body.
	reqBody := `{"field": "value"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up the custom validator.
	e.Validator = &CustomValidator{Validator: validator.New()}

	var input TestRequest
	err := ValidateAndBind(c, &input)
	assert.NoError(t, err)
	assert.Equal(t, "value", input.Field)
}

func TestValidateAndBind_InvalidJSON(t *testing.T) {
	e := echo.New()
	// Provide an invalid JSON body.
	reqBody := `{"field": }`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var input TestRequest
	err := ValidateAndBind(c, &input)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Equal(t, "Invalid request body", httpErr.Message)
}

func TestValidateAndBind_ValidationError(t *testing.T) {
	e := echo.New()
	// Provide a valid JSON but missing the required field.
	reqBody := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up the custom validator.
	e.Validator = &CustomValidator{Validator: validator.New()}

	var input TestRequest
	err := ValidateAndBind(c, &input)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)

	// The error message is a map[string]string with our field errors.
	msg, ok := httpErr.Message.(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "This field is required", msg["field"])
}

// ---
// Test for GetValidationFieldError

func TestGetValidationFieldError(t *testing.T) {
	// Define a struct with several validation tags.
	type TestStruct struct {
		Email string `validate:"required,email"`
		Phone string `validate:"e164"`
		Age   int    `validate:"min=18"`
	}

	// Create an instance with invalid values.
	ts := TestStruct{
		Email: "invalid", // not a valid email
		Phone: "1234",    // not a valid E.164 phone number
		Age:   10,        // less than minimum 18
	}

	v := validator.New()
	err := v.Struct(ts)
	assert.Error(t, err)

	errorMap := GetValidationFieldError(err)
	// Check that the error map contains expected messages.
	assert.Equal(t, "Invalid email address", errorMap["email"])
	assert.Equal(t, "Invalid phone number", errorMap["phone"])
	assert.Equal(t, "Minimum value is 18", errorMap["age"])
}
