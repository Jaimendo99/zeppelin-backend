package services

import (
	"errors"
	"zeppelin/internal/config"

	"github.com/clerkinc/clerk-sdk-go/clerk"
)

type AuthService struct {
	client clerk.Client
}

func NewAuthService() (*AuthService, error) {
	apiKey := config.GetClerkConfig()
	if apiKey == "" {
		return nil, errors.New("CLERK_API_KEY no está configurada")
	}

	client, err := clerk.NewClient(apiKey)
	if err != nil {
		return nil, err
	}

	return &AuthService{client: client}, nil
}

func (s *AuthService) VerifyToken(token string) (*clerk.SessionClaims, error) {
	claims, err := s.client.VerifyToken(token)
	if err != nil || claims == nil {
		return nil, errors.New("token inválido o sesión no encontrada")
	}

	return claims, nil
}
