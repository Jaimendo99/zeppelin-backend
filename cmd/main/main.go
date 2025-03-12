package main

import (
	"github.com/labstack/echo/v4/middleware"
	"log"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/routes"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	if err := config.InitDb(config.GetConnectionString()); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	/*if err := config.InitMQ(config.GetMQConnectionString()); err != nil {
		log.Fatalf("error connecting to message queue: %v", err)
	}

	if err := config.InitFCM(config.GetFirebaseConn()); err != nil {
		log.Fatalf("error connecting to firebase: %v", err)
	}

	config.InitSmtp(config.GetSmtpPassword())
	if err := config.CheckSmtpAuth(config.GetSmtpConfig()); err != nil {
		log.Fatalf("error authenticating smtp: %v", err)
	}*/

}

func main() {
	e := echo.New()
	// ✅ Agregar Middleware CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"}, // Permitir peticiones desde el frontend
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	routes.DefineRepresentativeRoutes(e)
	routes.DefineTeacherRoutes(e)
	routes.DefineStudentRoutes(e)
	routes.DefineCourseRoutes(e)
	//routes.DefineNotificationRoutes(e)

	defer config.MQConn.Close()

	if err := e.Start("0.0.0.0:8080"); err != nil {
		e.Logger.Fatal("error starting server: ", err)
	}

}
