package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	resendServiceInstance *ResendService
	resendOnce            sync.Once
	resendInitErr         error
)

type ResendService struct {
	APIKey string
}

func InitResend() (*ResendService, error) {
	resendOnce.Do(func() {
		_ = godotenv.Load() // Carga el archivo .env si existe (útil para local)

		apiKey := os.Getenv("RESEND_API_KEY")
		if apiKey == "" {
			resendInitErr = fmt.Errorf("RESEND_API_KEY no está definido en variables de entorno")
			return
		}

		resendServiceInstance = &ResendService{
			APIKey: apiKey,
		}
	})

	return resendServiceInstance, resendInitErr
}

func (r *ResendService) SendParentalConsentEmail(toEmail string, token string) error {
	url := "https://api.resend.com/emails"

	body := map[string]interface{}{
		"from":    "Zeppelin <onboarding@message.focused.uno>",
		"to":      []string{toEmail},
		"subject": "Consentimiento parental requerido",
		"text": fmt.Sprintf(`Hola,
			Por favor da tu consentimiento para que tu hijo/a use la plataforma Zeppelin.
			Haz clic en el siguiente enlace para revisar y aceptar los términos:
			https://www.focused.uno/consentimiento-parental?token=%s
			Gracias.`, token),
	}

	jsonData, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("error al enviar correo con Resend: %s", resp.Status)
}
