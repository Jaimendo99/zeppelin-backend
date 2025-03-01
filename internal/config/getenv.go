package config

import (
	"os"
)

func GetConnectionString() string {
	return os.Getenv("CONNECTION_STRING")
}

func GetClerkConfig() string {
	return os.Getenv("CLERK_API_KEY")
}
