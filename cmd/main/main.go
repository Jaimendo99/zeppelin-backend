package main

import (
	"fmt"
	"log"
	"zeppelin/internal/config"
	"zeppelin/internal/db"
	"zeppelin/internal/routes"

	"github.com/labstack/echo/v4"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Fatal(err)
	}
	dns := config.GetConnectionString()
	err := db.InitDb(dns)
	if err != nil {
		fmt.Println("Error connecting to database: ")
	}
}

func main() {
	e := echo.New()
	routes.DefineRepresentativeRoutes(e)
	e.Logger.Info(e.Start("0.0.0.0:3000"))
}
