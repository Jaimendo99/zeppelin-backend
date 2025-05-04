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

// MockUserPomodoroRepo defines a mock for the UserPomodoroRepo interface
type MockUserPomodoroRepo struct {
	GetByUserIDFn    func(userID string) (*domain.UserPomodoro, error)
	UpdateByUserIDFn func(userID string, input domain.UpdatePomodoroInput) error
}

func (m MockUserPomodoroRepo) GetByUserID(userID string) (*domain.UserPomodoro, error) {
	if m.GetByUserIDFn != nil {
		return m.GetByUserIDFn(userID)
	}
	return nil, errors.New("GetByUserID not implemented")
}

func (m MockUserPomodoroRepo) UpdateByUserID(userID string, input domain.UpdatePomodoroInput) error {
	if m.UpdateByUserIDFn != nil {
		return m.UpdateByUserIDFn(userID, input)
	}
	return errors.New("UpdateByUserID not implemented")
}

func TestGetPomodoroByUserID_Success(t *testing.T) {
	testUserID := "user-123"
	mockPomodoro := &domain.UserPomodoro{
		PomodoroID:   1,
		UserID:       testUserID,
		ActiveTime:   25,
		RestTime:     5,
		LongRestTime: 15,
		Iterations:   4,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pomodoro", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	mockRepo := MockUserPomodoroRepo{
		GetByUserIDFn: func(userID string) (*domain.UserPomodoro, error) {
			assert.Equal(t, testUserID, userID)
			return mockPomodoro, nil
		},
	}

	ctrl := controller.PomodoroController{Repo: mockRepo}
	handler := ctrl.GetPomodoroByUserID()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{
			"pomodoro_id": 1,
			"user_id": "user-123",
			"active_time": 25,
			"rest_time": 5,
			"long_rest_time": 15,
			"iterations": 4
		}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetPomodoroByUserID_RepoError(t *testing.T) {
	testUserID := "user-123"
	repoErr := errors.New("database error")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/pomodoro", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)

	mockRepo := MockUserPomodoroRepo{
		GetByUserIDFn: func(userID string) (*domain.UserPomodoro, error) {
			assert.Equal(t, testUserID, userID)
			return nil, repoErr
		},
	}

	ctrl := controller.PomodoroController{Repo: mockRepo}
	handler := ctrl.GetPomodoroByUserID()

	err := handler(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Expected HTTPError")
		assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	}
}

func TestUpdatePomodoroByUserID_Success(t *testing.T) {
	updateJSON := `{
		"active_time": 30,
		"rest_time": 10,
		"long_rest_time": 20,
		"iterations": 5
	}`
	testUserID := "user-123"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/pomodoro", strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserPomodoroRepo{
		UpdateByUserIDFn: func(userID string, input domain.UpdatePomodoroInput) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, 30, input.ActiveTime)
			assert.Equal(t, 10, input.RestTime)
			assert.Equal(t, 20, input.LongRestTime)
			assert.Equal(t, 5, input.Iterations)
			return nil
		},
	}

	ctrl := controller.PomodoroController{Repo: mockRepo}
	handler := ctrl.UpdatePomodoroByUserID()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{"Body":{"message":"Configuración actualizada"}}`
		assert.JSONEq(t, expected, rec.Body.String())
	}
}

func TestUpdatePomodoroByUserID_ValidationError(t *testing.T) {
	// Use malformed JSON to ensure validation failure
	invalidJSON := `{` // Incomplete JSON
	testUserID := "user-123"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/pomodoro", strings.NewReader(invalidJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserPomodoroRepo{
		UpdateByUserIDFn: func(userID string, input domain.UpdatePomodoroInput) error {
			assert.Fail(t, "UpdateByUserID should not be called on validation error")
			return nil
		},
	}

	ctrl := controller.PomodoroController{Repo: mockRepo}
	handler := ctrl.UpdatePomodoroByUserID()

	err := handler(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Expected HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	}
}

func TestUpdatePomodoroByUserID_RepoError(t *testing.T) {
	updateJSON := `{
		"active_time": 30,
		"rest_time": 10,
		"long_rest_time": 20,
		"iterations": 5
	}`
	testUserID := "user-123"
	repoErr := errors.New("database error")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/pomodoro", strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockUserPomodoroRepo{
		UpdateByUserIDFn: func(userID string, input domain.UpdatePomodoroInput) error {
			assert.Equal(t, testUserID, userID)
			return repoErr
		},
	}

	ctrl := controller.PomodoroController{Repo: mockRepo}
	handler := ctrl.UpdatePomodoroByUserID()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"message":"database error"}`, rec.Body.String())
	}
}
