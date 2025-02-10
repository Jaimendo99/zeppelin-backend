package main

import (
	"zeppelin/usecases"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		e.StdLogger.Print("Hola my friend")
		sum := usecases.Add(20, 1)
		user := User{
			Name: "Anthony Cochea",
			Age:  sum,
		}
		return c.JSON(200, user)
	})

	e.Logger.Info(e.Start("0.0.0.0:3000"))
}

type User struct {
	Name string
	Age  int
}
