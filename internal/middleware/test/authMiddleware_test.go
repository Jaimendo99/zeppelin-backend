package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time" // Needed for clerk claims
	"zeppelin/internal/domain"
	"zeppelin/internal/middleware"

	"github.com/stretchr/testify/mock"

	"github.com/go-jose/go-jose/v3/jwt"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a mock Echo context for testing middleware
func createMockContext(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	e := echo.New() // Create a new Echo instance
	// You might need to register your custom error handler if ReturnWriteResponse relies on it
	// e.HTTPErrorHandler = customErrorHandler
	c := e.NewContext(req, rec)
	return c, rec
}

// Mock handler to check if 'next' was called
func mockNextHandler(t *testing.T, called *bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		*called = true
		return c.String(http.StatusOK, "next handler called") // Simple success response
	}
}
func TestValidateTokenAndRole(t *testing.T) {
	// Keep only truly common variables outside, like the token string
	validToken := "valid-token"

	// Helper function can stay outside if used in multiple subtests,
	// or move inside if only used once or twice. Let's keep it out for now.
	numericDatePtr := func(t int64) *jwt.NumericDate {
		nd := jwt.NumericDate(t)
		return &nd
	}

	t.Run("Success - Role Required and Matched", func(t *testing.T) {
		// Define needed variables and claims INSIDE t.Run
		now := time.Now()
		iat := now.Unix() - 1000
		exp := now.Unix() + 3600
		subject := "user_123"
		validClaims := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			Extra: map[string]interface{}{"role": "admin"},
		}
		validSessionClaims := &clerk.SessionClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			SessionID: "sess_abc",
		}

		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)

		mockAuthService.On("DecodeToken", validToken).Return(validClaims, nil).Once()
		mockAuthService.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService, "admin", "editor")

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, validClaims, claims)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Success - No Roles Required", func(t *testing.T) {
		now := time.Now()
		iat := now.Unix() - 1000
		exp := now.Unix() + 3600
		subject := "user_123"

		validClaims := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			Extra: map[string]interface{}{"role": "admin"},
		}
		validSessionClaims := &clerk.SessionClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			SessionID: "sess_abc",
		}

		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)

		mockAuthService.On("DecodeToken", validToken).Return(validClaims, nil).Once()
		mockAuthService.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, validClaims, claims)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Error - Empty Token", func(t *testing.T) {
		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)

		claims, err := middleware.ValidateTokenAndRole("", mockAuthService, "admin")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "token requerido")
		mockAuthService.AssertNotCalled(t, "DecodeToken", mock.Anything)
		mockAuthService.AssertNotCalled(t, "VerifyToken", mock.Anything)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Error - DecodeToken Fails", func(t *testing.T) {
		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)

		decodeErr := errors.New("decode failed")
		mockAuthService.On("DecodeToken", validToken).Return(nil, decodeErr).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService, "admin")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "token inválido o sesión no encontrada")
		mockAuthService.AssertNotCalled(t, "VerifyToken", mock.Anything)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Error - VerifyToken Fails", func(t *testing.T) {
		now := time.Now()
		iat := now.Unix() - 1000
		exp := now.Unix() + 3600
		subject := "user_123"

		validClaims := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			Extra: map[string]interface{}{"role": "org:admin"},
		}

		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)
		verifyErr := errors.New("verify failed")
		// Need validClaims to be returned by DecodeToken before VerifyToken is called
		mockAuthService.On("DecodeToken", validToken).Return(validClaims, nil).Once()
		mockAuthService.On("VerifyToken", validToken).Return(nil, verifyErr).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService, "admin")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "token inválido o sesión no encontrada")
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Error - Role Extraction Fails (No Extra)", func(t *testing.T) {
		// Define needed variables and claims INSIDE t.Run
		now := time.Now()
		iat := now.Unix() - 1000
		exp := now.Unix() + 3600
		subject := "user_123"
		claimsNoExtra := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			Extra: nil,
		}
		// Need validSessionClaims for the VerifyToken mock
		validSessionClaims := &clerk.SessionClaims{ /* ... define as needed ... */ }

		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)
		mockAuthService.On("DecodeToken", validToken).Return(claimsNoExtra, nil).Once()
		mockAuthService.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService, "admin")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "no se pudo extraer el rol del usuario")
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Error - Role Extraction Fails (No Role Field)", func(t *testing.T) {
		// Define needed variables and claims INSIDE t.Run
		now := time.Now()
		iat := now.Unix() - 1000
		exp := now.Unix() + 3600
		subject := "user_123"
		claimsNoRole := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			Extra: map[string]interface{}{"other_field": "value"},
		}
		validSessionClaims := &clerk.SessionClaims{ /* ... define as needed ... */ }

		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)
		mockAuthService.On("DecodeToken", validToken).Return(claimsNoRole, nil).Once()
		mockAuthService.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService, "admin")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "no se pudo extraer el rol del usuario")
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Error - Role Not Authorized", func(t *testing.T) {
		now := time.Now()
		iat := now.Unix() - 1000
		exp := now.Unix() + 3600
		subject := "user_123"

		validClaims := &clerk.TokenClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			Extra: map[string]interface{}{"role": "org:admin"},
		}
		validSessionClaims := &clerk.SessionClaims{
			Claims: jwt.Claims{
				Subject:  subject,
				IssuedAt: numericDatePtr(iat),
				Expiry:   numericDatePtr(exp),
			},
			SessionID: "sess_abc",
		}

		mockAuthService := new(domain.MockAuthService)
		mockAuthService.Mock.Test(t)
		mockAuthService.On("DecodeToken", validToken).Return(validClaims, nil).Once()
		mockAuthService.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()

		claims, err := middleware.ValidateTokenAndRole(validToken, mockAuthService, "superadmin", "manager")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "acceso denegado: rol no autorizado")
		mockAuthService.AssertExpectations(t)
	})
}

