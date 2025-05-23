package controller_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

func TestForceString(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		assert.Equal(t, "hello", controller.ForceString("hello"))
	})
	t.Run("[]byte input", func(t *testing.T) {
		assert.Equal(t, "world", controller.ForceString([]byte("world")))
	})
	t.Run("int input", func(t *testing.T) {
		assert.Equal(t, "42", controller.ForceString(42))
	})
}

func TestGenerateUID(t *testing.T) {
	uid := controller.GenerateUID()
	assert.NotEmpty(t, uid)
	assert.Len(t, strings.Split(uid, "-"), 5)
}

func TestReturnWriteResponse_Error_ValidationFailed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	err := controller.ReturnWriteResponse(ctx, domain.ErrValidationFailed, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Validation failed")
}

func TestReturnWriteResponse_Error_AuthorizationFailed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	err := controller.ReturnWriteResponse(ctx, domain.ErrAuthorizationFailed, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Authorization failed")
}

func TestReturnWriteResponse_Error_HTTPError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	httpErr := echo.NewHTTPError(http.StatusNotFound, "custom message")
	err := controller.ReturnWriteResponse(ctx, httpErr, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "custom message")
}

func TestReturnWriteResponse_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	body := map[string]string{"status": "ok"}
	err := controller.ReturnWriteResponse(ctx, nil, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "status")
}
