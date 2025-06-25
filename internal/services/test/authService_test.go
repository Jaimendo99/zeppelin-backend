// auth_service_test.go
package services_test

import (
	"errors"
	"github.com/go-jose/go-jose/v3/jwt"
	"net/http"
	"testing"
	"zeppelin/internal/domain"
	"zeppelin/internal/services"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClerk implements the ClerkInterface for testing
type MockClerk struct {
	mock.Mock
}

func (m *MockClerk) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	panic("implement me")
}

func (m *MockClerk) Do(req *http.Request, v interface{}) (*http.Response, error) {
	panic("implement me")
}

func (m *MockClerk) VerifyToken(token string, opts ...clerk.VerifyTokenOption) (*clerk.SessionClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clerk.SessionClaims), args.Error(1)
}

func (m *MockClerk) DecodeToken(token string) (*clerk.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clerk.TokenClaims), args.Error(1)
}

func (m *MockClerk) CreateUser(params clerk.CreateUserParams) (*clerk.User, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clerk.User), args.Error(1)
}

func (m *MockClerk) CreateOrganizationMembership(orgID string, params clerk.CreateOrganizationMembershipParams) (*clerk.OrganizationMembership, error) {
	args := m.Called(orgID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clerk.OrganizationMembership), args.Error(1)
}

func TestVerifyToken(t *testing.T) {
	mockClerk := new(MockClerk)
	authService := &services.AuthService{Clerk: mockClerk}

	t.Run("Valid token", func(t *testing.T) {
		expectedClaims := &clerk.SessionClaims{
			Claims: jwt.Claims{
				Subject: "user_123",
			},
			SessionID: "session_123",
		}

		mockClerk.On("VerifyToken", "valid-token").Return(expectedClaims, nil).Once()

		claims, err := authService.VerifyToken("valid-token")

		assert.NoError(t, err)
		assert.Equal(t, expectedClaims, claims)
		mockClerk.AssertExpectations(t)
	})

	t.Run("Invalid token", func(t *testing.T) {
		mockClerk.On("VerifyToken", "invalid-token").Return(nil, errors.New("invalid token")).Once()

		claims, err := authService.VerifyToken("invalid-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, "token inválido o sesión no encontrada", err.Error())
		mockClerk.AssertExpectations(t)
	})
}

func TestDecodeToken(t *testing.T) {
	mockClerk := new(MockClerk)
	authService := &services.AuthService{Clerk: mockClerk}

	t.Run("Valid token", func(t *testing.T) {
		expectedClaims := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject: "user_123",
			},
			Extra: map[string]interface{}{
				"email": "test@example.com",
			},
		}

		mockClerk.On("DecodeToken", "valid-token").Return(expectedClaims, nil).Once()

		claims, err := authService.DecodeToken("valid-token")

		assert.NoError(t, err)
		assert.Equal(t, expectedClaims, claims)
		mockClerk.AssertExpectations(t)
	})

	t.Run("Invalid token", func(t *testing.T) {
		mockClerk.On("DecodeToken", "invalid-token").Return(nil, errors.New("invalid token")).Once()

		claims, err := authService.DecodeToken("invalid-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, "token inválido o sesión no encontrada", err.Error())
		mockClerk.AssertExpectations(t)
	})
}

func TestCreateUser(t *testing.T) {
	mockClerk := new(MockClerk)
	authService := &services.AuthService{Clerk: mockClerk}

	t.Run("Create user without organization", func(t *testing.T) {
		input := domain.UserInput{
			Name:     "John",
			Lastname: "Doe",
			Email:    "john.doe@example.com",
		}

		expectedUser := &clerk.User{
			ID: "user_123",
		}

		// Match the parameters that will be passed to CreateUser
		mockClerk.On("CreateUser", mock.MatchedBy(func(params clerk.CreateUserParams) bool {
			return params.EmailAddresses[0] == input.Email &&
				*params.FirstName == input.Name &&
				*params.LastName == input.Lastname &&
				*params.SkipPasswordRequirement == true
		})).Return(expectedUser, nil).Once()

		user, err := authService.CreateUser(input, "", "user")

		assert.NoError(t, err)
		assert.Equal(t, "user_123", user.UserID)
		mockClerk.AssertExpectations(t)
	})

	t.Run("Create user with organization", func(t *testing.T) {
		input := domain.UserInput{
			Name:     "Jane",
			Lastname: "Smith",
			Email:    "jane.smith@example.com",
		}

		expectedUser := &clerk.User{
			ID: "user_456",
		}

		expectedMembership := &clerk.OrganizationMembership{
			ID: "membership_123",
		}

		mockClerk.On("CreateUser", mock.MatchedBy(func(params clerk.CreateUserParams) bool {
			return params.EmailAddresses[0] == input.Email &&
				*params.FirstName == input.Name &&
				*params.LastName == input.Lastname
		})).Return(expectedUser, nil).Once()

		mockClerk.On("CreateOrganizationMembership", "org_123", mock.MatchedBy(func(params clerk.CreateOrganizationMembershipParams) bool {
			return params.UserID == "user_456" && params.Role == "admin"
		})).Return(expectedMembership, nil).Once()

		user, err := authService.CreateUser(input, "org_123", "admin")

		assert.NoError(t, err)
		assert.Equal(t, "user_456", user.UserID)
		mockClerk.AssertExpectations(t)
	})

	t.Run("Error creating user", func(t *testing.T) {
		input := domain.UserInput{
			Name:     "Error",
			Lastname: "User",
			Email:    "error@example.com",
		}

		mockClerk.On("CreateUser", mock.MatchedBy(func(params clerk.CreateUserParams) bool {
			return params.EmailAddresses[0] == input.Email
		})).Return(nil, errors.New("failed to create user")).Once()

		user, err := authService.CreateUser(input, "", "user")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "failed to create user", err.Error())
		mockClerk.AssertExpectations(t)
	})

	t.Run("Error creating organization membership", func(t *testing.T) {
		input := domain.UserInput{
			Name:     "Org",
			Lastname: "Error",
			Email:    "org.error@example.com",
		}

		expectedUser := &clerk.User{
			ID: "user_789",
		}

		mockClerk.On("CreateUser", mock.MatchedBy(func(params clerk.CreateUserParams) bool {
			return params.EmailAddresses[0] == input.Email
		})).Return(expectedUser, nil).Once()

		mockClerk.On("CreateOrganizationMembership", "org_error", mock.MatchedBy(func(params clerk.CreateOrganizationMembershipParams) bool {
			return params.UserID == "user_789"
		})).Return(nil, errors.New("failed to create membership")).Once()

		user, err := authService.CreateUser(input, "org_error", "user")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "failed to create membership", err.Error())
		mockClerk.AssertExpectations(t)
	})
}
