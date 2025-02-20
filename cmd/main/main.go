package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/routes"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	if err := config.InitDb(config.GetConnectionString()); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err := config.InitMQ(config.GetMQConnectionString()); err != nil {
		log.Fatalf("error connecting to message queue: %v", err)
	}

	config.InitSmtp(config.GetSmtpPassword())
	// if err := config.CheckSmtpAuth(config.GetSmtpConfig()); err != nil {
	// 	log.Fatalf("error authenticating smtp: %v", err)
	// }

}

func main() {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}
	log.Println("Working directory:", wd)

	ctx := context.Background()
	opt := option.WithCredentialsFile("firebase-conn.json")
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "zeppelin-app-859e0", // Replace with your actual project ID.
	}, opt)
	if err != nil {
		log.Fatal("Error creating new firebase app")
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}
	registrationToken := "cMJMNKEcSqWqTHxA-DW7VL:APA91bFJxQDoMlJssjjTj45KKLu91RzMFzlDN1llQ_OBDqaky07jrGjUNp0tQUTnFppU-7wICkipnCZUl-doiX3PJrRtUGn6NapTJQRDyaZE_I4ClTeDUf4"

	// Define the message payload.
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "Hello from Go!",
			Body:  "This is a push notification sent from the Go server.",
		},
		// You can also use Data messages:
		Data: map[string]string{
			"title": "Esto es una prueba",
			"body":  "Esto es una prueba de notificación",
		},
		Token: registrationToken,
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalf("error sending message: %v", err)
	}

	fmt.Println("Successfully sent message:", response)

	e := echo.New()
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	routes.DefineRepresentativeRoutes(e)
	routes.DefineNotificationRoutes(e)

	defer config.MQConn.Close()

	if err := e.Start("0.0.0.0:8080"); err != nil {
		e.Logger.Fatal("error starting server: ", err)
	}
}
