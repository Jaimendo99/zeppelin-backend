package main

import (
	"log"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/db"
	"zeppelin/internal/routes"
	"zeppelin/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	dns := config.GetConnectionString()
	if err := config.InitDb(dns); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err := config.InitMQ(config.GetMQConnectionString()); err != nil {
		log.Fatalf("error connecting to message queue: %v", err)
	} else {
		log.Println("connected to message queue")
	}
}

func main() {
	e := echo.New()
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	routes.DefineRepresentativeRoutes(e)
	routes.DefineNotificationRoutes(e)

	notificationMq := db.NewNotificationMq(config.ProducerChannel, services.NotificationPrinter{})

	go func() {
		if err := notificationMq.ConsumeFromQueue("notification"); err != nil {
			e.Logger.Error("error consuming queue: ", err)
		}
	}()

	defer config.MQConn.Close()

	if err := e.Start("0.0.0.0:8080"); err != nil {
		e.Logger.Fatal("error starting server: ", err)
	}
}
