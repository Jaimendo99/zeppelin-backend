package middleware_test

import (
	"bytes"
	"github.com/labstack/echo/v4"
	echoMW "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeppelin/internal/middleware" // Import your actual middleware package as a different name
)

// customLogger implements echo.Logger interface
type customLogger struct {
	*log.Logger
	buf *bytes.Buffer
}

func newCustomLogger() (*customLogger, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	l := log.New("test")
	l.SetOutput(buf)
	return &customLogger{Logger: l, buf: buf}, buf
}

func (l *customLogger) Output() io.Writer {
	return l.buf
}

func TestRequestLogger(t *testing.T) {
	// Create a custom logger with buffer
	logger, buf := newCustomLogger()

	// Create new Echo instance with custom logger
	e := echo.New()
	e.Logger = logger
	e.Use(middleware.RequestLogger())

	// Define a test handler
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()

	// Perform the request
	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check if log contains expected elements
	logOutput := buf.String()
	t.Logf("Log output: %s", logOutput)
	assert.Contains(t, logOutput, "192.168.1.1")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "200")
}

func TestRequestLoggerWithError(t *testing.T) {
	// Create a custom logger with buffer
	logger, buf := newCustomLogger()

	// Create new Echo instance with custom logger
	e := echo.New()
	e.Logger = logger

	// Important: Configure Echo to handle errors properly
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.JSON(code, map[string]interface{}{
			"error": err.Error(),
		})
	}

	e.Use(middleware.RequestLogger())

	// Define a test handler that returns an error
	e.POST("/test-error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodPost, "/test-error", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()

	// Perform the request
	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Check if log contains expected elements
	logOutput := buf.String()
	t.Logf("Log output: %s", logOutput)
	assert.Contains(t, logOutput, "192.168.1.1")
	assert.Contains(t, logOutput, "POST")
	assert.Contains(t, logOutput, "/test-error")
}

func TestRequestLoggerWithPanic(t *testing.T) {
	// Create a custom logger with buffer
	logger, buf := newCustomLogger()

	// Create new Echo instance with custom logger
	e := echo.New()
	e.Logger = logger
	e.Use(middleware.RequestLogger())
	// Use the correct middleware from Echo's middleware package
	e.Use(echoMW.Recover())

	// Define a test handler that panics
	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()

	// Perform the request
	e.ServeHTTP(rec, req)

	// Check if log contains expected elements
	logOutput := buf.String()
	t.Logf("Log output: %s", logOutput)
	assert.Contains(t, logOutput, "192.168.1.1")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/panic")
	// Status code should be 500 due to panic
	assert.Contains(t, logOutput, "500")
}
