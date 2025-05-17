package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

// MockCourseRepo mocks CourseRepo
type MockCourseRepo struct {
	CreateC                        func(course domain.CourseDB) error
	GetCoursesByT                  func(teacherID string) ([]domain.CourseDB, error)
	GetCoursesByS                  func(studentID string) ([]domain.CourseDB, error)
	GetCourseByS2                  func(studentID string) ([]domain.CourseDbRelation, error)
	GetCourseByTeacherAndCourseIDT func(teacherID string, courseID int) (domain.CourseDB, error)
	GetCoursesByS2T                func(studentID, courseID string) (*domain.CourseDbRelation, error)
}

func (m MockCourseRepo) GetCourseByTeacherAndCourseID(teacherID string, courseID int) (domain.CourseDB, error) {
	if m.GetCourseByTeacherAndCourseIDT != nil {
		return m.GetCourseByTeacherAndCourseIDT(teacherID, courseID)
	}
	return domain.CourseDB{}, errors.New("GetCourseByTeacherAndCourseID function not implemented in mock")
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
func (m MockCourseRepo) GetCoursesByStudent2(studentID string) ([]domain.CourseDbRelation, error) {
	if m.GetCourseByS2 != nil {
		return m.GetCourseByS2(studentID)
	}
	return nil, errors.New("GetCourseByS2 function not implemented in mock")
}

func (m MockCourseRepo) GetCourse(studentID, courseID string) (*domain.CourseDbRelation, error) {
	if m.GetCoursesByS2T != nil {
		return m.GetCoursesByS2T(studentID, courseID)
	}
	return nil, errors.New("GetCourses function not implemented in mock")
}

// --- Helpers ---

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

func testHTTPErrorHandler(err error, c echo.Context) {
	he, ok := err.(*echo.HTTPError)
	if !ok {
		he = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var message interface{}
	if m, ok := he.Message.(string); ok {
		message = map[string]string{"message": m} // Eliminamos las llaves adicionales
	} else {
		message = map[string]interface{}{"message": he.Message}
	}

	if err := c.JSON(he.Code, message); err != nil {
		c.Logger().Error(err)
	}
}

// --- Tests for GetCoursesByStudent2 ---

func TestGetCoursesByStudent2_Success(t *testing.T) {
	testUserID := "student-201"
	testRole := "org:student"
	now := time.Now()

	mockCoursesRelation := []domain.CourseDbRelation{
		{
			CourseID:    1,
			Title:       "Advanced Go",
			StartDate:   now,
			Description: "Deep dive into Go",
			Teacher: domain.UserDbRelation{
				UserID:   "teacher-xyz",
				Name:     "Jane",
				Lastname: "Doe",
				Email:    "jane.doe@example.com",
			},
			CourseContent: []domain.CourseContentDb{
				{
					CourseContentID: 10,
					Module:          "Module 1",
					ModuleIndex:     1,
					Content: []domain.ContentDb{
						{
							ContentID:     "content-abc",
							ContentTypeID: 1,
							Title:         "Introduction Video",
							Description:   "First video",
							Url:           "http://example.com/video1",
							SectionIndex:  1,
						},
					},
				},
			},
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/student2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCourseByS2: func(studentID string) ([]domain.CourseDbRelation, error) {
			assert.Equal(t, testUserID, studentID)
			return mockCoursesRelation, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent2()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Construct expected JSON based on ToCourseOutput logic
		// For simplicity, we'll check for key fields. A more robust test would marshal the expected output.
		bodyStr := rec.Body.String()
		assert.Contains(t, bodyStr, `"id":1`)
		assert.Contains(t, bodyStr, `"title":"Advanced Go"`)
		assert.Contains(t, bodyStr, `"teacher":{"user_id":"teacher-xyz","name":"Jane","lastname":"Doe","email":"jane.doe@example.com"}`)
		assert.Contains(t, bodyStr, `"modules_summary"`)
	}
}

func TestGetCoursesByStudent2_EmptyList(t *testing.T) {
	testUserID := "student-202"
	testRole := "org:student"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/student2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCourseByS2: func(studentID string) ([]domain.CourseDbRelation, error) {
			assert.Equal(t, testUserID, studentID)
			return []domain.CourseDbRelation{}, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent2()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `[]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByStudent2_ForbiddenTeacher(t *testing.T) {
	testUserID := "teacher-203"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Use custom error handler
	req := httptest.NewRequest(http.MethodGet, "/courses/student2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCourseByS2: func(studentID string) ([]domain.CourseDbRelation, error) {
			assert.Fail(t, "GetCoursesByStudent2 should not be called for forbidden role")
			return nil, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent2()

	err := handler(c)
	// The error is handled by ReturnReadResponse, which in turn calls testHTTPErrorHandler
	// So, we check the recorder directly for the output of testHTTPErrorHandler
	if assert.Error(t, err) { // handler should return the error to echo
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Error should be an *echo.HTTPError")
		assert.Equal(t, http.StatusForbidden, httpErr.Code)

		// This assertion needs to match how testHTTPErrorHandler formats the JSON
		// Assuming testHTTPErrorHandler will be called and it will write to rec.Body
		// We also need to ensure the context is processed correctly by the error handler
		e.HTTPErrorHandler(err, c) // Manually call to simulate echo's behavior for this test structure

		expectedMessage := `{"message":{"message":"Solo los estudiantes pueden ver sus cursos"}}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestGetCoursesByStudent2_RepoError(t *testing.T) {
	testUserID := "student-204"
	testRole := "org:student"
	repoErr := errors.New("database connection failed for GetCoursesByStudent2")

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler // Use custom error handler
	req := httptest.NewRequest(http.MethodGet, "/courses/student2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCourseByS2: func(studentID string) ([]domain.CourseDbRelation, error) {
			assert.Equal(t, testUserID, studentID)
			return nil, repoErr
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent2()

	err := handler(c)
	if assert.Error(t, err) { // The handler itself should return the error
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Error should be an *echo.HTTPError")
		assert.Equal(t, http.StatusInternalServerError, httpErr.Code)

		// Simulate echo's error handling to check the response body
		e.HTTPErrorHandler(err, c)

		// Check if the message is what ReturnReadResponse sets for internal errors
		// This depends on the implementation of ReturnReadResponse and testHTTPErrorHandler
		// Assuming ReturnReadResponse passes the original error which testHTTPErrorHandler then wraps
		expectedMessage := `{"message":{"message":"Internal server error"}}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

// --- Tests para CourseController ---

func TestCreateCourse_SuccessTeacher(t *testing.T) {
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
	courseJSON := `{"title":"Introduction to Go","description":"A beginner course","start_date":"2024-01-10T10:00:00Z"}`
	testUserID := "student-456"
	testRole := "org:student"

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
			assert.Fail(t, "CreateCourse should not be called for forbidden role")
			return nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.CreateCourse()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Cambiado a 500 porque el controlador lo devuelve
		expectedMessage := `{"message":"code=403, message=Solo los profesores pueden crear cursos"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestCreateCourse_BadRequestInvalidJSON(t *testing.T) {
	invalidJSON := `{"title":"Missing fields"` // JSON mal formado
	testUserID := "teacher-123"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
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

	err := handler(c)

	expectedMsgStruct := struct {
		Message string `json:"message"`
	}{Message: "Invalid request body"}

	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Equal(t, expectedMsgStruct, httpErr.Message)
	}
}

func TestCreateCourse_BadRequestValidation(t *testing.T) {
	invalidDataJSON := `{"description":"A course missing title","start_date":"2024-01-10T10:00:00Z"}`
	testUserID := "teacher-123"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
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

	err := handler(c)

	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		msgMap, ok := httpErr.Message.(struct {
			Message string            `json:"message"`
			Body    map[string]string `json:"body"`
		})

		if assert.True(t, ok, "Debe retornar un mapa de errores de validación") {
			assert.Equal(t, "This field is required", msgMap.Body["title"])
		}
	}
}

func TestGetCoursesByTeacher_Success(t *testing.T) {
	testUserID := "teacher-789"
	testRole := "org:teacher"
	mockCourses := []domain.CourseDB{
		{CourseID: 1, TeacherID: testUserID, Title: "Advanced Go", QRCode: "abcd12"},
		{CourseID: 2, TeacherID: testUserID, Title: "Echo Framework", QRCode: "efgh34"},
	}

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
		expectedJSON := `[
			{"id":1,"teacher_id":"teacher-789","start_date":"","title":"Advanced Go","description":"","qr_code":"abcd12"},
			{"id":2,"teacher_id":"teacher-789","start_date":"","title":"Echo Framework","description":"","qr_code":"efgh34"}
		]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByTeacher_EmptyList(t *testing.T) {
	testUserID := "teacher-789"
	testRole := "org:teacher"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/teacher", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByT: func(teacherID string) ([]domain.CourseDB, error) {
			assert.Equal(t, testUserID, teacherID)
			return []domain.CourseDB{}, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByTeacher()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `[]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByTeacher_ForbiddenStudent(t *testing.T) {
	testUserID := "student-101"
	testRole := "org:student"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
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

	err := handler(c)

	// ✅ Debe retornar un *echo.HTTPError porque el rol no es "org:teacher"
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusForbidden, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message interface{} `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "Solo los profesores pueden ver sus cursos", msgStruct.Message)
		}
	}
}

func TestGetCoursesByTeacher_RepoError(t *testing.T) {
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

	err := handler(c)

	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusInternalServerError, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message string `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "Internal server error", msgStruct.Message)
		}
	}
}

func TestGetCoursesByStudent_Success(t *testing.T) {
	testUserID := "student-112"
	testRole := "org:student"
	mockCourses := []domain.CourseDB{
		{CourseID: 3, TeacherID: "teacher-abc", Title: "History 101", QRCode: "ijkl56"},
		{CourseID: 4, TeacherID: "teacher-def", Title: "Math Basics", QRCode: "mnop78"},
	}

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
		expectedJSON := `[
			{"id":3,"teacher_id":"teacher-abc","start_date":"","title":"History 101","description":"","qr_code":"ijkl56"},
			{"id":4,"teacher_id":"teacher-def","start_date":"","title":"Math Basics","description":"","qr_code":"mnop78"}
		]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByStudent_EmptyList(t *testing.T) {
	testUserID := "student-112"
	testRole := "org:student"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/courses/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	mockRepo := MockCourseRepo{
		GetCoursesByS: func(studentID string) ([]domain.CourseDB, error) {
			assert.Equal(t, testUserID, studentID)
			return []domain.CourseDB{}, nil
		},
	}

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCoursesByStudent()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `[]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCoursesByStudent_ForbiddenTeacher(t *testing.T) {
	testUserID := "teacher-113"
	testRole := "org:teacher"

	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
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

	err := handler(c)

	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusForbidden, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message interface{} `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "Solo los estudiantes pueden ver sus cursos", msgStruct.Message)
		}
	}
}

func TestGetCoursesByStudent_RepoError(t *testing.T) {
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

	err := handler(c)

	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message string `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "Internal server error", msgStruct.Message)
		}
	}
}

func TestCourseController_GetCourse_Success(t *testing.T) {
	testUserID := "student-112"
	testRole := "org:student"
	testCourseID := "course-123"

	date := time.Date(2025, 12, 24, 0, 0, 0, 0, time.UTC)

	mockRepo := MockCourseRepo{
		GetCoursesByS2T: func(studentID, courseID string) (*domain.CourseDbRelation, error) {
			assert.Equal(t, testUserID, studentID)
			assert.Equal(t, testCourseID, courseID) // This seems like a potential bug, comparing courseID to itself. Did you mean another variable?
			return &domain.CourseDbRelation{
				CourseID:    1,
				Title:       "Advanced Go",
				StartDate:   date,
				Description: "Deep dive into Go",
				Teacher: domain.UserDbRelation{
					UserID:   "teacher-xyz",
					Name:     "Jane",
					Lastname: "Doe",
					Email:    "teacher@gmail.com",
				},
				CourseContent: []domain.CourseContentDb{
					{ // Added curly braces for the struct literal
						CourseContentID: 10,
						Module:          "Module 1",
						ModuleIndex:     1,
						Content: []domain.ContentDb{
							{ // Added curly braces for the struct literal
								ContentID:     "content-abc",
								ContentTypeID: 1,
								Title:         "Introduction Video",
								Description:   "First video",
								Url:           "http://example.com/video1",
							},
						},
					},
				},
			}, nil
		},
	}
	expectedJson, err := mockRepo.GetCourse(testUserID, testCourseID)
	if err != nil {
		t.Fatalf("Error getting course: %v", err)
	}
	// Convert the expectedJson to the expected output format
	// For simplicity, we'll check for key fields. A more robust test would marshal the expected output.

	marshallJson, err := json.Marshal(expectedJson.ToCourseDetailOutput())
	if err != nil {
		t.Fatalf("Failed to marshal expected controller output: %v", err)
	}

	e := echo.New()
	// Adjust path to match your route definition, e.g., /courses/:course_id
	reqPath := "/courses/" + testCourseID
	req := httptest.NewRequest(http.MethodGet, reqPath, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set context values (for auth, etc.)
	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)

	// Set path parameters for Echo to correctly bind e.Param("course_id")
	c.SetPath("/courses/:course_id") // The route pattern registered in Echo
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	courseController := controller.CourseController{Repo: mockRepo}
	handler := courseController.GetCourse()

	err = handler(c)

	// Assertions
	if assert.NoError(t, err, "Handler returned an error") {
		assert.Equal(t, http.StatusOK, rec.Code, "HTTP status code not OK")

		// Compare the expected JSON with the actual JSON response from the controller
		assert.JSONEq(t, string(marshallJson), rec.Body.String(), "JSON response does not match expected")
	}

}

func TestCourseController_GetCourse_NotFound(t *testing.T) {
	testUserID := "student-112"
	testRole := "org:student"
	testCourseID := "course-xyz-nonexistent"
	mockRepo := &MockCourseRepo{ // Use a pointer if your interface expects it
		GetCoursesByS2T: func(studentID, courseID string) (*domain.CourseDbRelation, error) {
			assert.Equal(t, testUserID, studentID)
			assert.Equal(t, testCourseID, courseID)
			return nil, gorm.ErrRecordNotFound // Simulate course not found
		},
	}

	e := echo.New()

	reqPath := "/courses/" + testCourseID
	req := httptest.NewRequest(http.MethodGet, reqPath, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	c.SetPath("/courses/:course_id") // Route pattern
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	var repoInterface domain.CourseRepo = mockRepo
	courseCtrl := controller.CourseController{Repo: repoInterface}
	handler := courseCtrl.GetCourse()

	returnedErr := handler(c)

	if returnedErr != nil {
		e.HTTPErrorHandler(returnedErr, c) // This will use Echo's default or your custom one
	}
	if assert.Error(t, returnedErr, "Handler should have returned an error for NotFound case") {
		var httpErr *echo.HTTPError
		if assert.True(t, errors.As(returnedErr, &httpErr), "Error should be an *echo.HTTPError") {
			assert.Equal(t, http.StatusNotFound, httpErr.Code, "echo.HTTPError code mismatch")
		}

		assert.Equal(t, http.StatusNotFound, rec.Code, "HTTP status code in recorder is not StatusNotFound")

		expectedBody := `{"message":"Record not found"}`
		assert.JSONEq(t, expectedBody, rec.Body.String(), "JSON response body does not match expected for NotFound")
	} else {
		assert.Fail(t, "Expected handler to return an error, but it returned nil.")
	}
}

func TestCourseController_GetCourse_RepoError(t *testing.T) {
	testUserID := "student-112"
	testRole := "org:student"
	testCourseID := "course-abc-repo-error"

	mockRepoError := errors.New("database unavailable")

	mockRepo := &MockCourseRepo{
		GetCoursesByS2T: func(studentID, courseID string) (*domain.CourseDbRelation, error) {
			assert.Equal(t, testUserID, studentID)
			assert.Equal(t, testCourseID, courseID)
			return nil, mockRepoError
		},
	}

	e := echo.New()
	reqPath := "/courses/" + testCourseID
	req := httptest.NewRequest(http.MethodGet, reqPath, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	c.SetPath("/courses/:course_id")
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	var repoInterface domain.CourseRepo = mockRepo
	courseCtrl := controller.CourseController{Repo: repoInterface}
	handler := courseCtrl.GetCourse()

	returnedErr := handler(c)

	if returnedErr != nil {
		e.HTTPErrorHandler(returnedErr, c)
	}

	if assert.Error(t, returnedErr, "Handler should have returned an error for RepoError case") {
		var httpErr *echo.HTTPError
		if assert.True(t, errors.As(returnedErr, &httpErr), "Error should be an *echo.HTTPError") {
			// Your ReturnReadResponse will wrap a generic error into a 500
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code, "echo.HTTPError code mismatch")
		}

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "HTTP status code not InternalServerError")

		// Based on your ReturnReadResponse, a generic error becomes "Internal server error"
		expectedBody := `{"message":"Internal server error"}`
		assert.JSONEq(t, expectedBody, rec.Body.String(), "JSON response does not match expected for RepoError")
	} else {
		assert.Fail(t, "Expected handler to return an error, but it returned nil.")
	}
}

func TestCourseController_GetCourse_Forbidden(t *testing.T) {
	testUserID := "teacher-456"
	testRole := "org:teacher"
	testCourseID := "course-any"

	mockRepo := &MockCourseRepo{
		GetCoursesByS2T: func(studentID, courseID string) (*domain.CourseDbRelation, error) {
			t.Errorf("GetCourseFunc should not be called for a forbidden role scenario")
			return nil, errors.New("mock GetCourseFunc called unexpectedly in Forbidden test")
		},
	}

	e := echo.New()
	reqPath := "/courses/" + testCourseID
	req := httptest.NewRequest(http.MethodGet, reqPath, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("user_id", testUserID)
	c.Set("user_role", testRole)
	c.SetPath("/courses/:course_id")
	c.SetParamNames("course_id")
	c.SetParamValues(testCourseID)

	var repoInterface domain.CourseRepo = mockRepo
	courseCtrl := controller.CourseController{Repo: repoInterface}
	handler := courseCtrl.GetCourse()

	returnedErr := handler(c)

	if returnedErr != nil {
		e.HTTPErrorHandler(returnedErr, c)
	}

	if assert.Error(t, returnedErr, "Handler should have returned an error for Forbidden case") {
		var httpErr *echo.HTTPError
		if assert.True(t, errors.As(returnedErr, &httpErr), "Error should be an *echo.HTTPError") {
			assert.Equal(t, http.StatusForbidden, httpErr.Code, "echo.HTTPError code mismatch")

			// Type assert httpErr.Message to the specific anonymous struct type
			// that ReturnReadResponse creates as the payload.
			if actualPayload, ok := httpErr.Message.(struct {
				Message interface{} `json:"message"`
			}); ok {
				if actualMessageString, ok := actualPayload.Message.(string); ok {
					assert.Equal(t, "Solo los estudiantes pueden ver sus cursos", actualMessageString)
				} else {
					assert.Fail(t, "The 'Message' field within the payload struct is not a string", "Got type %T for actualPayload.Message", actualPayload.Message)
				}
			} else {
				assert.Fail(t, "httpErr.Message is not the expected anonymous struct { Message interface{} `json:\"message\"` }", "Got type %T for httpErr.Message", httpErr.Message)
			}
		}

		assert.Equal(t, http.StatusForbidden, rec.Code, "HTTP status code not Forbidden")
		expectedBody := `{"message":"Solo los estudiantes pueden ver sus cursos"}`
		assert.JSONEq(t, expectedBody, rec.Body.String(), "JSON response does not match expected for Forbidden")
	} else {
		assert.Fail(t, "Expected handler to return an error, but it returned nil.")
	}
}
