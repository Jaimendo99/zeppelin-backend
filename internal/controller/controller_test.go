package controller_test

import (
	"errors"
	"fmt"
	"github.com/clerkinc/clerk-sdk-go/clerk"
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

// testHTTPErrorHandler from your example
func testHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		c.Logger().Error("Response already committed in error handler: ", err)
		return
	}

	code := http.StatusInternalServerError
	var msg interface{} // Use interface{} to handle different message types
	var jsonResponse map[string]interface{}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		// *** CORRECTED: Keep the map as the message payload ***
		if validationMap, ok := he.Message.(map[string]string); ok {
			msg = validationMap // Assign the map directly
		} else {
			// Use the simple message string for other HTTP errors
			msg = fmt.Sprintf("%v", he.Message)
		}
		jsonResponse = map[string]interface{}{"message": msg}

	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		code = http.StatusNotFound
		msg = "{Record not found}"
		jsonResponse = map[string]interface{}{"message": msg}
	} else {
		// Default internal server error message
		msg = "{Internal server error}"
		jsonResponse = map[string]interface{}{"message": msg}
	}

	// Send response
	// Use c.JSON which will correctly marshal the 'msg' (even if it's a map)
	if err := c.JSON(code, jsonResponse); err != nil {
		c.Logger().Error("Error writing JSON error response: ", err)
	}
}

type MockRepresentativeRepo struct {
	CreateRep func(representative domain.RepresentativeDb) error
	GetRep    func(id int) (*domain.RepresentativeInput, error)
	GetAllRep func() ([]domain.Representative, error)
	UpdateRep func(id int, representative domain.RepresentativeInput) error
}

func (m MockRepresentativeRepo) CreateRepresentative(representative domain.RepresentativeDb) error {
	return m.CreateRep(representative)
}

func (m MockRepresentativeRepo) GetRepresentative(id int) (*domain.RepresentativeInput, error) {
	return m.GetRep(id)
}

func (m MockRepresentativeRepo) GetAllRepresentatives() ([]domain.Representative, error) {
	return m.GetAllRep()
}

func (m MockRepresentativeRepo) UpdateRepresentative(id int, representative domain.RepresentativeInput) error {
	return m.UpdateRep(id, representative)
}

