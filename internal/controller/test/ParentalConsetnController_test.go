package controller_test

import (
	"bytes"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConsentRepo struct {
	GetConsentByTokenMock   func(token string) (*domain.ParentalConsent, error)
	UpdateConsentStatusMock func(token, status, ip, userAgent string) error
	GetConsentByUserIDMock  func(userID string) (*domain.ParentalConsent, error)
}

func (m mockConsentRepo) GetConsentByToken(token string) (*domain.ParentalConsent, error) {
	return m.GetConsentByTokenMock(token)
}

func (m mockConsentRepo) UpdateConsentStatus(token, status, ip, userAgent string) error {
	return m.UpdateConsentStatusMock(token, status, ip, userAgent)
}

func (m mockConsentRepo) GetConsentByUserID(userID string) (*domain.ParentalConsent, error) {
	return m.GetConsentByUserIDMock(userID)
}

func (m mockConsentRepo) CreateConsent(consent domain.ParentalConsent) error {
	panic("not implemented")
}

func TestParentalConsentController_GetConsentByToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/consent?token=test-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/consent")
	c.QueryParams().Add("token", "test-token")

	mockConsent := &domain.ParentalConsent{
		ConsentID: 1, UserID: "user-1", Token: "test-token", Status: "PENDING",
	}
	ctrl := controller.ParentalConsentController{
		Repo: mockConsentRepo{
			GetConsentByTokenMock: func(token string) (*domain.ParentalConsent, error) {
				assert.Equal(t, "test-token", token)
				return mockConsent, nil
			},
		},
	}

	handler := ctrl.GetConsentByToken()
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token":"test-token"`)
	assert.Contains(t, rec.Body.String(), `"status":"PENDING"`)
}

func TestParentalConsentController_UpdateConsentStatus(t *testing.T) {
	e := echo.New()
	input := map[string]string{
		"token":  "abc123",
		"status": "ACCEPTED",
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/consent/update", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctrl := controller.ParentalConsentController{
		Repo: mockConsentRepo{
			UpdateConsentStatusMock: func(token, status, ip, userAgent string) error {
				assert.Equal(t, "abc123", token)
				assert.Equal(t, "ACCEPTED", status)
				return nil
			},
		},
	}

	e.Validator = &CustomValidator{Validator: validator.New()}

	handler := ctrl.UpdateConsentStatus()
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"message":"Consentimiento aprobado exitosamente"`)
	assert.Contains(t, rec.Body.String(), `"status":"ACCEPTED"`)
}

func TestParentalConsentController_GetConsentByUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/consent/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-abc")

	mockConsent := &domain.ParentalConsent{
		ConsentID: 2, UserID: "user-abc", Token: "tok-2", Status: "REJECTED",
	}
	ctrl := controller.ParentalConsentController{
		Repo: mockConsentRepo{
			GetConsentByUserIDMock: func(userID string) (*domain.ParentalConsent, error) {
				assert.Equal(t, "user-abc", userID)
				return mockConsent, nil
			},
		},
	}

	handler := ctrl.GetConsentByUserID()
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token":"tok-2"`)
	assert.Contains(t, rec.Body.String(), `"status":"REJECTED"`)
}
