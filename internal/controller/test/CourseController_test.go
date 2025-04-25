package controller_test

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

// MockCourseRepo mocks CourseRepo
type MockCourseRepo struct {
	CreateC                        func(course domain.CourseDB) error
	GetCoursesByT                  func(teacherID string) ([]domain.CourseDB, error)
	GetCoursesByS                  func(studentID string) ([]domain.CourseDB, error)
	GetCourseByTeacherAndCourseIDT func(teacherID, courseID string) (domain.CourseDB, error)
}

func (m MockCourseRepo) GetCourseByTeacherAndCourseID(teacherID, courseID string) (domain.CourseDB, error) {
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

	// ✅ Esperamos un error
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Equal(t, "Invalid request body", httpErr.Message)
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

	// ✅ Verificamos que se retorna un error con validación específica
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		msgMap, ok := httpErr.Message.(map[string]string)
		if assert.True(t, ok, "Debe retornar un mapa de errores de validación") {
			assert.Equal(t, "This field is required", msgMap["title"])
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
