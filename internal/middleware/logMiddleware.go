package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
)

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			stop := time.Now()
			method := c.Request().Method
			path := c.Path()
			statusCode := c.Response().Status
			latency := stop.Sub(start)
			clientIP := c.RealIP()

			e := c.Echo() // Get the Echo instance from the context
			e.Logger.Infof(
				"%s %s %s %d %s %s",
				clientIP,
				method,
				path,
				statusCode,
				latency,
				time.Now().Format(time.RFC3339),
			)

			return err
		}
	}
}
