package config

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"zeppelin/internal/domain"
)

var smtpServer *domain.SmtpConfig

func InitSmtp(password string) {
	smtpServer = &domain.SmtpConfig{
		Host:     "smtp.gmail.com",
		Port:     "587",
		Username: "zepppelin1.1@gmail.com",
		Auth:     smtp.PlainAuth("", "zepppelin1.1@gmail.com", password, "smtp.gmail.com"),
	}
}

func CheckSmtpAuth(smtpServer *domain.SmtpConfig) error {
	conn, err := smtp.Dial(smtpServer.Host + ":" + smtpServer.Port)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer conn.Close()
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpServer.Host,
	}
	if err = conn.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}
	if err = conn.Auth(smtpServer.Auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

func GetSmtpConfig() *domain.SmtpConfig {
	return smtpServer
}
