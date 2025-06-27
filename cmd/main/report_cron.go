// cmd/cron/report_cron.go
package main

import (
	"fmt"
	"log"
	"time"
	"zeppelin/internal/config"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type ReportCronService struct {
	cron      *cron.Cron
	resend    *config.ResendService
	repo      domain.RepresentativeRepo
	secretKey string
}

func NewReportCronService(db *gorm.DB) (*ReportCronService, error) {
	resendService, err := config.InitResend()
	if err != nil {
		return nil, fmt.Errorf("error inicializando Resend: %v", err)
	}

	repo := data.NewRepresentativeRepo(db)
	secretKey := "zeppelin_weekly_report_2025"

	service := &ReportCronService{
		cron:      cron.New(),
		resend:    resendService,
		repo:      repo,
		secretKey: secretKey,
	}

	// Programar para domingos a las 10:00 AM
	_, err = service.cron.AddFunc("0 10 * * 0", service.sendWeeklyReports)
	if err != nil {
		return nil, fmt.Errorf("error programando cron: %v", err)
	}

	//  líneas de prueba:
	// go func() {
	//  log.Println("Ejecutando envío de reportes inmediatamente para pruebas...")
	//  service.sendWeeklyReports()
	// }()

	return service, nil
}
func (r *ReportCronService) Start() {
	log.Println("Iniciando servicio de reportes semanales...")
	r.cron.Start()
}

func (r *ReportCronService) Stop() {
	log.Println("Deteniendo servicio de reportes semanales...")
	r.cron.Stop()
}

func (r *ReportCronService) sendWeeklyReports() {
	log.Println("Iniciando envío de reportes semanales...")

	representatives, err := r.repo.GetAllRepresentatives()
	if err != nil {
		log.Printf("Error obteniendo representantes: %v", err)
		return
	}

	currentDate := time.Now().Format("2006-01-02")
	successCount := 0
	errorCount := 0

	for _, rep := range representatives {
		if rep.Email == "" || rep.UserID == "" {
			log.Printf("Representante %d sin email o UserID válido, omitiendo...", rep.RepresentativeId)
			continue
		}

		token := r.generateSimpleToken(rep.UserID, currentDate)
		reportURL := fmt.Sprintf("https://www.focused.uno/report/weekly/%s/%s?token=%s",
			rep.UserID, currentDate, token)

		err := r.sendWeeklyReportEmail(rep, reportURL)
		if err != nil {
			log.Printf("Error enviando correo a %s: %v", rep.Email, err)
			errorCount++
		} else {
			log.Printf("Reporte semanal enviado exitosamente a %s", rep.Email)
			successCount++
		}
	}

	log.Printf("Envío de reportes completado. Exitosos: %d, Errores: %d", successCount, errorCount)
}

func (r *ReportCronService) generateSimpleToken(userID, date string) string {
	combined := fmt.Sprintf("%s_%s_%s", userID, date, r.secretKey)
	var sum int
	for _, b := range []byte(combined) {
		sum += int(b)
	}
	return fmt.Sprintf("wkly_%x_%d", sum, len(combined))
}

func (r *ReportCronService) sendWeeklyReportEmail(rep domain.Representative, reportURL string) error {
	subject := "📊 Reporte Semanal de Progreso - Zeppelin"
	message := fmt.Sprintf(`¡Hola %s!

Esperamos que tengas una excelente semana. Nos complace informarte que el reporte semanal de progreso académico de tu hijo/a ya está disponible.

🎯 ¿Qué encontrarás en el reporte?
• Resumen detallado de actividades completadas
• Progreso en cursos y asignaciones
• Tiempo dedicado al estudio (sesiones Pomodoro)
• Logros y áreas de mejora identificadas
• Tendencias de rendimiento semanal

📈 Para revisar el reporte completo, simplemente haz clic en el siguiente enlace:
%s

Este reporte te ayudará a mantenerte al tanto del desarrollo académico y identificar oportunidades para apoyar el aprendizaje.

Si tienes alguna pregunta sobre el reporte o necesitas asistencia, no dudes en contactarnos.

¡Gracias por confiar en Zeppelin para el crecimiento educativo!

Saludos cordiales,
El equipo de Zeppelin
🚀 Impulsando el futuro de la educación`, rep.Name, reportURL)

	return r.resend.SendWeeklyReportEmail(rep.Email, subject, message)
}
