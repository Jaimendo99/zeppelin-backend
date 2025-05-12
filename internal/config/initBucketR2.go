package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

var (
	R2Client *s3.Client
	r2Once   sync.Once
	R2Error  error
)

func InitR2() error {
	r2Once.Do(func() {
		_ = godotenv.Load() // Carga .env (puedes eliminar esto si usas env vars directas)

		accessKey := os.Getenv("R2_ACCESS_KEY")
		secretKey := os.Getenv("R2_SECRET_KEY")
		accountID := os.Getenv("R2_ACCOUNT_ID")

		if accessKey == "" || secretKey == "" || accountID == "" {
			R2Error = fmt.Errorf("faltan variables de entorno para R2")
			return
		}

		endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
			config.WithRegion("auto"),
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: "auto",
					}, nil
				}),
			),
		)

		if err != nil {
			R2Error = fmt.Errorf("error al cargar configuración de R2: %w", err)
			return
		}

		R2Client = s3.NewFromConfig(cfg)
	})

	return R2Error
}