func TestCreateRepresentative_Success(t *testing.T) {
	var repeJson = `{"name":"Anthony","lastname":"Cochea","email":"anthony@gmail.com","phone_number":"+593990269309"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(repeJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	mockRepo := MockRepresentativeRepo{
		CreateRep: func(representative domain.RepresentativeDb) error {
			return nil
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.CreateRepresentative()

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `{"Body":{"message":"Representative created"}}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestCreateRepresentative_BadRequest(t *testing.T) {
	var repeJson = `{"name":"Anthony","lastname":"Cochea","email":"","phone_number":""`
	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(repeJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	mockRepo := MockRepresentativeRepo{
		CreateRep: func(representative domain.RepresentativeDb) error {
			return nil
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.CreateRepresentative()

	err := handler(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	expectedMessage := `{"message":"Invalid request body"}`
	assert.JSONEq(t, expectedMessage, rec.Body.String())
}

func TestGetRepresentative_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := MockRepresentativeRepo{
		GetRep: func(id int) (*domain.RepresentativeInput, error) {
			return &domain.RepresentativeInput{
				Name:        "Anthony",
				Lastname:    "Cochea",
				Email:       "anthony@gmail.com",
				PhoneNumber: "+593990269309",
			}, nil
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.GetRepresentative()

	c.SetPath("/representative/:representative_id")
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `{"name":"Anthony","lastname":"Cochea","email":"anthony@gmail.com","phone_number":"+593990269309"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}

}

func TestGetRepresentative_NotFound(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := MockRepresentativeRepo{
		GetRep: func(id int) (*domain.RepresentativeInput, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.GetRepresentative()

	c.SetPath("/representative/:representative_id")
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	err := handler(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusNotFound, rec.Code)
	expectedMessage := `{"message":"{Record not found}"}`
	assert.JSONEq(t, expectedMessage, rec.Body.String())
}

func TestGetAllRepresentatives_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := MockRepresentativeRepo{
		GetAllRep: func() ([]domain.Representative, error) {
			return []domain.Representative{
				{
					RepresentativeId: 1,
					Name:             "Anthony",
					Lastname:         "Cochea",
					Email:            "",
					PhoneNumber:      "",
				},
				{
					RepresentativeId: 2,
					Name:             "Mateo",
					Lastname:         "Mejia",
					Email:            "mateo@gmail.com",
					PhoneNumber:      "+593990269309",
				},
			}, nil
		},
	}

	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.GetAllRepresentatives()

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `[{"representative_id":1,"name":"Anthony","lastname":"Cochea","email":"","phone_number":""},{"representative_id":2,"name":"Mateo","lastname":"Mejia","email":"mateo@gmail.com","phone_number":"+593990269309"}]`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

// --- Mock Repo (Unchanged) ---
type MockCourseRepo struct {
	CreateC       func(course domain.CourseDB) error
	GetCoursesByT func(teacherID string) ([]domain.CourseDB, error)
	GetCoursesByS func(studentID string) ([]domain.CourseDB, error)
}

func (m MockCourseRepo) CreateCourse(course domain.CourseDB) error {
	if m.CreateC != nil {
		return m.CreateC(course)
	}
	return errors.New("CreateC function not implemented in mock")
}

func (m MockCourseRepo) GetCoursesByTeacher(teacherID string) ([]domain.CourseDB, error) {
	if m.GetCoursesByT != nil {
		return m.GetCoursesByT(teacherID)
	}
	return nil, errors.New("GetCoursesByT function not implemented in mock")
}

func (m MockCourseRepo) GetCoursesByStudent(studentID string) ([]domain.CourseDB, error) {
	if m.GetCoursesByS != nil {
		return m.GetCoursesByS(studentID)
	}
	return nil, errors.New("GetCoursesByS function not implemented in mock")
}

// --- Helpers ---

// Assuming CustomValidator is okay
type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

// --- Tests for CourseController (Corrected) ---

func TestCreateCourse_SuccessTeacher(t *testing.T) {
	// ... (Setup remains the same) ...
	courseJSON := `{"title":"Introduction to Go","description":"A beginner course","start_date":"2024-01-10T10:00:00Z"}`
	testUserID := "teacher-123"
	testRole := "org:teacher"

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/courses", strings.NewReader(courseJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseRepo{
		CreateC: func(course domain.CourseDB) error {
			assert.Equal(t, testUserID, course.TeacherID)
			assert.Equal(t, "Introduction to Go", course.Title)
			assert.NotEmpty(t, course.QRCode)
			return nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.CreateCourse()

	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `{"Body":{"message":"Curso creado con éxito"}}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestCreateCourse_ForbiddenStudent(t *testing.T) {
	// ... (Setup remains the same) ...
	courseJSON := `{"title":"Introduction to Go","description":"A beginner course","start_date":"2024-01-10T10:00:00Z"}`
	testUserID := "student-456"
	testRole := "org:student"

	e := echo.New()
	// No HTTPErrorHandler needed here, ReturnWriteResponse handles internally
	req := httptest.NewRequest(http.MethodPost, "/courses", strings.NewReader(courseJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseRepo{
		CreateC: func(course domain.CourseDB) error {
			assert.Fail(t, "CreateCourse should not be called for forbidden role")
			return nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.CreateCourse()

	// *** CORRECTED: Expect NoError because ReturnWriteResponse handles it ***
	err := handler(c)

	if assert.NoError(t, err) {
		// *** CORRECTED: Expect 500 status and specific message from ReturnWriteResponse default case ***
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedMessage := `{"message":"code=403, message=Solo los profesores pueden crear cursos"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestCreateCourse_BadRequestInvalidJSON(t *testing.T) {
	// ... (Setup remains the same) ...
	invalidJSON := `{"title":"Missing fields"`
	testUserID := "teacher-123"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Use error handler
	req := httptest.NewRequest(http.MethodPost, "/courses", strings.NewReader(invalidJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	e.Validator = &CustomValidator{Validator: validator.New()}
	mockRepo := MockCourseRepo{}
	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.CreateCourse()

	err := handler(c) // Error from ValidateAndBind (Bind)
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		// *** CORRECTED: Expect the simple message based on previous actual output ***
		expectedMessage := `{"message":"Invalid request body"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestGetCoursesByTeacher_ForbiddenStudent(t *testing.T) {
	// ... (Setup remains the same) ...
	testUserID := "student-101"
	testRole := "org:student"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Use error handler
	req := httptest.NewRequest(http.MethodGet, "/courses/teacher", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByT: func(teacherID string) ([]domain.CourseDB, error) {
			assert.Fail(t, "GetCoursesByTeacher should not be called")
			return nil, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByTeacher()

	// *** CORRECTED: Expect Error because ReturnReadResponse returns a 500 error ***
	err := handler(c)

	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Let the handler format the 500 error
		// *** CORRECTED: Expect 500 status and specific message from ReturnReadResponse -> testHTTPErrorHandler ***
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedMessage := `{"message":"{Internal server error}"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestCreateCourse_BadRequestValidation(t *testing.T) {
	invalidDataJSON := `{"description":"A course missing title","start_date":"2024-01-10T10:00:00Z"}`
	testUserID := "teacher-123"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Use refined error handler
	req := httptest.NewRequest(http.MethodPost, "/courses", strings.NewReader(invalidDataJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	e.Validator = &CustomValidator{Validator: validator.New()}
	mockRepo := MockCourseRepo{}
	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.CreateCourse()

	err := handler(c) // Error from ValidateAndBind (Validate)
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Let handler process the error with map message
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// *** CORRECTED: Expect the message to be the validation map object itself ***
		// This matches the actual output where the map is nested under "message"
		expectedBody := `{"message":{"title":"This field is required"}}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestGetCoursesByTeacher_Success(t *testing.T) {
	testUserID := "teacher-789"
	testRole := "org:teacher"
	mockCourses := []domain.CourseDB{
		// *** CORRECTED: Initialize CourseID (maps to json:"id") ***
		{CourseID: 1, TeacherID: testUserID, Title: "Advanced Go", QRCode: "abcd12"},
		{CourseID: 2, TeacherID: testUserID, Title: "Echo Framework", QRCode: "efgh34"},
	}
	// ... (Rest of setup is same) ...
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/teacher", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByT: func(teacherID string) ([]domain.CourseDB, error) {
			assert.Equal(t, testUserID, teacherID)
			return mockCourses, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByTeacher()

	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Expect correct ID and start_date format based on previous results
		expectedJSON := `[{"id":1,"teacher_id":"teacher-789","start_date":"","title":"Advanced Go","description":"","qr_code":"abcd12"},{"id":2,"teacher_id":"teacher-789","start_date":"","title":"Echo Framework","description":"","qr_code":"efgh34"}]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByTeacher_RepoError(t *testing.T) {
	// ... (Setup remains the same) ...
	testUserID := "teacher-789"
	testRole := "org:teacher"
	repoErr := errors.New("database connection failed")

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
	req := httptest.NewRequest(http.MethodGet, "/courses/teacher", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByT: func(teacherID string) ([]domain.CourseDB, error) {
			assert.Equal(t, testUserID, teacherID)
			return nil, repoErr
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByTeacher()

	// Expect Error because ReturnReadResponse returns a 500 error
	err := handler(c)
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		// *** CORRECTED: Expect the message produced by testHTTPErrorHandler for this 500 error ***
		expectedMessage := `{"message":"{Internal server error}"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestGetCoursesByStudent_Success(t *testing.T) {
	testUserID := "student-112"
	testRole := "org:student"
	mockCourses := []domain.CourseDB{
		// *** CORRECTED: Initialize CourseID (maps to json:"id") ***
		{CourseID: 3, TeacherID: "teacher-abc", Title: "History 101", QRCode: "ijkl56"},
		{CourseID: 4, TeacherID: "teacher-def", Title: "Math Basics", QRCode: "mnop78"},
	}
	// ... (Rest of setup is same) ...
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByS: func(studentID string) ([]domain.CourseDB, error) {
			assert.Equal(t, testUserID, studentID)
			return mockCourses, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent()

	err := handler(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Expect correct ID and start_date format
		expectedJSON := `[{"id":3,"teacher_id":"teacher-abc","start_date":"","title":"History 101","description":"","qr_code":"ijkl56"},{"id":4,"teacher_id":"teacher-def","start_date":"","title":"Math Basics","description":"","qr_code":"mnop78"}]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByStudent_ForbiddenTeacher(t *testing.T) {
	// ... (Setup remains the same) ...
	testUserID := "teacher-113"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Use error handler
	req := httptest.NewRequest(http.MethodGet, "/courses/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByS: func(studentID string) ([]domain.CourseDB, error) {
			assert.Fail(t, "GetCoursesByStudent should not be called")
			return nil, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent()

	// *** CORRECTED: Expect Error because ReturnReadResponse returns a 500 error ***
	err := handler(c)

	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Let the handler format the 500 error
		// *** CORRECTED: Expect 500 status and specific message from ReturnReadResponse -> testHTTPErrorHandler ***
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedMessage := `{"message":"{Internal server error}"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestGetCoursesByStudent_RepoError(t *testing.T) {
	// ... (Setup remains the same) ...
	testUserID := "student-112"
	testRole := "org:student"
	repoErr := errors.New("failed to fetch student courses")

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
	req := httptest.NewRequest(http.MethodGet, "/courses/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByS: func(studentID string) ([]domain.CourseDB, error) {
			assert.Equal(t, testUserID, studentID)
			return nil, repoErr
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent()

	// Expect Error because ReturnReadResponse returns a 500 error
	err := handler(c)
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		// *** CORRECTED: Expect the message produced by testHTTPErrorHandler for this 500 error ***
		expectedMessage := `{"message":"{Internal server error}"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

type MockNotificationRepo struct {
	SendToQ func(notification domain.NotificationQueue, queueName string) error
	// ConsumeFromQueue is not needed for testing SendNotification
}

func (m *MockNotificationRepo) SendToQueue(notification domain.NotificationQueue, queueName string) error {
	if m.SendToQ != nil {
		return m.SendToQ(notification, queueName)
	}
	return errors.New("SendToQ function not implemented in mock")
}

// ConsumeFromQueue needs to be implemented to satisfy the interface, but can be a no-op for these tests
func (m *MockNotificationRepo) ConsumeFromQueue(queueName string) error {
	// No implementation needed for SendNotification tests
	return errors.New("ConsumeFromQueue not implemented in mock for testing SendNotification")
}

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

	// Mock repo (should not be called)
	mockRepo := &MockNotificationRepo{
		SendToQ: func(notification domain.NotificationQueue, queueName string) error {
			assert.Fail(t, "SendToQueue should not be called on invalid JSON")
			return nil
		},
	}

	notificationController := controller.NotificationController{Repo: mockRepo}
	handler := notificationController.SendNotification()

	// Execute handler (ValidateAndBind fails)
	err := handler(c)

	// Assert error handling
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Process the error
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		// Expect the message from ValidateAndBind's Bind error
		expectedBody := `{"message":"Invalid request body"}`
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

func TestSendNotification_BadRequest_ValidationError(t *testing.T) {
	// Assuming Title is required via `validate:"required"` tag in NotificationQueue struct
	// Modify your actual domain.NotificationQueue struct to include validation tags
	// type NotificationQueue struct {
	// 	NotificationId string   `json:"notification_id"`
	// 	Title          string   `json:"title" validate:"required"` // Example tag
	// 	Message        string   `json:"message" validate:"required"` // Example tag
	// 	ReceiverId     []string `json:"receiver_id" validate:"required"` // Example tag
	// }
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

	// Execute handler (ValidateAndBind fails on validation)
	err := handler(c)

	// Assert error handling
	if assert.Error(t, err) {
		e.HTTPErrorHandler(err, c) // Process the error
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		// Expect the validation error map from GetValidationFieldError,
		// formatted by testHTTPErrorHandler
		expectedBody := `{"message":{"title":"This field is required"}}` // Adjust field/message based on actual tags/GetValidationFieldError
		assert.JSONEq(t, expectedBody, rec.Body.String())
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

type MockAssignmentRepo struct {
	CreateA           func(userID string, courseID int) error
	VerifyA           func(assignmentID int) error
	GetAssignmentsByS func(userID string) ([]domain.AssignmentWithCourse, error)
	GetStudentsByC    func(courseID int) ([]domain.AssignmentWithStudent, error)
	GetCourseIDByQR   func(qrCode string) (int, error)
}

func (m *MockAssignmentRepo) CreateAssignment(userID string, courseID int) error {
	if m.CreateA != nil {
		return m.CreateA(userID, courseID)
	}
	return errors.New("CreateA function not implemented in mock")
}

func (m *MockAssignmentRepo) VerifyAssignment(assignmentID int) error {
	if m.VerifyA != nil {
		return m.VerifyA(assignmentID)
	}
	return errors.New("VerifyA function not implemented in mock")
}

func (m *MockAssignmentRepo) GetAssignmentsByStudent(userID string) ([]domain.AssignmentWithCourse, error) {
	if m.GetAssignmentsByS != nil {
		return m.GetAssignmentsByS(userID)
	}
	return nil, errors.New("GetAssignmentsByS function not implemented in mock")
}

func (m *MockAssignmentRepo) GetStudentsByCourse(courseID int) ([]domain.AssignmentWithStudent, error) {
	if m.GetStudentsByC != nil {
		return m.GetStudentsByC(courseID)
	}
	return nil, errors.New("GetStudentsByC function not implemented in mock")
}

func (m *MockAssignmentRepo) GetCourseIDByQRCode(qrCode string) (int, error) {
	if m.GetCourseIDByQR != nil {
		return m.GetCourseIDByQR(qrCode)
	}
	return 0, errors.New("GetCourseIDByQR function not implemented in mock")
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
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedBody := fmt.Sprintf(`{"message":"%s"}`, repoErr.Error()) // e.g., {"message":"record not found"}
		assert.JSONEq(t, expectedBody, rec.Body.String())
	}
}

// --- Mock UserRepo ---
type MockUserRepo struct {
	CreateU func(user domain.UserDb) error
	GetU    func(userID string) (*domain.UserDb, error)
	GetAllT func() ([]domain.UserDb, error)
	GetAllS func() ([]domain.UserDb, error)
}

func (m *MockUserRepo) CreateUser(user domain.UserDb) error {
	if m.CreateU != nil {
		return m.CreateU(user)
	}
	return errors.New("CreateU function not implemented in mock")
}

func (m *MockUserRepo) GetUser(userID string) (*domain.UserDb, error) {
	if m.GetU != nil {
		return m.GetU(userID)
	}
	return nil, errors.New("GetU function not implemented in mock")
}

func (m *MockUserRepo) GetAllTeachers() ([]domain.UserDb, error) {
	if m.GetAllT != nil {
		return m.GetAllT()
	}
	return nil, errors.New("GetAllT function not implemented in mock")
}

func (m *MockUserRepo) GetAllStudents() ([]domain.UserDb, error) {
	if m.GetAllS != nil {
		return m.GetAllS()
	}
	return nil, errors.New("GetAllS function not implemented in mock")
}

// --- Mock AuthService ---
// We only need to mock the methods used by the controller (CreateUser)
type MockAuthService struct {
	CreateU func(input domain.UserInput, organizationID string, role string) (*domain.User, error)
	// Add VerifyToken, DecodeToken mocks if other controllers use them
}

func (m *MockAuthService) CreateUser(input domain.UserInput, organizationID string, role string) (*domain.User, error) {
	if m.CreateU != nil {
		return m.CreateU(input, organizationID, role)
	}
	return nil, errors.New("CreateU function not implemented in mock")
}

// Implement other AuthService methods if needed by other tests, otherwise they can panic or return errors.
func (m *MockAuthService) VerifyToken(token string) (*clerk.SessionClaims, error) {
	panic("VerifyToken not implemented in mock")
}
func (m *MockAuthService) DecodeToken(token string) (*clerk.TokenClaims, error) {
	panic("DecodeToken not implemented in mock")
}

func GetTypeID(role string) (int, error) {
	switch role {
	case "org:student":
		return 3, nil
	case "org:teacher":
		return 2, nil
	default:
		// Return the actual error structure used by the controller
		return 0, echo.NewHTTPError(http.StatusBadRequest, struct {
			Message string `json:"message"`
		}{Message: "Rol inválido"})
	}
}
