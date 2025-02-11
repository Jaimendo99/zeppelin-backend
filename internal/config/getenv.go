package config

import (
	"os"
)

func GetConnectionString() string {
	return os.Getenv("CONNECTION_STRING")
}