// --- Test RoleMiddleware ---
func TestRoleMiddleware(t *testing.T) {
	validToken := "valid-token-for-middleware"

	// Helper function
	numericDatePtr := func(t int64) *jwt.NumericDate {
		nd := jwt.NumericDate(t)
		return &nd
	}

	now := time.Now()
	iat := now.Unix() - 1000
	exp := now.Unix() + 3600
	subject := "user_456"

	// Corrected Initialization
	validClaims := &clerk.TokenClaims{
		Claims: jwt.Claims{
			Subject:  subject,
			IssuedAt: numericDatePtr(iat),
			Expiry:   numericDatePtr(exp),
		},
		Extra: map[string]interface{}{"role": "editor"},
	}
	validSessionClaims := &clerk.SessionClaims{
		Claims: jwt.Claims{
			Subject:  subject,
			IssuedAt: numericDatePtr(iat),
			Expiry:   numericDatePtr(exp),
		},
		SessionID: "sess_def",
	} // --- Test Cases ---
	tests := []struct {
		name               string
		requiredRoles      []string
		setupRequest       func() *http.Request
		setupMocks         func(mockAuth *domain.MockAuthService)
		expectNextCalled   bool
		expectStatusCode   int
		expectBodyContains string // Substring to check in the response body
		expectUserID       string // Expected user ID set in context
		expectUserRole     string // Expected user role set in context
	}{
		{
			name:          "Success - Role Required and Matched",
			requiredRoles: []string{"editor", "admin"},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				mockAuth.On("DecodeToken", validToken).Return(validClaims, nil).Once()
				mockAuth.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()
			},
			expectNextCalled:   true,
			expectStatusCode:   http.StatusOK, // From mockNextHandler
			expectBodyContains: "next handler called",
			expectUserID:       "user_456",
			expectUserRole:     "editor",
		},
		{
			name:          "Success - No Roles Required",
			requiredRoles: []string{}, // Empty slice
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				mockAuth.On("DecodeToken", validToken).Return(validClaims, nil).Once()
				mockAuth.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()
			},
			expectNextCalled:   true,
			expectStatusCode:   http.StatusOK,
			expectBodyContains: "next handler called",
			expectUserID:       "user_456",
			expectUserRole:     "editor",
		},
		{
			name:          "Fail - No Authorization Header",
			requiredRoles: []string{"editor"},
			setupRequest: func() *http.Request {
				// No Auth header set
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				// No calls expected
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized, // Assuming ReturnWriteResponse maps this way
			expectBodyContains: "token is missing",
		},
		{
			name:          "Fail - Invalid Header Format (No Bearer)",
			requiredRoles: []string{"editor"},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", validToken) // Missing "Bearer " prefix
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				// DecodeToken will be called with the wrong token format
				mockAuth.On("DecodeToken", validToken).Return(nil, errors.New("decode failed")).Once()
				// VerifyToken might or might not be called depending on DecodeToken failure handling
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized,
			expectBodyContains: "token is invalid", // Or session not found
		},
		{
			name:          "Fail - DecodeToken Error",
			requiredRoles: []string{"editor"},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				mockAuth.On("DecodeToken", validToken).Return(nil, errors.New("decode failed")).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized,
			expectBodyContains: "token is invalid",
		},
		{
			name:          "Fail - VerifyToken Error",
			requiredRoles: []string{"editor"},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				mockAuth.On("DecodeToken", validToken).Return(validClaims, nil).Once()
				mockAuth.On("VerifyToken", validToken).Return(nil, errors.New("verify failed")).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized,
			expectBodyContains: "token is invalid",
		},
		{
			name:          "Fail - Role Extraction Error",
			requiredRoles: []string{"editor"},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				claimsNoRole := &clerk.TokenClaims{ // Claims without role info
					Claims: jwt.Claims{Subject: "user_456"},
					Extra:  map[string]interface{}{},
				}
				mockAuth.On("DecodeToken", validToken).Return(claimsNoRole, nil).Once()
				mockAuth.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized,
			expectBodyContains: "Role extraction failed",
		},
		{
			name:          "Fail - Role Not Authorized",
			requiredRoles: []string{"admin", "superadmin"}, // User has 'editor' role
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			setupMocks: func(mockAuth *domain.MockAuthService) {
				mockAuth.On("DecodeToken", validToken).Return(validClaims, nil).Once() // validClaims has 'editor' role
				mockAuth.On("VerifyToken", validToken).Return(validSessionClaims, nil).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusForbidden, // Or StatusUnauthorized depending on ReturnWriteResponse
			expectBodyContains: "Authorization failed",
			expectUserID:       "user_456", // Context values might still be set before the final check
			expectUserRole:     "editor",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// --- Setup ---
			mockAuthService := new(domain.MockAuthService)
			mockAuthService.Mock.Test(t) // Link mock to test
			tc.setupMocks(mockAuthService)

			req := tc.setupRequest()
			c, rec := createMockContext(req)

			// Need to pass the actual AuthService implementation type if the middleware expects it,
			// but the mock should satisfy the interface. If AuthService is a struct with methods,
			// you might need to adjust how you pass the mock or wrap it.
			// Assuming AuthService is an interface or the mock can be used directly:
			middlewareFunc := middleware.RoleMiddleware(mockAuthService, tc.requiredRoles...)

			nextCalled := false
			handler := middlewareFunc(mockNextHandler(t, &nextCalled))

			// --- Execute ---
			_ = handler(c)

			// --- Assert ---
			// Check if handler returned an error (it might, even if response is written)
			// Often, Echo handlers write the response and return nil unless there's an unrecoverable error.
			// We primarily check the recorder's status and body.
			// require.NoError(t, err) // This might fail if ReturnWriteResponse returns the error

			assert.Equal(t, tc.expectStatusCode, rec.Code, "Status code mismatch")
			if tc.expectBodyContains != "" {
				// Check if the body contains the expected error message substring
				bodyBytes := rec.Body.Bytes()
				// Try unmarshalling if ReturnWriteResponse sends JSON
				var respBody map[string]interface{}
				jsonErr := json.Unmarshal(bodyBytes, &respBody)
				if jsonErr == nil && respBody["message"] != nil {
					assert.Contains(t, respBody["message"], tc.expectBodyContains, "Body message mismatch")
				} else {
					// Fallback to string contains if not JSON or message field missing
					assert.Contains(t, rec.Body.String(), tc.expectBodyContains, "Body content mismatch")
				}
			}

			assert.Equal(t, tc.expectNextCalled, nextCalled, "'next' handler call expectation mismatch")

			// Check context values only if they are expected to be set
			if tc.expectUserID != "" {
				assert.Equal(t, tc.expectUserID, c.Get("user_id"), "Context user_id mismatch")
			}
			if tc.expectUserRole != "" {
				assert.Equal(t, tc.expectUserRole, c.Get("user_role"), "Context user_role mismatch")
			}

			mockAuthService.AssertExpectations(t) // Verify mocks
		})
	}
}
