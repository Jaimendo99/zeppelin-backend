package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	if os.IsNotExist(err) {
		log.Println("info: .env file not found, relying on environment variables")
	}
	return nil
}
