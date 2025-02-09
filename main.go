package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		e.StdLogger.Print("Hola my friend")
		user := User{
			Name: "Jon",
			Age:  20,
		}
		return c.JSON(200, user)
	})

	e.Logger.Info(e.Start("0.0.0.0:3000"))
}

type User struct {
	Name string
	Age  int
}
