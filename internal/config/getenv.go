package config

import (
	"os"
)

func GetConnectionString() string {
	return os.Getenv("CONNECTION_STRING")
}

func GetMQConnectionString() string {

	return os.Getenv("MQ_CONN_STRING")
}

func GetSmtpPassword() string {
	return os.Getenv("SMTP_PASSWORD")
}
