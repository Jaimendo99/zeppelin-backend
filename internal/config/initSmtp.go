package config

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"zeppelin/internal/domain"
)

var smtpServer *domain.SmtpConfig

type SmtpClient interface {
	StartTLS(*tls.Config) error
	Auth(smtp.Auth) error
	Close() error
}

var smtpDial func(addr string) (SmtpClient, error)

func init() {
	smtpDial = func(addr string) (SmtpClient, error) {
		client, err := smtp.Dial(addr)
		if err != nil {
			var nilClient SmtpClient = nil // Return typed nil interface on error
			return nilClient, err
		}
		return client, nil
	}
}

var SmtpDial = &smtpDial

func InitSmtp(password string) {
	smtpServer = &domain.SmtpConfig{
		Host:     "smtp.gmail.com",
		Port:     "587",
		Username: "zepppelin1.1@gmail.com",
		Auth:     smtp.PlainAuth("", "zepppelin1.1@gmail.com", password, "smtp.gmail.com"),
	}
}

func GetSmtpConfig() *domain.SmtpConfig {
	return smtpServer
}

func CheckSmtpAuth(cfg *domain.SmtpConfig) error {
	conn, err := smtpDial(cfg.Host + ":" + cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer conn.Close() // Note: Error from Close is ignored here

	tlsconfig := &tls.Config{
		InsecureSkipVerify: false, // Keep false for production safety
		ServerName:         cfg.Host,
	}

	if err = conn.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if err = conn.Auth(cfg.Auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

func ResetSmtpState() {
	smtpServer = nil
	smtpDial = func(addr string) (SmtpClient, error) {
		client, err := smtp.Dial(addr)
		if err != nil {
			var nilClient SmtpClient = nil
			return nilClient, err
		}
		return client, nil
	}
}
