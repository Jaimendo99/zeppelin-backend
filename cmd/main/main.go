package main

import (
	"fmt"
	"log"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/routes"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Print(err)
	}
	dns := config.GetConnectionString()
	err := config.InitDb(dns)
	if err != nil {
		fmt.Println("Error connecting to database: ")
	}
}

func main() {
	e := echo.New()
	e.Validator = &controller.CustomValidator{Validator: validator.New()}
	routes.DefineRepresentativeRoutes(e)
	e.Logger.Info(e.Start("0.0.0.0:3000"))
}
