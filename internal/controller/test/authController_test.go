package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"zeppelin/internal/controller"
)

type MockClerk struct {
	mock.Mock
}

func (m *MockClerk) VerifyToken(token string, opts ...clerk.VerifyTokenOption) (*clerk.SessionClaims, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) DecodeToken(token string) (*clerk.TokenClaims, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) CreateUser(params clerk.CreateUserParams) (*clerk.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) CreateOrganizationMembership(orgID string, params clerk.CreateOrganizationMembershipParams) (*clerk.OrganizationMembership, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	// We pass body as separate args to Called for easier matching if needed
	allArgs := []interface{}{method, url}
	allArgs = append(allArgs, body...)
	args := m.Called(allArgs...)

	// Return value handling
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Request), args.Error(1)
}

func (m *MockClerk) Do(req *http.Request, v interface{}) (*http.Response, error) {
	args := m.Called(req, v) // Pass 'v' so mock.Run can access it

	// Return value handling
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func NewMockClerk(t *testing.T) *MockClerk {
	m := new(MockClerk)
	m.Mock.Test(t)
	return m
}

func createMockContext(e *echo.Echo, req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestAuthController_GetTokenFromSession(t *testing.T) {
	testSessionID := "sess_12345"
	testTemplate := "my_jwt_template"
	expectedURLPath := fmt.Sprintf("sessions/%s/tokens/%s", testSessionID, testTemplate)
	expectedJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	// --- Test Cases ---
	tests := []struct {
		name               string
		sessionIDParam     string
		templateParam      string
		setupMocks         func(mockClerk *MockClerk)
		expectedStatus     int
		expectedBody       interface{} // Can be clerk.SessionToken or error response struct
		expectBodyContains string      // For asserting substrings in error messages
	}{
		{
			name:           "Success",
			sessionIDParam: testSessionID,
			templateParam:  testTemplate,
			setupMocks: func(mockClerk *MockClerk) {
				// Mock NewRequest
				mockReq, _ := http.NewRequest("POST", "http://clerk.example"+expectedURLPath, nil) // Dummy URL for request object
				mockClerk.On("NewRequest", "POST", expectedURLPath, mock.Anything).Return(mockReq, nil).Once()

				// Mock Do - Use Run to populate the response struct
				mockResp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{}`))} // Dummy response
				mockClerk.On("Do", mockReq, mock.AnythingOfType("*clerk.SessionToken")).
					Run(func(args mock.Arguments) {
						// Get the pointer passed to Do (the second argument, index 1)
						tokenPtr := args.Get(1).(*clerk.SessionToken)
						// Populate the struct it points to
						tokenPtr.JWT = expectedJWT
					}).
					Return(mockResp, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   clerk.SessionToken{JWT: expectedJWT},
		},
		{
			name:               "Fail - Missing Session ID",
			sessionIDParam:     "", // Missing
			templateParam:      testTemplate,
			setupMocks:         func(mockClerk *MockClerk) { /* No calls expected */ },
			expectedStatus:     http.StatusBadRequest,             // Assuming ReturnWriteResponse maps ErrRequiredParamsMissing to 400
			expectBodyContains: "Required parameters are missing", // Or match domain.ErrRequiredParamsMissing.Error() if handled that way
		},
		{
			name:               "Fail - Missing Template",
			sessionIDParam:     testSessionID,
			templateParam:      "", // Missing
			setupMocks:         func(mockClerk *MockClerk) { /* No calls expected */ },
			expectedStatus:     http.StatusBadRequest,
			expectBodyContains: "Required parameters are missing",
		},
		{
			name:           "Fail - NewRequest Error",
			sessionIDParam: testSessionID,
			templateParam:  testTemplate,
			setupMocks: func(mockClerk *MockClerk) {
				mockClerk.On("NewRequest", "POST", expectedURLPath, mock.Anything).
					Return(nil, errors.New("failed to create request")).Once()
				// Do should not be called
			},
			expectedStatus:     http.StatusInternalServerError, // Assuming ReturnWriteResponse maps generic errors to 500
			expectBodyContains: "error creating auth request",
		},
		{
			name:           "Fail - Do Error",
			sessionIDParam: testSessionID,
			templateParam:  testTemplate,
			setupMocks: func(mockClerk *MockClerk) {
				mockReq, _ := http.NewRequest("POST", "http://clerk.example"+expectedURLPath, nil)
				mockClerk.On("NewRequest", "POST", expectedURLPath, mock.Anything).Return(mockReq, nil).Once()

				// Mock Do to return an error
				doErr := errors.New("clerk API error")
				mockClerk.On("Do", mockReq, mock.AnythingOfType("*clerk.SessionToken")).
					Return(nil, doErr).Once() // Return nil response, non-nil error
			},
			expectedStatus:     http.StatusInternalServerError,
			expectBodyContains: "error processing auth response", // Check prefix
			// Optionally check if the original error message is included:
			// expectBodyContains: "clerk API error",
		},
		{
			name:           "Fail - Empty JWT in Response",
			sessionIDParam: testSessionID,
			templateParam:  testTemplate,
			setupMocks: func(mockClerk *MockClerk) {
				mockReq, _ := http.NewRequest("POST", "http://clerk.example"+expectedURLPath, nil)
				mockClerk.On("NewRequest", "POST", expectedURLPath, mock.Anything).Return(mockReq, nil).Once()

				// Mock Do to succeed but not populate JWT
				mockResp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{}`))}
				mockClerk.On("Do", mockReq, mock.AnythingOfType("*clerk.SessionToken")).
					Run(func(args mock.Arguments) {
						tokenPtr := args.Get(1).(*clerk.SessionToken)
						tokenPtr.JWT = "" // Explicitly empty
					}).
					Return(mockResp, nil).Once()
			},
			expectedStatus:     http.StatusInternalServerError,
			expectBodyContains: "empty JWT in token response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// --- Setup ---
			mockClerk := NewMockClerk(t)
			authController := controller.AuthController{Clerk: mockClerk}
			tc.setupMocks(mockClerk)

			// Build request URL with query params
			q := make(url.Values)
			q.Set("sessionId", tc.sessionIDParam)
			q.Set("template", tc.templateParam)
			targetURL := "/some/path?" + q.Encode() // Path doesn't matter much here

			req := httptest.NewRequest(http.MethodGet, targetURL, nil) // Method doesn't matter much for query params

			e := echo.New()
			// Set custom error handler if ReturnWriteResponse relies on it
			// e.HTTPErrorHandler = yourCustomErrorHandler
			c, rec := createMockContext(e, req)

			handler := authController.GetTokenFromSession()

			// --- Execute ---
			err := handler(c)

			// --- Assert ---
			// Handler should return nil if ReturnWriteResponse handles the response
			require.NoError(t, err, "Handler returned an unexpected error")
			assert.Equal(t, tc.expectedStatus, rec.Code, "Status code mismatch")

			// Assert Body
			if tc.expectedStatus == http.StatusOK {
				var actualResp clerk.SessionToken
				err := json.Unmarshal(rec.Body.Bytes(), &actualResp)
				require.NoError(t, err, "Failed to unmarshal success response body")
				assert.Equal(t, tc.expectedBody, actualResp, "Success response body mismatch")
			} else {
				// Check error response body contains expected message
				// Assuming ReturnWriteResponse sends {"message": "..."}
				var errResp map[string]interface{}
				jsonErr := json.Unmarshal(rec.Body.Bytes(), &errResp)
				if jsonErr == nil && errResp["message"] != nil {
					assert.Contains(t, errResp["message"], tc.expectBodyContains, "Error response message mismatch")
				} else {
					// Fallback if not standard JSON error format
					assert.Contains(t, rec.Body.String(), tc.expectBodyContains, "Error response body content mismatch")
				}
			}

			mockClerk.AssertExpectations(t) // Verify mocks
		})
	}
}
