package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"zeppelin/internal/config"
	"zeppelin/internal/controller"
	"zeppelin/internal/routes"
	"zeppelin/internal/services"

	inMW "zeppelin/internal/middleware"

	"github.com/labstack/echo/v4/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	elog "github.com/labstack/gommon/log"
)

var globalReportCron *ReportCronService

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

	err := config.InitR2()
	if err != nil {
		log.Fatalf("Error al inicializar R2: %v", err)
	}

	// Usar el cliente para listar objetos
	output, err := config.R2Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String("zeppelin"),
	})
	if err != nil {
		log.Fatalf("Error al listar objetos: %v", err)
	}

	for _, obj := range output.Contents {
		fmt.Println("Archivo:", *obj.Key)
	}

	_, err = config.InitResend()

	_, err = config.InitResend()
	if err != nil {
		log.Fatalf("error inicializando Resend: %v", err)
	}

	// Inicializar el servicio de reportes
	reportCron, err := NewReportCronService(config.DB)
	if err != nil {
		log.Fatalf("Error inicializando servicio de reportes: %v", err)
	}
	reportCron.Start()

	// Guardar referencia global para poder detenerlo
	globalReportCron = reportCron

}

func main() {
	e := echo.New()
	e.Logger.SetHeader("[echo-zeppelin] | ${time_rfc3339} | ${level}${message} | ")
	e.Logger.SetLevel(elog.DEBUG) // Set the desired log level
	e.Use(inMW.RequestLogger())
	e.Logger.SetOutput(e.Logger.Output())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173",
			"https://www.focused.uno",
		},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	auth, err := services.NewAuthService()
	if err != nil {
		e.Logger.Fatal("Error initializing AuthService: ", err)
	}

	roleMiddlewareProvider := func(roles ...string) echo.MiddlewareFunc {
		return inMW.RoleMiddleware(auth, roles...)
	}

	routes.DefineRepresentativeRoutes(e, roleMiddlewareProvider)
	routes.DefineTeacherRoutes(e, auth, roleMiddlewareProvider)
	routes.DefineStudentRoutes(e)
	routes.DefineCourseRoutes(e, auth, roleMiddlewareProvider)
	routes.DefineAssignmentRoutes(e, roleMiddlewareProvider)
	routes.DefineNotificationRoutes(e, roleMiddlewareProvider)
	routes.DefineWebSocketRoutes(e, auth)
	routes.DefineCourseContentRoutes(e)
	routes.DefineAuthRoutes(e, auth.Clerk, roleMiddlewareProvider)
	routes.DefineUserFcmTokenRoutes(e, auth, roleMiddlewareProvider)
	routes.DefinePomodoroRoutes(e, auth, roleMiddlewareProvider)
	routes.DefineQuizAnswerRoutes(e, auth, roleMiddlewareProvider)
	routes.DefineParentalConsentRoutes(e)
	defer func(MQConn config.AmqpConnection) {
		err := MQConn.Close()
		if err != nil {

		}
	}(config.MQConn)

	port := config.GetPort()
	if port == "" {
		port = "3000"
	}
	defer func() {
		if globalReportCron != nil {
			globalReportCron.Stop()
		}
	}()
	e.Logger.Error(e.Start("0.0.0.0:" + port))

}
