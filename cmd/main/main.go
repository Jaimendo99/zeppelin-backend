package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/routes"
	"zeppelin/internal/services"

	inMW "zeppelin/internal/middleware"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/labstack/echo/v4/middleware"

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

	if err := config.InitMQ(config.GetMQConnectionString()); err != nil {
		log.Fatalf("error connecting to message queue: %v", err)
	}

	if err := config.InitFCM(config.GetFirebaseConn()); err != nil {
		log.Fatalf("error connecting to firebase: %v", err)
	}

	config.InitSmtp(config.GetSmtpPassword())
	if err := config.CheckSmtpAuth(config.GetSmtpConfig()); err != nil {
		log.Fatalf("error authenticating smtp: %v", err)
	}

}

func main() {
	e := echo.New()
	e.Use(inMW.RequestLogger())
	e.Logger.SetOutput(os.Stdout)
	e.Logger.SetPrefix("[echo] ")

	// ✅ Agregar Middleware CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"}, // Permitir peticiones desde el frontend
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	auth, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error initializing AuthService: ", err)
	}

	e.GET("tokenFromSession", func(c echo.Context) error {
		sessionId := c.QueryParam("sessionId")
		template := c.QueryParam("template")

		if sessionId == "" || template == "" {
			return c.String(http.StatusBadRequest, "Missing sessionId or template query parameter")
		}

		urlPath := fmt.Sprintf("sessions/%s/tokens/%s", sessionId, template)

		req, err := auth.Client.NewRequest("POST", urlPath, nil)
		if err != nil {
			c.Logger().Errorf("Error creating Clerk request: %v", err)
			return c.String(http.StatusInternalServerError, "Error preparing token request")
		}

		var tokenResponse = clerk.SessionToken{}

		_, err = auth.Client.Do(req, &tokenResponse)
		if err != nil {
			c.Logger().Errorf("Error executing Clerk request or processing response: %v", err)
			return c.String(http.StatusInternalServerError, "Error creating token: "+err.Error())
		}

		if tokenResponse.JWT == "" {
			c.Logger().Warnf("Clerk token response successful but JWT was empty for session %s, template %s", sessionId, template)
			return c.String(http.StatusInternalServerError, "Failed to retrieve token content")
		}

		return c.JSON(http.StatusOK, tokenResponse)
	}, inMW.RoleMiddleware(auth, "org:admin", "org:teacher", "org:student"))

	routes.DefineRepresentativeRoutes(e)
	routes.DefineTeacherRoutes(e)
	routes.DefineStudentRoutes(e)
	routes.DefineCourseRoutes(e)
	routes.DefineAssignmentRoutes(e)
	routes.DefineNotificationRoutes(e)
	routes.DefineWebSocketRoutes(e)
	defer func(MQConn config.AmqpConnection) {
		err := MQConn.Close()
		if err != nil {

		}
	}(config.MQConn)

	if err := e.Start("0.0.0.0:8081"); err != nil {
		e.Logger.Fatal("error starting server: ", err)
	}

}
