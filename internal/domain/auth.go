package domain

import (
	"fmt"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/stretchr/testify/mock"
	"net/http"
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

// ClerkUserService defines the interface for user-related Clerk operations
type ClerkUserService interface {
	Create(params clerk.CreateUserParams) (*clerk.User, error)
	// Add other user methods if needed by AuthService in the future
}

// ClerkOrganizationService defines the interface for organization-related Clerk operations
type ClerkOrganizationService interface {
	CreateMembership(organizationID string, params clerk.CreateOrganizationMembershipParams) (*clerk.OrganizationMembership, error)
	// Add other org methods if needed
}

// ClerkTokenService defines the interface for token verification/decoding
type ClerkTokenService interface {
	VerifyToken(token string) (*clerk.SessionClaims, error)
	DecodeToken(token string) (*clerk.TokenClaims, error)
}

// ClerkClientInterface wraps the necessary Clerk client functionalities
type ClerkClientInterface interface {
	Users() ClerkUserService
	Organizations() ClerkOrganizationService
	VerifyToken(token string) (*clerk.SessionClaims, error)
	DecodeToken(token string) (*clerk.TokenClaims, error)
}

type ClerkInterface interface {
	VerifyToken(token string, opts ...clerk.VerifyTokenOption) (*clerk.SessionClaims, error)
	DecodeToken(token string) (*clerk.TokenClaims, error)
	CreateUser(params clerk.CreateUserParams) (*clerk.User, error)
	CreateOrganizationMembership(orgID string, params clerk.CreateOrganizationMembershipParams) (*clerk.OrganizationMembership, error)
	NewRequest(method, url string, body ...interface{}) (*http.Request, error)
	Do(req *http.Request, v interface{}) (*http.Response, error)
}

// ClerkWrapper implements ClerkInterface using the real Clerk client
type ClerkWrapper struct {
	Client clerk.Client
}

func (c *ClerkWrapper) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	return c.Client.NewRequest(method, url, body...)
}

func (c *ClerkWrapper) Do(req *http.Request, v interface{}) (*http.Response, error) {
	return c.Client.Do(req, v)
}

func (c *ClerkWrapper) VerifyToken(token string, opts ...clerk.VerifyTokenOption) (*clerk.SessionClaims, error) {
	return c.Client.VerifyToken(token, opts...)
}

func (c *ClerkWrapper) DecodeToken(token string) (*clerk.TokenClaims, error) {
	return c.Client.DecodeToken(token)
}

func (c *ClerkWrapper) CreateUser(params clerk.CreateUserParams) (*clerk.User, error) {
	return c.Client.Users().Create(params)
}

func (c *ClerkWrapper) CreateOrganizationMembership(orgID string, params clerk.CreateOrganizationMembershipParams) (*clerk.OrganizationMembership, error) {
	return c.Client.Organizations().CreateMembership(orgID, params)
}

// AuthService uses the ClerkInterface instead of directly using clerk.Client
type AuthService struct {
	Clerk ClerkInterface
}
