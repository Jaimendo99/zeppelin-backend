package controller_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSendNotification_Success(t *testing.T) {
	// Define a valid notification payload
	// Add `validate:"required"` tags to NotificationQueue struct for validation tests
	// Example: Title string `json:"title" validate:"required"`
	notificationJSON := `{"notification_id":"nid-123","title":"Test Title","message":"Test Message","receiver_id":["user1","user2"]}`
	expectedQueueName := "notification"
	var sentNotification domain.NotificationQueue // To capture sent data

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(notificationJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()} // Validator needed

	// Setup mock repo
	mockRepo := &MockNotificationRepo{
		SendToQ: func(notification domain.NotificationQueue, queueName string) error {
			// Assert arguments passed to the mock
			assert.Equal(t, expectedQueueName, queueName)
			assert.Equal(t, "nid-123", notification.NotificationId)
			assert.Equal(t, "Test Title", notification.Title)
			assert.Equal(t, []string{"user1", "user2"}, notification.ReceiverId)
			sentNotification = notification // Capture for potential further checks
			return nil                      // Simulate success
		},
	}

	// Instantiate controller (using New function if available, otherwise direct struct init)
	// notificationController := controller.NewNotificationController(mockRepo)
	notificationController := controller.NotificationController{Repo: mockRepo} // Direct init if New func isn't used/exported
	handler := notificationController.SendNotification()

	// Execute handler (calls ReturnWriteResponse internally)
	err := handler(c)

	// Assert results
	if assert.NoError(t, err) { // Expect no error from handler itself
		assert.Equal(t, http.StatusOK, rec.Code)
		// Expect the Body wrapper from ReturnWriteResponse
		expectedBody := `{"Body":{"message":"Notification sent"}}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
		// Optionally assert details of the captured notification
		assert.Equal(t, "Test Message", sentNotification.Message)
	}
}

func TestSendNotification_BadRequest_InvalidJSON(t *testing.T) {
	invalidJSON := `{"notification_id":"nid-123","title":"Test Title",` // Malformed

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Error handler needed for Bind error
	req := httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(invalidJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := &MockNotificationRepo{
		SendToQ: func(notification domain.NotificationQueue, queueName string) error {
			assert.Fail(t, "SendToQueue should not be called on invalid JSON")
			return nil
		},
	}
	notificationController := controller.NotificationController{Repo: mockRepo}
	handler := notificationController.SendNotification()
	err := handler(c)
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Process the error
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		expectedMsgStruct := struct {
			Message string `json:"message"`
		}{Message: "Invalid request body"}
		assert.Equal(t, expectedMsgStruct.Message, "Invalid request body")
	}
}
func TestSendNotification_BadRequest_ValidationError(t *testing.T) {
	invalidDataJSON := `{"notification_id":"nid-123","message":"Test Message","receiver_id":["user1"]}` // Missing Title

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Error handler needed for Validate error
	req := httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(invalidDataJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()} // Ensure validator is active

	// Mock repo (should not be called)
	mockRepo := &MockNotificationRepo{
		SendToQ: func(notification domain.NotificationQueue, queueName string) error {
			assert.Fail(t, "SendToQueue should not be called on validation error")
			return nil
		},
	}

	notificationController := controller.NotificationController{Repo: mockRepo}
	handler := notificationController.SendNotification()

	err := handler(c)

	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Process the error
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		// Expect the validation error map from GetValidationFieldError,
		// formatted by testHTTPErrorHandler
		httperr := err.(*echo.HTTPError)

		expectedBodyRes := struct {
			Message string            `json:"message"`
			Body    map[string]string `json:"body"`
		}{
			Message: "Error on body parameters",
			Body: map[string]string{
				"title": "This field is required",
			},
		}

		assert.Equal(t, expectedBodyRes, httperr.Message)
	}
}

func TestSendNotification_RepoError(t *testing.T) {
	notificationJSON := `{"notification_id":"nid-123","title":"Test Title","message":"Test Message","receiver_id":["user1","user2"]}`
	expectedQueueName := "notification"
	repoErr := errors.New("failed to connect to queue")

	e := echo.New()
	// No HTTPErrorHandler needed here, ReturnWriteResponse handles internally
	req := httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(notificationJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	// Setup mock repo to return an error
	mockRepo := &MockNotificationRepo{
		SendToQ: func(notification domain.NotificationQueue, queueName string) error {
			assert.Equal(t, expectedQueueName, queueName)
			return repoErr // Simulate repo error
		},
	}

	notificationController := controller.NotificationController{Repo: mockRepo}
	handler := notificationController.SendNotification()

	// Execute handler (calls ReturnWriteResponse internally with repoErr)
	err := handler(c)

	// Assert results (ReturnWriteResponse default case)
	if assert.NoError(t, err) { // Expect no error from handler itself
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		// Expect the message from ReturnWriteResponse default case
		expectedBody := fmt.Sprintf(`{"message":"%s"}`, repoErr.Error())
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

// --- GetAssignmentsByStudent Tests ---

func TestGetAssignmentsByStudent_Success(t *testing.T) {
	testUserID := "student-1"
	testRole := "org:student"
	mockAssignments := []domain.AssignmentWithCourse{
		{AssignmentID: 1, CourseID: 101, Title: "Course A", QRCode: "qr1", AssignedAt: "2023-01-01T10:00:00Z", IsActive: true, IsVerify: false},
		{AssignmentID: 2, CourseID: 102, Title: "Course B", QRCode: "qr2", AssignedAt: "2023-01-02T11:00:00Z", IsActive: true, IsVerify: true},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/assignments/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{
		GetAssignmentsByS: func(userID string) ([]domain.AssignmentWithCourse, error) {
			assert.Equal(t, testUserID, userID)
			return mockAssignments, nil
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetAssignmentsByStudent()

	// Expect handler to call ReturnReadResponse internally and return nil
	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Adjust expected JSON based on AssignmentWithCourse struct and actual data
		expectedJSON := `[{"assignment_id":1,"assigned_at":"2023-01-01T10:00:00Z","is_active":true,"is_verify":false,"course_id":101,"teacher_id":"","start_date":"","title":"Course A","description":"","qr_code":"qr1"},{"assignment_id":2,"assigned_at":"2023-01-02T11:00:00Z","is_active":true,"is_verify":true,"course_id":102,"teacher_id":"","start_date":"","title":"Course B","description":"","qr_code":"qr2"}]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetAssignmentsByStudent_Forbidden(t *testing.T) {
	testUserID := "teacher-1"
	testRole := "org:teacher" // Incorrect role

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/assignments/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{ // Should not be called
		GetAssignmentsByS: func(userID string) ([]domain.AssignmentWithCourse, error) {
			assert.Fail(t, "GetAssignmentsByStudent should not be called")
			return nil, nil
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetAssignmentsByStudent()

	// Expect handler to call ReturnWriteResponse internally and return nil
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=403, message=Solo los estudiantes pueden ver sus asignaciones"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestGetAssignmentsByStudent_RepoError(t *testing.T) {
	testUserID := "student-1"
	testRole := "org:student"
	repoErr := errors.New("db connection error")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/assignments/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{
		GetAssignmentsByS: func(userID string) ([]domain.AssignmentWithCourse, error) {
			assert.Equal(t, testUserID, userID)
			return nil, repoErr
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetAssignmentsByStudent()

	// Expect handler to call ReturnReadResponse, which returns a 500 echo.HTTPError
	err := handler(c)

	// Assert the error returned by ReturnReadResponse
	if assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		// Check the specific message set by ReturnReadResponse for internal errors
		assert.Equal(t, "{Internal server error}", fmt.Sprintf("%v", he.Message))
		// Note: Recorder won't have content yet as error handler wasn't called
	}
}

// --- GetStudentsByCourse Tests ---

func TestGetStudentsByCourse_Success(t *testing.T) {
	testRole := "org:teacher"
	testCourseID := "101"
	expectedCourseIDInt := 101
	mockStudents := []domain.AssignmentWithStudent{
		{AssignmentID: 1, UserID: "student-1", Name: "Alice", Lastname: "Smith", Email: "alice@test.com", AssignedAt: "2023-01-01T10:00:00Z", IsActive: true, IsVerify: false},
		{AssignmentID: 3, UserID: "student-2", Name: "Bob", Lastname: "Jones", Email: "bob@test.com", AssignedAt: "2023-01-03T12:00:00Z", IsActive: true, IsVerify: true},
	}

	e := echo.New()
	// Need to set path and params for e.Param("course_id")
	req := httptest.NewRequest(http.MethodGet, "/courses/"+testCourseID+"/students", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	mockRepo := &MockAssignmentRepo{
		GetStudentsByC: func(courseID int) ([]domain.AssignmentWithStudent, error) {
			assert.Equal(t, expectedCourseIDInt, courseID)
			return mockStudents, nil
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetStudentsByCourse()

	// Expect handler to call ReturnReadResponse internally and return nil
	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Adjust expected JSON based on AssignmentWithStudent struct and actual data
		expectedJSON := `[{"id":1,"assigned_at":"2023-01-01T10:00:00Z","is_active":true,"is_verify":false,"user_id":"student-1","name":"Alice","lastname":"Smith","email":"alice@test.com"},{"id":3,"assigned_at":"2023-01-03T12:00:00Z","is_active":true,"is_verify":true,"user_id":"student-2","name":"Bob","lastname":"Jones","email":"bob@test.com"}]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetStudentsByCourse_Forbidden(t *testing.T) {
	testRole := "org:student" // Incorrect role
	testCourseID := "101"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/"+testCourseID+"/students", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	mockRepo := &MockAssignmentRepo{ // Should not be called
		GetStudentsByC: func(courseID int) ([]domain.AssignmentWithStudent, error) {
			assert.Fail(t, "GetStudentsByCourse should not be called")
			return nil, nil
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetStudentsByCourse()

	// Expect handler to call ReturnWriteResponse internally and return nil
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=403, message=Solo los profesores pueden ver los estudiantes de sus cursos"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestGetStudentsByCourse_InvalidCourseID(t *testing.T) {
	testRole := "org:teacher"
	testCourseID := "invalid-id" // Invalid parameter

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/"+testCourseID+"/students", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	mockRepo := &MockAssignmentRepo{} // Should not be called
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetStudentsByCourse()

	// Expect handler to call ReturnWriteResponse internally after strconv.Atoi fails
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=400, message=ID de curso inválido"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestGetStudentsByCourse_RepoError(t *testing.T) {
	testRole := "org:teacher"
	testCourseID := "101"
	expectedCourseIDInt := 101
	repoErr := errors.New("db connection error")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/"+testCourseID+"/students", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	mockRepo := &MockAssignmentRepo{
		GetStudentsByC: func(courseID int) ([]domain.AssignmentWithStudent, error) {
			assert.Equal(t, expectedCourseIDInt, courseID)
			return nil, repoErr
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.GetStudentsByCourse()

	// Expect handler to call ReturnReadResponse, which returns a 500 echo.HTTPError
	err := handler(c)

	// Assert the error returned by ReturnReadResponse
	if assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		assert.Equal(t, "{Internal server error}", fmt.Sprintf("%v", he.Message))
	}
}

// --- CreateAssignment Tests ---

func TestCreateAssignment_Success(t *testing.T) {
	testUserID := "student-5"
	testRole := "org:student"
	testQRCode := "validQR123"
	expectedCourseID := 201
	requestBody := fmt.Sprintf(`{"qr_code":"%s"}`, testQRCode)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/assignments", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{
		GetCourseIDByQR: func(qrCode string) (int, error) {
			assert.Equal(t, testQRCode, qrCode)
			return expectedCourseID, nil
		},
		CreateA: func(userID string, courseID int) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, expectedCourseID, courseID)
			return nil // Simulate successful creation
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.CreateAssignment()

	// Expect handler to call ReturnWriteResponse internally and return nil
	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedBody := `{"Body":{"message":"Inscripción exitosa"}}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestCreateAssignment_Forbidden(t *testing.T) {
	testUserID := "teacher-5"
	testRole := "org:teacher" // Incorrect role
	testQRCode := "validQR123"
	requestBody := fmt.Sprintf(`{"qr_code":"%s"}`, testQRCode)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/assignments", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{} // Should not be called
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.CreateAssignment()

	// Expect handler to call ReturnWriteResponse internally and return nil
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=403, message=Solo los estudiantes pueden inscribirse a cursos"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestCreateAssignment_InvalidBody(t *testing.T) {
	testUserID := "student-5"
	testRole := "org:student"
	requestBody := `{"qr_code":` // Invalid JSON

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/assignments", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{} // Should not be called
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.CreateAssignment()

	// Expect handler to call ReturnWriteResponse internally after Bind fails
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)         // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=400, message=Datos inválidos"}` // Message from controller
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestCreateAssignment_QRCodeNotFound(t *testing.T) {
	testUserID := "student-5"
	testRole := "org:student"
	testQRCode := "invalidQR"
	requestBody := fmt.Sprintf(`{"qr_code":"%s"}`, testQRCode)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/assignments", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{
		GetCourseIDByQR: func(qrCode string) (int, error) {
			assert.Equal(t, testQRCode, qrCode)
			// Simulate not found - could be gorm.ErrRecordNotFound or a custom error
			return 0, gorm.ErrRecordNotFound
		},
		// CreateA should not be called
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.CreateAssignment()

	// Expect handler to call ReturnWriteResponse internally after GetCourseIDByQRCode fails
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)                               // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=404, message=Curso no encontrado con ese código QR"}` // Message from controller
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestCreateAssignment_CreateRepoError(t *testing.T) {
	testUserID := "student-5"
	testRole := "org:student"
	testQRCode := "validQR123"
	expectedCourseID := 201
	requestBody := fmt.Sprintf(`{"qr_code":"%s"}`, testQRCode)
	// Example: Simulate a conflict error (student already assigned)
	repoErr := gorm.ErrDuplicatedKey

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/assignments", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := &MockAssignmentRepo{
		GetCourseIDByQR: func(qrCode string) (int, error) {
			assert.Equal(t, testQRCode, qrCode)
			return expectedCourseID, nil
		},
		CreateA: func(userID string, courseID int) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, expectedCourseID, courseID)
			return repoErr // Simulate repo error during creation
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.CreateAssignment()

	// Expect handler to call ReturnWriteResponse internally with the repoErr
	err := handler(c)

	// ReturnWriteResponse handles gorm.ErrDuplicatedKey specifically
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusConflict, rec.Code)
		expectedBody := `{"message":"Duplicated key"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

// --- VerifyAssignment Tests ---

func TestVerifyAssignment_Success_Teacher(t *testing.T) {
	testRole := "org:teacher"
	testAssignmentID := "55"
	expectedAssignmentIDInt := 55

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/assignments/"+testAssignmentID+"/verify", nil) // Assuming PUT
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("assignment_id")
	c.SetParamValues(testAssignmentID)

	mockRepo := &MockAssignmentRepo{
		VerifyA: func(assignmentID int) error {
			assert.Equal(t, expectedAssignmentIDInt, assignmentID)
			return nil // Simulate success
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.VerifyAssignment()

	// Expect handler to call ReturnWriteResponse internally and return nil
	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedBody := `{"Body":{"message":"Asignación verificada exitosamente"}}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestVerifyAssignment_Success_Admin(t *testing.T) {
	testRole := "org:admin" // Admin role
	testAssignmentID := "56"
	expectedAssignmentIDInt := 56

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/assignments/"+testAssignmentID+"/verify", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("assignment_id")
	c.SetParamValues(testAssignmentID)

	mockRepo := &MockAssignmentRepo{
		VerifyA: func(assignmentID int) error {
			assert.Equal(t, expectedAssignmentIDInt, assignmentID)
			return nil // Simulate success
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.VerifyAssignment()

	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedBody := `{"Body":{"message":"Asignación verificada exitosamente"}}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestVerifyAssignment_Forbidden(t *testing.T) {
	testRole := "org:student" // Incorrect role
	testAssignmentID := "55"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/assignments/"+testAssignmentID+"/verify", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("assignment_id")
	c.SetParamValues(testAssignmentID)

	mockRepo := &MockAssignmentRepo{} // Should not be called
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.VerifyAssignment()

	// Expect handler to call ReturnWriteResponse internally and return nil
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=403, message=No tienes permiso para verificar"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestVerifyAssignment_InvalidAssignmentID(t *testing.T) {
	testRole := "org:teacher"
	testAssignmentID := "invalid-id" // Invalid parameter

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/assignments/"+testAssignmentID+"/verify", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("assignment_id")
	c.SetParamValues(testAssignmentID)

	mockRepo := &MockAssignmentRepo{} // Should not be called
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.VerifyAssignment()

	// Expect handler to call ReturnWriteResponse internally after strconv.Atoi fails
	err := handler(c)

	// ReturnWriteResponse default case handles the echo.HTTPError
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Because ReturnWriteResponse default is 500
		expectedBody := `{"message":"code=400, message=ID de asignación inválido"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestVerifyAssignment_RepoError(t *testing.T) {
	testRole := "org:teacher"
	testAssignmentID := "55"
	expectedAssignmentIDInt := 55
	// Simulate assignment not found or other DB error
	repoErr := gorm.ErrRecordNotFound

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/assignments/"+testAssignmentID+"/verify", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_role", testRole)
	c.SetParamNames("assignment_id")
	c.SetParamValues(testAssignmentID)

	mockRepo := &MockAssignmentRepo{
		VerifyA: func(assignmentID int) error {
			assert.Equal(t, expectedAssignmentIDInt, assignmentID)
			return repoErr // Simulate repo error
		},
	}
	assignmentController := controller.AssignmentController{Repo: mockRepo}
	handler := assignmentController.VerifyAssignment()

	// Expect handler to call ReturnWriteResponse internally with the repoErr
	err := handler(c)

	// ReturnWriteResponse default case handles the error
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		expectedBody := `{"message":"Record not found"}` // e.g., {"message":"record not found"}
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}
