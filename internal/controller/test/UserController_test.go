package controller

import (
	"encoding/json"
	"errors"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

type MockParentalConsentRepo struct {
	CreateConsentFunc       func(consent domain.ParentalConsent) error
	UpdateConsentStatusFunc func(token string, status string, ip string, userAgent string) error
	GetConsentByTokenFunc   func(token string) (*domain.ParentalConsent, error)
	GetConsentByUserIDFunc  func(userID string) (*domain.ParentalConsent, error)
}

func (m *MockParentalConsentRepo) CreateConsent(consent domain.ParentalConsent) error {
	if m.CreateConsentFunc != nil {
		return m.CreateConsentFunc(consent)
	}
	return nil
}

func (m *MockParentalConsentRepo) UpdateConsentStatus(token string, status string, ip string, userAgent string) error {
	if m.UpdateConsentStatusFunc != nil {
		return m.UpdateConsentStatusFunc(token, status, ip, userAgent)
	}
	return nil
}

func (m *MockParentalConsentRepo) GetConsentByToken(token string) (*domain.ParentalConsent, error) {
	if m.GetConsentByTokenFunc != nil {
		return m.GetConsentByTokenFunc(token)
	}
	return &domain.ParentalConsent{Status: "ACCEPTED"}, nil
}

func (m *MockParentalConsentRepo) GetConsentByUserID(userID string) (*domain.ParentalConsent, error) {
	if m.GetConsentByUserIDFunc != nil {
		return m.GetConsentByUserIDFunc(userID)
	}
	return &domain.ParentalConsent{Status: "ACCEPTED"}, nil
}

func setupTest(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = &controller.CustomValidator{Validator: validator.New()}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestUserController_RegisterUser(t *testing.T) {
	mockAuthService := new(domain.MockAuthService)

	mockUserRepo := new(domain.MockUserRepo)
	mockRepRepo := new(domain.MockRepresentativeRepo)
	mockConsentRepo := &MockParentalConsentRepo{
		CreateConsentFunc: func(consent domain.ParentalConsent) error {
			return nil
		},
	}

	mockSendEmail := func(to string, token string) error {
		return nil
	}

	userController := controller.UserController{
		AuthService:   mockAuthService,
		UserRepo:      mockUserRepo,
		RepRepo:       mockRepRepo,
		ConsentRepo:   mockConsentRepo,
		SendEmailFunc: mockSendEmail,
	}

	// --- Test Case 1: Success - Student ---
	t.Run("Success_Student", func(t *testing.T) {
		userInput := domain.UserInput{
			Name:     "Test",
			Lastname: "User",
			Email:    "test.student@example.com",
			Representative: domain.RepresentativeInput{
				Name:        "Parent",
				Lastname:    "One",
				Email:       "parent@example.com",
				PhoneNumber: "123456789",
			},
		}

		userInputJSON, _ := json.Marshal(userInput)
		req := httptest.NewRequest(http.MethodPost, "/register/student", strings.NewReader(string(userInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		role := "org:student"
		expectedOrgID := "org_2tjxBeJV0WLJUFU6Q3AwjzMyXTs"
		mockClerkUserID := "user_clerk_123"
		// Adjust the mock return type based on your actual domain.User struct
		mockAuthService.On("CreateUser", userInput, expectedOrgID, role).Return(&domain.User{UserID: mockClerkUserID}, nil).Once()

		expectedUserDb := domain.UserDb{
			UserID:   mockClerkUserID,
			Name:     "Test",
			Lastname: "User",
			Email:    "test.student@example.com",
			TypeID:   3, // Student TypeID
		}

		mockUserRepo.On("CreateUser", expectedUserDb).Return(nil).Once()
		mockRepRepo.On("CreateRepresentative", mock.AnythingOfType("domain.RepresentativeDb")).Return(123, nil)

		handler := userController.RegisterUser(role)
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		// Check response body based on ReturnWriteResponse's success case
		expectedResp := `{"Body":{"message":"Usuario registrado con éxito"}}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockAuthService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 2: Success - Teacher ---
	t.Run("Success_Teacher", func(t *testing.T) {
		// Similar setup as student, but with role "org:teacher" and TypeID 2
		userInput := domain.UserInput{
			Name:     "Test",
			Lastname: "Teacher",
			Email:    "test.teacher@example.com",
			Representative: domain.RepresentativeInput{
				Name:        "Dummy",
				Lastname:    "Dummy",
				Email:       "dummy@example.com",
				PhoneNumber: "0000000000",
			},
		}

		userInputJSON, _ := json.Marshal(userInput)
		req := httptest.NewRequest(http.MethodPost, "/register/teacher", strings.NewReader(string(userInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		role := "org:teacher"
		expectedOrgID := "org_2tjxBeJV0WLJUFU6Q3AwjzMyXTs"
		mockClerkUserID := "user_clerk_456"
		mockAuthService.On("CreateUser", userInput, expectedOrgID, role).Return(&domain.User{UserID: mockClerkUserID}, nil).Once()

		expectedUserDb := domain.UserDb{UserID: mockClerkUserID, Name: "Test", Lastname: "Teacher", Email: "test.teacher@example.com", TypeID: 2}
		mockUserRepo.On("CreateUser", expectedUserDb).Return(nil).Once()

		handler := userController.RegisterUser(role)
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedResp := `{"Body":{"message":"Usuario registrado con éxito"}}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockAuthService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 3: Failure - Invalid Role ---
	t.Run("Failure_InvalidRole", func(t *testing.T) {
		userInput := domain.UserInput{
			Name:     "Test",
			Lastname: "Teacher",
			Email:    "test.teacher@example.com",
			Representative: domain.RepresentativeInput{
				Name:        "Dummy",
				Lastname:    "Dummy",
				Email:       "dummy@example.com",
				PhoneNumber: "0000000000",
			},
		}

		userInputJSON, _ := json.Marshal(userInput)
		req := httptest.NewRequest(http.MethodPost, "/register/invalid", strings.NewReader(string(userInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		role := "invalid-role"
		handler := userController.RegisterUser(role)
		err := handler(c) // This will call ReturnWriteResponse internally

		// Assert based on how ReturnWriteResponse handles the specific error from GetTypeID
		assert.NoError(t, err) // The handler itself doesn't return error, it writes response
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		// The error message comes from GetTypeID's HTTPError
		expectedResp := `{"message":"code=400, message={Rol inválido}"}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockAuthService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 4: Failure - AuthService Error ---
	t.Run("Failure_AuthServiceError", func(t *testing.T) {
		userInput := domain.UserInput{
			Name:     "Test",
			Lastname: "Teacher",
			Email:    "test.teacher@example.com",
			Representative: domain.RepresentativeInput{
				Name:        "Dummy",
				Lastname:    "Dummy",
				Email:       "dummy@example.com",
				PhoneNumber: "0000000000",
			},
		}

		userInputJSON, _ := json.Marshal(userInput)
		req := httptest.NewRequest(http.MethodPost, "/register/student", strings.NewReader(string(userInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		role := "org:student"
		expectedOrgID := "org_2tjxBeJV0WLJUFU6Q3AwjzMyXTs"
		authError := errors.New("error al crear usuario en Clerk")
		mockAuthService.On("CreateUser", userInput, expectedOrgID, role).Return(nil, authError).Once()

		handler := userController.RegisterUser(role)
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Assuming ReturnWriteResponse maps this to 500
		// The specific error message comes from the handler
		expectedResp := `{"message":"error al crear usuario en Clerk"}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockAuthService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t) // UserRepo.CreateUser should not be called
	})

	// --- Test Case 6: Failure - Binding/Validation Error ---
	// Inside TestUserController_RegisterUser
	t.Run("Failure_BindingError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register/student", strings.NewReader(`{"name": "Test",`)) // Malformed JSON
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		role := "org:student"
		handler := userController.RegisterUser(role)
		err := handler(c)
		assert.Error(t, err)

		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Expected error to be *echo.HTTPError")

		if ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, struct {
				Message string `json:"message"`
			}{Message: "Invalid request body"}, httpErr.Message)
		}

		assert.Empty(t, rec.Body.String(), "Response body should be empty as error was returned early")
		mockAuthService.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserController_GetUser(t *testing.T) {
	mockUserRepo := new(domain.MockUserRepo)
	// No AuthService needed for GetUser
	userController := controller.UserController{
		UserRepo: mockUserRepo,
		// AuthService: nil, // Or a mock if other methods need it
	}

	// --- Test Case 1: Success ---
	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		c, rec := setupTest(req)

		mockUserID := "user_test_123"
		claims := &clerk.SessionClaims{
			Claims: jwt.Claims{
				Subject: mockUserID,
			},
		}
		c.Set("user", claims)

		expectedUser := &domain.UserDb{
			UserID:   mockUserID,
			Name:     "Test",
			Lastname: "User",
			Email:    "test@example.com",
			TypeID:   3,
		}
		mockUserRepo.On("GetUser", mockUserID).Return(expectedUser, nil).Once()

		handler := userController.GetUser()
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON, _ := json.Marshal(expectedUser)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
		mockUserRepo.AssertExpectations(t)
	})

	// Inside TestUserController_GetUser

	// --- Test Case 2: Failure - Unauthorized (No Claims) ---
	t.Run("Failure_Unauthorized_NoClaims", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		c, rec := setupTest(req)
		// Do NOT set "user" in context

		handler := userController.GetUser()
		err := handler(c) // Calls ReturnReadResponse with echo.ErrUnauthorized

		// --- Assertions Adjusted ---
		assert.Error(t, err) // Expect an error from the handler

		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Expected error to be *echo.HTTPError")

		if ok {
			assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{
				Message: "Unauthorized", // Default message for 401
			}
			assert.Equal(t, expectedMsgStruct, httpErr.Message, "Unexpected message structure in returned HTTPError")
		}

		assert.Empty(t, rec.Body.String(), "Response body should be empty")
		mockUserRepo.AssertExpectations(t) // GetUser should not be called
	})

	// --- Test Case 3: Failure - Unauthorized (Claims Wrong Type) ---
	t.Run("Failure_Unauthorized_WrongClaimType", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		c, rec := setupTest(req)
		c.Set("user", "not a claim object") // Set wrong type

		handler := userController.GetUser()
		err := handler(c) // Calls ReturnReadResponse with echo.ErrUnauthorized

		// --- Assertions Adjusted ---
		assert.Error(t, err) // Expect an error from the handler

		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Expected error to be *echo.HTTPError")

		if ok {
			assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
			// Expect the standard message structure
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{
				Message: "Unauthorized", // Default message for 401
			}
			assert.Equal(t, expectedMsgStruct, httpErr.Message, "Unexpected message structure in returned HTTPError")
		}

		assert.Empty(t, rec.Body.String(), "Response body should be empty")
		mockUserRepo.AssertExpectations(t) // GetUser should not be called
	})

	t.Run("Failure_UserRepoError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		c, rec := setupTest(req)

		mockUserID := "user_test_456"
		claims := &clerk.SessionClaims{
			Claims: jwt.Claims{
				Subject: mockUserID,
			},
		}
		c.Set("user", claims)

		repoError := errors.New("database error") // The original error from the repo
		mockUserRepo.On("GetUser", mockUserID).Return(nil, repoError).Once()

		handler := userController.GetUser()
		err := handler(c)
		assert.Error(t, err)

		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Expected error returned by handler to be *echo.HTTPError")
		if ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{Message: repoError.Error()}

			assert.Equal(t, expectedMsgStruct, httpErr.Message, "Unexpected message structure in returned HTTPError")
		}
		assert.Empty(t, rec.Body.String(), "Response body should be empty as error was returned by handler")
		mockUserRepo.AssertExpectations(t)
	})

}

func TestUserController_GetAllTeachers(t *testing.T) {
	mockUserRepo := new(domain.MockUserRepo)
	userController := controller.UserController{UserRepo: mockUserRepo}

	// --- Test Case 1: Success ---
	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
		c, rec := setupTest(req)

		expectedTeachers := []domain.UserDb{
			{UserID: "t1", Name: "Teacher", Lastname: "One", Email: "t1@example.com", TypeID: 2},
			{UserID: "t2", Name: "Teacher", Lastname: "Two", Email: "t2@example.com", TypeID: 2},
		}
		mockUserRepo.On("GetAllTeachers").Return(expectedTeachers, nil).Once()

		handler := userController.GetAllTeachers()
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON, _ := json.Marshal(expectedTeachers)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 2: Success - Empty List ---
	t.Run("Success_Empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
		c, rec := setupTest(req)

		expectedTeachers := []domain.UserDb{} // Empty slice
		mockUserRepo.On("GetAllTeachers").Return(expectedTeachers, nil).Once()

		handler := userController.GetAllTeachers()
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[]`, rec.Body.String()) // Expect empty JSON array
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 3: Failure - UserRepo Error ---
	t.Run("Failure_UserRepoError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
		c, rec := setupTest(req)
		repoError := errors.New("db query failed")
		mockUserRepo.On("GetAllTeachers").Return(nil, repoError).Once()
		handler := userController.GetAllTeachers()
		err := handler(c)
		assert.Error(t, err)
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Expected error to be *echo.HTTPError")
		if ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{
				Message: "Internal server error",
			}
			assert.Equal(t, expectedMsgStruct, httpErr.Message, "Unexpected message structure in returned HTTPError")
		}
		assert.Empty(t, rec.Body.String(), "Response body should be empty")
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserController_GetAllStudents(t *testing.T) {
	mockUserRepo := new(domain.MockUserRepo)
	userController := controller.UserController{UserRepo: mockUserRepo}

	// --- Test Case 1: Success ---
	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/students", nil)
		c, rec := setupTest(req)

		expectedStudents := []domain.UserDb{
			{UserID: "s1", Name: "Student", Lastname: "Alpha", Email: "s1@example.com", TypeID: 3},
			{UserID: "s2", Name: "Student", Lastname: "Beta", Email: "s2@example.com", TypeID: 3},
		}
		mockUserRepo.On("GetAllStudents").Return(expectedStudents, nil).Once()

		handler := userController.GetAllStudents()
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON, _ := json.Marshal(expectedStudents)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 2: Success - Empty List ---
	t.Run("Success_Empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/students", nil)
		c, rec := setupTest(req)

		expectedStudents := []domain.UserDb{} // Empty slice
		mockUserRepo.On("GetAllStudents").Return(expectedStudents, nil).Once()

		handler := userController.GetAllStudents()
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[]`, rec.Body.String()) // Expect empty JSON array
		mockUserRepo.AssertExpectations(t)
	})

	// --- Test Case 3: Failure - UserRepo Error ---
	t.Run("Failure_UserRepoError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/students", nil)
		c, rec := setupTest(req)

		repoError := errors.New("db query failed students")
		mockUserRepo.On("GetAllStudents").Return(nil, repoError).Once()

		handler := userController.GetAllStudents()
		err := handler(c)

		assert.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Expected error to be *echo.HTTPError")

		if ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{
				Message: "Internal server error", // Default message for unhandled errors
			}
			assert.Equal(t, expectedMsgStruct, httpErr.Message, "Unexpected message structure in returned HTTPError")
		}
		assert.Empty(t, rec.Body.String(), "Response body should be empty")
		mockUserRepo.AssertExpectations(t) // Mock was called
	})
}

// You might also want a separate test for GetTypeID if it were more complex
func TestGetTypeID(t *testing.T) {
	t.Run("Student", func(t *testing.T) {
		id, err := controller.GetTypeID("org:student")
		assert.NoError(t, err)
		assert.Equal(t, 3, id)
	})
	t.Run("Teacher", func(t *testing.T) {
		id, err := controller.GetTypeID("org:teacher")
		assert.NoError(t, err)
		assert.Equal(t, 2, id)
	})
	t.Run("Invalid", func(t *testing.T) {
		id, err := controller.GetTypeID("invalid")
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		// Check the inner message if needed
		expectedMsg := struct {
			Message string `json:"message"`
		}{Message: "Rol inválido"}
		assert.Equal(t, expectedMsg, httpErr.Message)

	})
}
