package middleware_test

import (
	"encoding/json"
	"errors"
	"github.com/go-jose/go-jose/v3/jwt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"zeppelin/internal/domain"
	"zeppelin/internal/middleware"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Assume createMockContext, mockNextHandler, MockAuthService exist from previous examples
// Assume domain.ErrAuthTokenMissing, ErrAuthTokenInvalid, ErrRoleExtractionFailed, ErrAuthorizationFailed exist

func TestWsAuthMiddleware(t *testing.T) {
	validToken := "valid-ws-token-123"

	// Helper function
	numericDatePtr := func(t int64) *jwt.NumericDate {
		nd := jwt.NumericDate(t)
		return &nd
	}

	// --- Test Cases ---
	tests := []struct {
		name                           string
		queryParamToken                string // Token value in query param
		requiredRoles                  []string
		setupMocks                     func(mockAuth *domain.MockAuthService, token string) // Pass token for expectation matching
		expectNextCalled               bool
		expectStatusCode               int
		expectBodyContains             string // Expected error message string from custom errors
		expectContextUserID            string
		expectContextUserRole          string
		expectContextUserSession       *clerk.TokenClaims   // Expected claims object
		expectContextUserSessionClaims *clerk.SessionClaims // Expected session claims object
	}{
		{
			name:            "Success - Role Required and Matched",
			queryParamToken: validToken,
			requiredRoles:   []string{"ws-user", "admin"},
			setupMocks: func(mockAuth *domain.MockAuthService, token string) {
				now := time.Now()
				claims := &clerk.TokenClaims{
					Claims: jwt.Claims{Subject: "ws_user_1", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					Extra:  map[string]interface{}{"role": "ws-user"},
				}
				sessionClaims := &clerk.SessionClaims{
					Claims:    jwt.Claims{Subject: "ws_user_1", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					SessionID: "sess_ws_1",
				}
				mockAuth.On("DecodeToken", token).Return(claims, nil).Once()
				mockAuth.On("VerifyToken", token).Return(sessionClaims, nil).Once()
			},
			expectNextCalled:               true,
			expectStatusCode:               http.StatusOK, // From mockNextHandler
			expectBodyContains:             "next handler called",
			expectContextUserID:            "ws_user_1",
			expectContextUserRole:          "ws-user",
			expectContextUserSession:       &clerk.TokenClaims{Claims: jwt.Claims{Subject: "ws_user_1"}, Extra: map[string]interface{}{"role": "ws-user"}}, // Simplified for comparison focus
			expectContextUserSessionClaims: &clerk.SessionClaims{Claims: jwt.Claims{Subject: "ws_user_1"}, SessionID: "sess_ws_1"},                         // Simplified
		},
		{
			name:            "Success - No Roles Required",
			queryParamToken: validToken,
			requiredRoles:   []string{}, // No roles
			setupMocks: func(mockAuth *domain.MockAuthService, token string) {
				now := time.Now()
				claims := &clerk.TokenClaims{
					Claims: jwt.Claims{Subject: "ws_user_2", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					Extra:  map[string]interface{}{"role": "guest"}, // Role doesn't matter here
				}
				sessionClaims := &clerk.SessionClaims{
					Claims:    jwt.Claims{Subject: "ws_user_2", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					SessionID: "sess_ws_2",
				}
				mockAuth.On("DecodeToken", token).Return(claims, nil).Once()
				mockAuth.On("VerifyToken", token).Return(sessionClaims, nil).Once()
			},
			expectNextCalled:               true,
			expectStatusCode:               http.StatusOK,
			expectBodyContains:             "next handler called",
			expectContextUserID:            "ws_user_2",
			expectContextUserRole:          "guest",
			expectContextUserSession:       &clerk.TokenClaims{Claims: jwt.Claims{Subject: "ws_user_2"}, Extra: map[string]interface{}{"role": "guest"}},
			expectContextUserSessionClaims: &clerk.SessionClaims{Claims: jwt.Claims{Subject: "ws_user_2"}, SessionID: "sess_ws_2"},
		},
		{
			name:               "Fail - Missing Token Query Param",
			queryParamToken:    "", // Empty token
			requiredRoles:      []string{"ws-user"},
			setupMocks:         func(mockAuth *domain.MockAuthService, token string) { /* No calls expected */ },
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized,          // Based on ReturnWriteResponse handling ErrAuthTokenMissing
			expectBodyContains: "authorization token is missing", // Match the specific error message
		},
		{
			name:            "Fail - DecodeToken Error",
			queryParamToken: "invalid-token",
			requiredRoles:   []string{"ws-user"},
			setupMocks: func(mockAuth *domain.MockAuthService, token string) {
				mockAuth.On("DecodeToken", token).Return(nil, errors.New("decode failed")).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized, // Based on ReturnWriteResponse handling ErrAuthTokenInvalid
			expectBodyContains: "authorization token is invalid",
		},
		{
			name:            "Fail - VerifyToken Error",
			queryParamToken: validToken,
			requiredRoles:   []string{"ws-user"},
			setupMocks: func(mockAuth *domain.MockAuthService, token string) {
				now := time.Now()
				// Decode succeeds
				claims := &clerk.TokenClaims{
					Claims: jwt.Claims{Subject: "ws_user_3", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					Extra:  map[string]interface{}{"role": "ws-user"},
				}
				mockAuth.On("DecodeToken", token).Return(claims, nil).Once()
				// Verify fails
				mockAuth.On("VerifyToken", token).Return(nil, errors.New("verify failed")).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized, // Based on ReturnWriteResponse handling ErrAuthTokenInvalid
			expectBodyContains: "authorization token is invalid",
		},
		{
			name:            "Fail - Role Extraction Error",
			queryParamToken: validToken,
			requiredRoles:   []string{"ws-user"},
			setupMocks: func(mockAuth *domain.MockAuthService, token string) {
				now := time.Now()
				// Decode succeeds with claims missing role
				claimsNoRole := &clerk.TokenClaims{
					Claims: jwt.Claims{Subject: "ws_user_4", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					Extra:  map[string]interface{}{}, // No role
				}
				// Verify succeeds
				sessionClaims := &clerk.SessionClaims{
					Claims:    jwt.Claims{Subject: "ws_user_4", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					SessionID: "sess_ws_4",
				}
				mockAuth.On("DecodeToken", token).Return(claimsNoRole, nil).Once()
				mockAuth.On("VerifyToken", token).Return(sessionClaims, nil).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusUnauthorized, // Assuming ReturnWriteResponse maps ErrRoleExtractionFailed to 500
			expectBodyContains: "token invalid: role extraction failed",
		},
		{
			name:            "Fail - Role Not Authorized",
			queryParamToken: validToken,
			requiredRoles:   []string{"admin"}, // Require 'admin'
			setupMocks: func(mockAuth *domain.MockAuthService, token string) {
				now := time.Now()
				// Decode/Verify succeed, but role is 'ws-user'
				claims := &clerk.TokenClaims{
					Claims: jwt.Claims{Subject: "ws_user_5", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					Extra:  map[string]interface{}{"role": "ws-user"}, // User has wrong role
				}
				sessionClaims := &clerk.SessionClaims{
					Claims:    jwt.Claims{Subject: "ws_user_5", IssuedAt: numericDatePtr(now.Unix() - 10), Expiry: numericDatePtr(now.Unix() + 3600)},
					SessionID: "sess_ws_5",
				}
				mockAuth.On("DecodeToken", token).Return(claims, nil).Once()
				mockAuth.On("VerifyToken", token).Return(sessionClaims, nil).Once()
			},
			expectNextCalled:   false,
			expectStatusCode:   http.StatusForbidden, // Based on ReturnWriteResponse handling ErrAuthorizationFailed
			expectBodyContains: "authorization failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// --- Setup ---
			mockAuthService := new(domain.MockAuthService)
			mockAuthService.Mock.Test(t)
			// Pass the specific token for this test case to setupMocks
			tc.setupMocks(mockAuthService, tc.queryParamToken)

			// Construct URL with query parameter
			url := "/ws"
			if tc.queryParamToken != "" {
				url = "/ws?token=" + tc.queryParamToken
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			//e := echo.New() // Fresh Echo instance for isolation
			// If ReturnWriteResponse relies on a custom error handler, set it here:
			// e.HTTPErrorHandler = yourCustomErrorHandler
			c, rec := createMockContext(req) // Use helper

			middlewareFunc := middleware.WsAuthMiddleware(mockAuthService, tc.requiredRoles...)

			nextCalled := false
			handler := middlewareFunc(mockNextHandler(t, &nextCalled)) // mockNextHandler sets status 200 on success

			// --- Execute ---
			err := handler(c)

			// --- Assert ---
			// Middleware should return nil after calling ReturnWriteResponse
			require.NoError(t, err, "Middleware handler returned an unexpected error")

			assert.Equal(t, tc.expectStatusCode, rec.Code, "Status code mismatch")
			if tc.expectBodyContains != "" {
				// Check body based on how ReturnWriteResponse formats it (assuming JSON with "message")
				var respBody map[string]interface{}
				jsonErr := json.Unmarshal(rec.Body.Bytes(), &respBody)
				if jsonErr == nil && respBody["message"] != nil {
					assert.Contains(t, strings.ToLower(respBody["message"].(string)), strings.ToLower(tc.expectBodyContains), "Body message mismatch")
				} else {
					assert.Contains(t, rec.Body.String(), strings.ToLower(tc.expectBodyContains), "Body content mismatch")
				}
			}

			assert.Equal(t, tc.expectNextCalled, nextCalled, "'next' handler call expectation mismatch")

			// Check context values only on success path
			if tc.expectNextCalled {
				assert.Equal(t, tc.expectContextUserID, c.Get("user_id"), "Context user_id mismatch")
				assert.Equal(t, tc.expectContextUserRole, c.Get("user_role"), "Context user_role mismatch")

				// Assert pointer types carefully
				actualSession := c.Get("user_session")
				require.NotNil(t, actualSession, "Context user_session is nil")
				require.IsType(t, &clerk.TokenClaims{}, actualSession, "Context user_session type mismatch")
				// Compare relevant fields, not timestamps which are hard to match exactly
				assert.Equal(t, tc.expectContextUserSession.Subject, actualSession.(*clerk.TokenClaims).Subject)
				assert.Equal(t, tc.expectContextUserSession.Extra, actualSession.(*clerk.TokenClaims).Extra)

				actualSessionClaims := c.Get("user_session_claims")
				require.NotNil(t, actualSessionClaims, "Context user_session_claims is nil")
				require.IsType(t, &clerk.SessionClaims{}, actualSessionClaims, "Context user_session_claims type mismatch")
				assert.Equal(t, tc.expectContextUserSessionClaims.Subject, actualSessionClaims.(*clerk.SessionClaims).Subject)
				assert.Equal(t, tc.expectContextUserSessionClaims.SessionID, actualSessionClaims.(*clerk.SessionClaims).SessionID)

			} else {
				// Ensure context values were NOT set on failure paths
				assert.Nil(t, c.Get("user_id"), "Context user_id should be nil on failure")
				assert.Nil(t, c.Get("user_role"), "Context user_role should be nil on failure")
				assert.Nil(t, c.Get("user_session"), "Context user_session should be nil on failure")
				assert.Nil(t, c.Get("user_session_claims"), "Context user_session_claims should be nil on failure")
			}

			mockAuthService.AssertExpectations(t) // Verify mocks
		})
	}
}
