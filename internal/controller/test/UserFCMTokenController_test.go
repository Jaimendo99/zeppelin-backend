package controller_test

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

// --- Mock del repositorio ---

type MockUserFcmTokenRepo struct {
	CreateUserFcmTokenFn        func(token domain.UserFcmTokenDb) error
	GetUserFcmTokensByUserIDFn  func(userID string) ([]domain.UserFcmTokenDb, error)
	DeleteUserFcmTokenByTokenFn func(firebaseToken string) error
	UpdateDeviceInfoFn          func(firebaseToken string, deviceInfo string) error
	UpdateFirebaseTokenFn       func(userID, deviceType, newToken string) error
}

func (m MockUserFcmTokenRepo) CreateUserFcmToken(token domain.UserFcmTokenDb) error {
	if m.CreateUserFcmTokenFn != nil {
		return m.CreateUserFcmTokenFn(token)
	}
	return errors.New("CreateUserFcmToken not implemented")
}

func (m MockUserFcmTokenRepo) GetUserFcmTokensByUserID(userID string) ([]domain.UserFcmTokenDb, error) {
	if m.GetUserFcmTokensByUserIDFn != nil {
		return m.GetUserFcmTokensByUserIDFn(userID)
	}
	return nil, errors.New("GetUserFcmTokensByUserID not implemented")
}

func (m MockUserFcmTokenRepo) DeleteUserFcmTokenByToken(firebaseToken string) error {
	if m.DeleteUserFcmTokenByTokenFn != nil {
		return m.DeleteUserFcmTokenByTokenFn(firebaseToken)
	}
	return errors.New("DeleteUserFcmTokenByToken not implemented")
}

func (m MockUserFcmTokenRepo) UpdateDeviceInfo(firebaseToken string, deviceInfo string) error {
	if m.UpdateDeviceInfoFn != nil {
		return m.UpdateDeviceInfoFn(firebaseToken, deviceInfo)
	}
	return errors.New("UpdateDeviceInfo not implemented")
}

func (m MockUserFcmTokenRepo) UpdateFirebaseToken(userID string, firebaseToken string, deviceInfo string) error {
	if m.UpdateFirebaseTokenFn != nil {
		return m.UpdateFirebaseTokenFn(userID, firebaseToken, deviceInfo)
	}
	return errors.New("UpdateDeviceInfo not implemented")
}

func TestCreateUserFcmToken_Success(t *testing.T) {
	tokenJSON := `{"firebase_token":"token123","device_type":"MOBILE","device_info":"iPhone 14"}`
	testUserID := "user-123"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/tokens", strings.NewReader(tokenJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserFcmTokenRepo{
		CreateUserFcmTokenFn: func(token domain.UserFcmTokenDb) error {
			assert.Equal(t, testUserID, token.UserID)
			assert.Equal(t, "token123", token.FirebaseToken)
			return nil
		},
	}

	ctrl := controller.UserFcmTokenController{Repo: mockRepo}
	handler := ctrl.CreateUserFcmToken()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"Body":{"message":"Token registrado con éxito"}}`
		assert.JSONEq(t, expected, rec.Body.String())
	}
}

func TestGetUserFcmTokens_Success(t *testing.T) {
	testUserID := "user-456"
	mockTokens := []domain.UserFcmTokenDb{
		{TokenID: 1, UserID: testUserID, FirebaseToken: "token1", DeviceType: "MOBILE"},
		{TokenID: 2, UserID: testUserID, FirebaseToken: "token2", DeviceType: "WEB"},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tokens", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	mockRepo := MockUserFcmTokenRepo{
		GetUserFcmTokensByUserIDFn: func(userID string) ([]domain.UserFcmTokenDb, error) {
			assert.Equal(t, testUserID, userID)
			return mockTokens, nil
		},
	}

	ctrl := controller.UserFcmTokenController{Repo: mockRepo}
	handler := ctrl.GetUserFcmTokens()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `[
			{"token_id":1,"user_id":"user-456","firebase_token":"token1","device_type":"MOBILE","device_info":"","updated_at":""},
			{"token_id":2,"user_id":"user-456","firebase_token":"token2","device_type":"WEB","device_info":"","updated_at":""}
		]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestDeleteUserFcmToken_Success(t *testing.T) {
	deleteJSON := `{"firebase_token":"token123"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/tokens", strings.NewReader(deleteJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserFcmTokenRepo{
		DeleteUserFcmTokenByTokenFn: func(firebaseToken string) error {
			assert.Equal(t, "token123", firebaseToken)
			return nil
		},
	}

	ctrl := controller.UserFcmTokenController{Repo: mockRepo}
	handler := ctrl.DeleteUserFcmToken()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"Body":{"message":"Token eliminado con éxito"}}`
		assert.JSONEq(t, expected, rec.Body.String())
	}
}

func TestUpdateDeviceInfo_Success(t *testing.T) {
	updateJSON := `{"firebase_token":"token123","device_info":"Updated Info"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/tokens/device-info", strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserFcmTokenRepo{
		UpdateDeviceInfoFn: func(firebaseToken, deviceInfo string) error {
			assert.Equal(t, "token123", firebaseToken)
			assert.Equal(t, "Updated Info", deviceInfo)
			return nil
		},
	}

	ctrl := controller.UserFcmTokenController{Repo: mockRepo}
	handler := ctrl.UpdateDeviceInfo()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"Body":{"message":"Información del dispositivo actualizada con éxito"}}`
		assert.JSONEq(t, expected, rec.Body.String())
	}
}
func TestUpdateWebToken_Success(t *testing.T) {
	updateJSON := `{"firebase_token":"new_web_token"}`

	testUserID := "user-789"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/fcm/token/web", strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserFcmTokenRepo{
		UpdateDeviceInfoFn: func(firebaseToken, deviceInfo string) error {
			assert.Fail(t, "UpdateDeviceInfo should not be called in UpdateWebToken")
			return nil
		},
		// Simulamos UpdateFirebaseToken (deberías agregarlo en el mock para producción, aquí lo simulamos inline)
	}
	mockRepo.UpdateFirebaseTokenFn = func(userID, deviceType, newToken string) error {
		assert.Equal(t, testUserID, userID)
		assert.Equal(t, "WEB", deviceType)
		assert.Equal(t, "new_web_token", newToken)
		return nil
	}

	ctrl := controller.UserFcmTokenController{Repo: mockRepo}
	handler := ctrl.UpdateWebToken()

	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"Body":{"message":"Firebase token WEB actualizado con éxito"}}`
		assert.JSONEq(t, expected, rec.Body.String())
	}
}

func TestUpdateMobileToken_Success(t *testing.T) {
	updateJSON := `{"firebase_token":"new_mobile_token"}`

	testUserID := "user-789"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/fcm/token/mobile", strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserFcmTokenRepo{
		UpdateDeviceInfoFn: func(firebaseToken, deviceInfo string) error {
			assert.Fail(t, "UpdateDeviceInfo should not be called in UpdateMobileToken")
			return nil
		},
	}
	mockRepo.UpdateFirebaseTokenFn = func(userID, deviceType, newToken string) error {
		assert.Equal(t, testUserID, userID)
		assert.Equal(t, "MOBILE", deviceType)
		assert.Equal(t, "new_mobile_token", newToken)
		return nil
	}

	ctrl := controller.UserFcmTokenController{Repo: mockRepo}
	handler := ctrl.UpdateMobileToken()

	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"Body":{"message":"Firebase token MOBILE actualizado con éxito"}}`
		assert.JSONEq(t, expected, rec.Body.String())
	}
}
