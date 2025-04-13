package domain

import (
	"fmt"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/stretchr/testify/mock"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type AuthServiceI interface {
	VerifyToken(token string) (*clerk.SessionClaims, error)
	DecodeToken(token string) (*clerk.TokenClaims, error)
	CreateUser(input UserInput, organizationID string, role string) (*User, error)
}

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyToken(token string) (*clerk.SessionClaims, error) {
	args := m.Called(token)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(*clerk.SessionClaims), args.Error(1)
}

func (m *MockAuthService) DecodeToken(token string) (*clerk.TokenClaims, error) {
	args := m.Called(token)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(*clerk.TokenClaims), args.Error(1)
}

func (m *MockAuthService) CreateUser(input UserInput, organizationID string, role string) (*User, error) {
	args := m.Called(input, organizationID, role)
	res := args.Get(0) // This gets the first return value from the test's .Return() call
	if res == nil {
		return nil, args.Error(1)
	}
	user, ok := res.(*User)
	if !ok {
		panic(fmt.Sprintf("MockAuthService.CreateUser: unexpected type received in Return(). Expected *domain.User, got %T", res))
	}
	return user, args.Error(1)
}
