package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		e.StdLogger.Print("Hola my friend")
		return c.String(200, "Hello, World!")
	})

	e.Logger.Info(e.Start("0.0.0.0:3000"))
}
