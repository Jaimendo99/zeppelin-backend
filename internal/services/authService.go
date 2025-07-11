package services

import (
	"encoding/json"
	"errors"
	"zeppelin/internal/config"
	"zeppelin/internal/domain"

	"github.com/clerkinc/clerk-sdk-go/clerk"
)

// AuthService uses the ClerkInterface instead of directly using clerk.Client
type AuthService struct {
	Clerk domain.ClerkInterface
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

	clerkWrapper := &domain.ClerkWrapper{Client: client}
	return &AuthService{Clerk: clerkWrapper}, nil
}

func (s *AuthService) VerifyToken(token string) (*clerk.SessionClaims, error) {
	claims, err := s.Clerk.VerifyToken(token)
	if err != nil || claims == nil {
		return nil, errors.New("token inválido o sesión no encontrada")
	}
	return claims, nil
}

func (s *AuthService) DecodeToken(token string) (*clerk.TokenClaims, error) {
	claims, err := s.Clerk.DecodeToken(token)
	if err != nil || claims == nil {
		return nil, errors.New("token inválido o sesión no encontrada")
	}
	return claims, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func (s *AuthService) CreateUser(input domain.UserInput, organizationID string, role string) (*domain.User, error) {
	if s.Clerk == nil {
		return nil, errors.New("error interno: Clerk Client no está inicializado")
	}

	publicMetadata := map[string]string{"role": role}
	publicMetadataJSON, _ := json.Marshal(publicMetadata)

	newUser, err := s.Clerk.CreateUser(clerk.CreateUserParams{
		EmailAddresses:          []string{input.Email},
		FirstName:               &input.Name,
		LastName:                &input.Lastname,
		SkipPasswordRequirement: boolPtr(true),
		PublicMetadata:          (*json.RawMessage)(&publicMetadataJSON),
	})
	if err != nil {
		return nil, err
	}

	if organizationID != "" {
		_, err := s.Clerk.CreateOrganizationMembership(organizationID, clerk.CreateOrganizationMembershipParams{
			UserID: newUser.ID,
			Role:   role,
		})
		if err != nil {
			return nil, err
		}
	}

	return &domain.User{
		UserID: newUser.ID,
	}, nil
}
