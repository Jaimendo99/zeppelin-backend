package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockCourseContentRepo mocks CourseContentRepo
type MockCourseContentRepo struct {
	GetContentByCourseF           func(courseID int, onlyActive bool) ([]domain.CourseContentWithDetails, error)
	CreateVideoF                  func(url, title, description string) (string, error)
	AddVideoSectionF              func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error
	CreateQuizF                   func(title, url, description string, jsonContent json.RawMessage) (string, error)
	AddQuizSectionF               func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error
	CreateTextF                   func(title, url string, jsonContent json.RawMessage) (string, error)
	AddTextSectionF               func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error
	UpdateVideoF                  func(contentID, title, url, description string) error
	UpdateQuizF                   func(contentID, title, url, description string, jsonContent json.RawMessage) error
	UpdateTextF                   func(contentID, title, url string, jsonContent json.RawMessage) error
	UpdateContentStatusF          func(contentID string, isActive bool) error
	UpdateModuleTitleF            func(courseContentID int, moduleTitle string) error
	GetContentByCourseForStudentF func(courseID int, isActive bool, userID string) ([]domain.CourseContentWithDetails, error)
	UpdateUserContentStatusF      func(userID, contentID string, statusID int) error
}

func (m MockCourseContentRepo) GetContentByCourse(courseID int, onlyActive bool) ([]domain.CourseContentWithDetails, error) {
	if m.GetContentByCourseF != nil {
		return m.GetContentByCourseF(courseID, onlyActive)
	}
	return nil, errors.New("GetContentByCourse not implemented")
}

func (m MockCourseContentRepo) CreateVideo(url, title, description string) (string, error) {
	if m.CreateVideoF != nil {
		return m.CreateVideoF(url, title, description)
	}
	return "", errors.New("CreateVideo not implemented")
}

func (m MockCourseContentRepo) AddVideoSection(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
	if m.AddVideoSectionF != nil {
		return m.AddVideoSectionF(courseID, contentID, module, sectionIndex, moduleIndex)
	}
	return errors.New("AddVideoSection not implemented")
}

func (m MockCourseContentRepo) CreateQuiz(title, url, description string, jsonContent json.RawMessage) (string, error) {
	if m.CreateQuizF != nil {
		return m.CreateQuizF(title, url, description, jsonContent)
	}
	return "", errors.New("CreateQuiz not implemented")
}

func (m MockCourseContentRepo) AddQuizSection(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
	if m.AddQuizSectionF != nil {
		return m.AddQuizSectionF(courseID, contentID, module, sectionIndex, moduleIndex)
	}
	return errors.New("AddQuizSection not implemented")
}

func (m MockCourseContentRepo) CreateText(title, url string, jsonContent json.RawMessage) (string, error) {
	if m.CreateTextF != nil {
		return m.CreateTextF(title, url, jsonContent)
	}
	return "", errors.New("CreateText not implemented")
}

func (m MockCourseContentRepo) AddTextSection(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
	if m.AddTextSectionF != nil {
		return m.AddTextSectionF(courseID, contentID, module, sectionIndex, moduleIndex)
	}
	return errors.New("AddTextSection not implemented")
}

func (m MockCourseContentRepo) UpdateVideo(contentID, title, url, description string) error {
	if m.UpdateVideoF != nil {
		return m.UpdateVideoF(contentID, title, url, description)
	}
	return errors.New("UpdateVideo not implemented")
}

func (m MockCourseContentRepo) UpdateQuiz(contentID, title, url, description string, jsonContent json.RawMessage) error {
	if m.UpdateQuizF != nil {
		return m.UpdateQuizF(contentID, title, url, description, jsonContent)
	}
	return errors.New("UpdateQuiz not implemented")
}

func (m MockCourseContentRepo) UpdateText(contentID, title, url string, jsonContent json.RawMessage) error {
	if m.UpdateTextF != nil {
		return m.UpdateTextF(contentID, title, url, jsonContent)
	}
	return errors.New("UpdateText not implemented")
}

func (m MockCourseContentRepo) UpdateContentStatus(contentID string, isActive bool) error {
	if m.UpdateContentStatusF != nil {
		return m.UpdateContentStatusF(contentID, isActive)
	}
	return errors.New("UpdateContentStatus not implemented")
}

func (m MockCourseContentRepo) UpdateModuleTitle(courseContentID int, moduleTitle string) error {
	if m.UpdateModuleTitleF != nil {
		return m.UpdateModuleTitleF(courseContentID, moduleTitle)
	}
	return errors.New("UpdateModuleTitle not implemented")
}

func setupContext(e *echo.Echo, req *http.Request, rec *httptest.ResponseRecorder, userID, userRole string) echo.Context {
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	if userRole != "" {
		c.Set("user_role", userRole)
	}
	return c
}

func (m MockCourseContentRepo) GetContentByCourseForStudent(courseID int, isActive bool, userID string) ([]domain.CourseContentWithDetails, error) {
	if m.GetContentByCourseForStudentF != nil {
		return m.GetContentByCourseForStudentF(courseID, isActive, userID)
	}
	return nil, errors.New("GetContentByCourseForStudent not implemented")
}

func (m MockCourseContentRepo) UpdateUserContentStatus(userID, contentID string, statusID int) error {
	if m.UpdateUserContentStatusF != nil {
		return m.UpdateUserContentStatusF(userID, contentID, statusID)
	}
	return errors.New("UpdateUserContentStatus not implemented")
}

// --- Tests para CourseContentController ---

func TestGetCourseContentTeacher_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/teacher?course_id=1", nil)
	rec := httptest.NewRecorder()
	c := setupContext(e, req, rec, "teacher-123", "")

	videoDetails := domain.VideoContent{
		ContentID:   "video-123",
		Url:         "https://example.com/video",
		Title:       "Video 1",
		Description: "Intro",
	}
	quizDetails := domain.QuizContent{
		ContentID:   "quiz-456",
		Title:       "Quiz 1",
		Description: "Test",
	}

	mockContent := []domain.CourseContentWithDetails{
		{
			CourseContentDB: domain.CourseContentDB{
				CourseContentID: 1,
				CourseID:        1,
				ContentID:       "video-123",
				ContentType:     "video",
				Module:          "Module 1",
				SectionIndex:    1,
				ModuleIndex:     0,
				IsActive:        true,
				CreatedAt:       time.Time{},
			},
			Details: videoDetails,
		},
		{
			CourseContentDB: domain.CourseContentDB{
				CourseContentID: 2,
				CourseID:        1,
				ContentID:       "quiz-456",
				ContentType:     "quiz",
				Module:          "Module 1",
				SectionIndex:    2,
				ModuleIndex:     0,
				IsActive:        false,
				CreatedAt:       time.Time{},
			},
			Details: quizDetails,
		},
	}

	mockCourseRepo := MockCourseRepo{
		GetCourseByTeacherAndCourseIDT: func(teacherID string, courseID int) (domain.CourseDB, error) {
			assert.Equal(t, "teacher-123", teacherID)
			assert.Equal(t, 1, courseID)
			return domain.CourseDB{CourseID: 1}, nil
		},
	}

	mockContentRepo := MockCourseContentRepo{
		GetContentByCourseF: func(courseID int, onlyActive bool) ([]domain.CourseContentWithDetails, error) {
			assert.Equal(t, 1, courseID)
			assert.False(t, onlyActive) // Teachers get all content
			return mockContent, nil
		},
	}

	controller := controller.CourseContentController{
		Repo:       mockContentRepo,
		RepoCourse: mockCourseRepo,
	}
	handler := controller.GetCourseContentTeacher()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `[
			{
				"course_content_id": 1,
				"course_id": 1,
				"content_id": "video-123",
				"content_type": "video",
				"module": "Module 1",
				"section_index": 1,
				"module_index": 0,
				"is_active": true,
				"created_at": "0001-01-01T00:00:00Z",
				"Details": {"content_id": "video-123", "url": "https://example.com/video", "title": "Video 1", "description": "Intro"}
			},
			{
				"course_content_id": 2,
				"course_id": 1,
				"content_id": "quiz-456",
				"content_type": "quiz",
				"module": "Module 1",
				"section_index": 2,
				"module_index": 0,
				"is_active": false,
				"created_at": "0001-01-01T00:00:00Z",
				"Details": {"content_id": "quiz-456", "title": "Quiz 1", "description": "Test"}
			}
		]`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestGetCourseContentTeacher_InvalidCourseID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/teacher?course_id=invalid", nil)
	rec := httptest.NewRecorder()
	c := setupContext(e, req, rec, "teacher-123", "")

	controller := controller.CourseContentController{}
	handler := controller.GetCourseContentTeacher()

	err := handler(c)

	// 🔥 Aquí verificamos que efectivamente se retorne un error
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")

		// ✅ Verificamos el código y mensaje esperados
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		// El httpErr.Message es un struct { Message interface{} }
		msgStruct, ok := httpErr.Message.(struct {
			Message interface{} `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "course_id inválido", msgStruct.Message)
		}

	}
}

func TestGetCourseContentTeacher_CourseNotOwned(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/teacher?course_id=1", nil)
	rec := httptest.NewRecorder()
	c := setupContext(e, req, rec, "teacher-123", "")

	mockCourseRepo := MockCourseRepo{
		GetCourseByTeacherAndCourseIDT: func(teacherID string, courseID int) (domain.CourseDB, error) {
			return domain.CourseDB{}, errors.New("course not found") // Dispara error esperado
		},
	}

	controller := controller.CourseContentController{
		RepoCourse: mockCourseRepo,
	}
	handler := controller.GetCourseContentTeacher()

	err := handler(c)

	// ✅ Validar que se devuelve un *echo.HTTPError con status 403 y el mensaje correcto
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")

		assert.Equal(t, http.StatusForbidden, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message interface{} `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "Este curso no le pertenece al profesor", msgStruct.Message)
		}
	}
}

func TestGetCourseContentForStudent_NonStudentRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/student?course_id=1", nil)
	rec := httptest.NewRecorder()
	c := setupContext(e, req, rec, "teacher-123", "org:teacher") // Rol incorrecto

	controller := controller.CourseContentController{}
	handler := controller.GetCourseContentForStudent()

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	expectedJSON := `{"message": "code=403, message=Solo los estudiantes pueden ver el contenido de los cursos"}`
	assert.JSONEq(t, expectedJSON, rec.Body.String())
}

func TestGetCourseContentForStudent_InvalidCourseID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/student?course_id=invalid", nil)
	rec := httptest.NewRecorder()
	c := setupContext(e, req, rec, "student-123", "org:student")

	controller := controller.CourseContentController{}
	handler := controller.GetCourseContentForStudent()

	err := handler(c)

	// ✅ El handler devuelve un *echo.HTTPError porque usa ReturnReadResponse
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")

		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message interface{} `json:"message"`
		})
		if assert.True(t, ok) {
			assert.Equal(t, "course_id inválido", msgStruct.Message)
		}
	}
}

type MockAssignmentRepos struct {
	GetAssignmentsByStudentAndCourseF func(studentID string, courseID int) (domain.AssignmentWithCourse, error)
}

func (m *MockAssignmentRepos) CreateAssignment(userID string, courseID int) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockAssignmentRepos) VerifyAssignment(assignmentID int) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockAssignmentRepos) GetAssignmentsByStudent(userID string) ([]domain.AssignmentWithCourse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockAssignmentRepos) GetStudentsByCourse(courseID int) ([]domain.AssignmentWithStudent, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockAssignmentRepos) GetCourseIDByQRCode(qrCode string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockAssignmentRepos) GetAssignmentsByStudentAndCourse(studentID string, courseID int) (domain.AssignmentWithCourse, error) {
	if m.GetAssignmentsByStudentAndCourseF != nil {
		return m.GetAssignmentsByStudentAndCourseF(studentID, courseID)
	}
	return domain.AssignmentWithCourse{}, errors.New("GetAssignmentsByStudentAndCourse not implemented")
}

func TestAddVideoSection_Success(t *testing.T) {
	videoJSON := `{"url":"https://example.com/video","title":"Video 1","description":"Intro","module":"Module 1","section_index":1,"module_index":0}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/video?course_id=1", strings.NewReader(videoJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		CreateVideoF: func(url, title, description string) (string, error) {
			assert.Equal(t, "https://example.com/video", url)
			assert.Equal(t, "Video 1", title)
			assert.Equal(t, "Intro", description)
			return "video-123", nil
		},
		AddVideoSectionF: func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
			assert.Equal(t, 1, courseID)
			assert.Equal(t, "video-123", contentID)
			assert.Equal(t, "Module 1", module)
			assert.Equal(t, 1, sectionIndex)
			assert.Equal(t, 0, moduleIndex)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.AddVideoSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Sección de video agregada","content_id":"video-123"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestAddVideoSection_InvalidCourseID(t *testing.T) {
	videoJSON := `{"url":"https://example.com/video","title":"Video 1","description":"Intro","module":"Module 1","section_index":1,"module_index":0}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/video?course_id=invalid", strings.NewReader(videoJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	controller := controller.CourseContentController{}
	handler := controller.AddVideoSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Or 400 if ReturnWriteResponse is fixed
		expectedJSON := `{"message": "code=400, message=course_id inválido"}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestAddVideoSection_ValidationError(t *testing.T) {
	videoJSON := `{"url":"","title":"Video 1","description":"Intro","module":"","section_index":1,"module_index":0}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/video?course_id=1", strings.NewReader(videoJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	controller := controller.CourseContentController{}
	handler := controller.AddVideoSection()

	err := handler(c)

	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		expectedMessageStruct := struct {
			Message string            `json:"message"`
			Body    map[string]string `json:"body"`
		}{
			Message: "Error on body parameters",
			Body: map[string]string{
				"url":    "This field is required",
				"module": "This field is required",
			},
		}
		assert.Equal(t, expectedMessageStruct, httpErr.Message, "El mensaje debe contener los errores de validación esperados")
	}
}

func TestAddQuizSection_Success(t *testing.T) {
	quizJSON := `{"title":"Quiz 1","description":"Test","url":"https://example.com/quiz","module":"Module 1","section_index":2,"module_index":1}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/quiz?course_id=1", strings.NewReader(quizJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		CreateQuizF: func(title, url, description string, jsonContent json.RawMessage) (string, error) {
			assert.Equal(t, "Quiz 1", title)
			assert.Equal(t, "https://example.com/quiz", url)
			assert.Equal(t, "Test", description)
			assert.Nil(t, jsonContent)
			return "quiz-456", nil
		},
		AddQuizSectionF: func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
			assert.Equal(t, 1, courseID)
			assert.Equal(t, "quiz-456", contentID)
			assert.Equal(t, "Module 1", module)
			assert.Equal(t, 2, sectionIndex)
			assert.Equal(t, 1, moduleIndex)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.AddQuizSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Sección de quiz agregada","content_id":"quiz-456"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}
func TestAddQuizSection_CreateQuizError(t *testing.T) {
	quizJSON := `{"title":"Quiz 1","description":"Test","url":"https://example.com/quiz","module":"Module 1","section_index":2,"module_index":1}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/quiz?course_id=1", strings.NewReader(quizJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	repoErr := errors.New("failed to create quiz")
	mockRepo := MockCourseContentRepo{
		CreateQuizF: func(title, url, description string, jsonContent json.RawMessage) (string, error) {
			assert.Equal(t, "Quiz 1", title)
			assert.Equal(t, "https://example.com/quiz", url)
			assert.Equal(t, "Test", description)
			assert.Nil(t, jsonContent)
			return "", repoErr
		},
		AddQuizSectionF: func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
			assert.Fail(t, "AddQuizSection should not be called")
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.AddQuizSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedJSON := `{"message":"failed to create quiz"}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}
func TestAddTextSection_Success(t *testing.T) {
	textJSON := `{"title":"Text 1","module":"Module 1","section_index":3,"module_index":2}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/text?course_id=1", strings.NewReader(textJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		CreateTextF: func(title, url string, jsonContent json.RawMessage) (string, error) {
			assert.Equal(t, "Text 1", title)
			assert.Equal(t, "", url)
			assert.Nil(t, jsonContent)
			return "text-789", nil
		},
		AddTextSectionF: func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
			assert.Equal(t, 1, courseID)
			assert.Equal(t, "text-789", contentID)
			assert.Equal(t, "Module 1", module)
			assert.Equal(t, 3, sectionIndex)
			assert.Equal(t, 2, moduleIndex)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.AddTextSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Sección de texto agregada","content_id":"text-789"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestAddTextSection_CreateTextError(t *testing.T) {
	textJSON := `{"title":"Text 1","module":"Module 1","section_index":3,"module_index":2}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/content/text?course_id=1", strings.NewReader(textJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	repoErr := errors.New("failed to create text")
	mockRepo := MockCourseContentRepo{
		CreateTextF: func(title, url string, jsonContent json.RawMessage) (string, error) {
			return "", repoErr
		},
		AddTextSectionF: func(courseID int, contentID, module string, sectionIndex, moduleIndex int) error {
			assert.Fail(t, "AddTextSection should not be called")
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.AddTextSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedJSON := `{"message":"failed to create text"}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestUpdateVideoContent_Success(t *testing.T) {
	videoJSON := `{"content_id":"video-123","title":"Updated Video","url":"https://example.com/updated","description":"Updated Intro"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/video", strings.NewReader(videoJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateVideoF: func(contentID, title, url, description string) error {
			assert.Equal(t, "video-123", contentID)
			assert.Equal(t, "Updated Video", title)
			assert.Equal(t, "https://example.com/updated", url)
			assert.Equal(t, "Updated Intro", description)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateVideoContent()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Video actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestUpdateVideoContent_ValidationError(t *testing.T) {
	videoJSON := `{"content_id":"","title":"Updated Video"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/video", strings.NewReader(videoJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	controller := controller.CourseContentController{}
	handler := controller.UpdateVideoContent()

	err := handler(c)

	// Esperamos un *echo.HTTPError
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message string            `json:"message"`
			Body    map[string]string `json:"body"`
		})

		if assert.True(t, ok, "Debe retornar un struct con mensaje y mapa de errores") {
			if assert.True(t, ok, "Debe retornar un mapa de errores de validación") {
				assert.Equal(t, "This field is required", msgStruct.Body["contentid"])
			}
		}
	}
}

func TestUpdateQuizContent_Success(t *testing.T) {
	quizJSON := `{"content_id":"quiz-456","title":"Updated Quiz","description":"Updated Test","url":"https://example.com/updated-quiz","json_content":{"questions":[{"id":1,"text":"Q1"}]}}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/quiz", strings.NewReader(quizJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	var jsonContent json.RawMessage
	json.Unmarshal([]byte(`{"questions":[{"id":1,"text":"Q1"}]}`), &jsonContent)

	mockRepo := MockCourseContentRepo{
		UpdateQuizF: func(contentID, title, url, description string, jsonContentArg json.RawMessage) error {
			assert.Equal(t, "quiz-456", contentID)
			assert.Equal(t, "Updated Quiz", title)
			assert.Equal(t, "https://example.com/updated-quiz", url)
			assert.Equal(t, "Updated Test", description)
			assert.JSONEq(t, string(jsonContent), string(jsonContentArg))
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateQuizContent()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Quiz actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}
func TestUpdateTextContent_Success(t *testing.T) {
	textJSON := `{"content_id":"text-789","title":"Updated Text","url":"https://example.com/text","json_content":{"content":"Updated content"}}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/text", strings.NewReader(textJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	var jsonContent json.RawMessage
	json.Unmarshal([]byte(`{"content":"Updated content"}`), &jsonContent)

	mockRepo := MockCourseContentRepo{
		UpdateTextF: func(contentID, title, url string, jsonContentArg json.RawMessage) error {
			assert.Equal(t, "text-789", contentID)
			assert.Equal(t, "Updated Text", title)
			assert.Equal(t, "https://example.com/text", url)
			assert.JSONEq(t, string(jsonContent), string(jsonContentArg))
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateTextContent()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Texto actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestUpdateContentStatus_Success(t *testing.T) {
	statusJSON := `{"content_id":"content-123","is_active":true}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/status", strings.NewReader(statusJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateContentStatusF: func(contentID string, isActive bool) error {
			assert.Equal(t, "content-123", contentID)
			assert.True(t, isActive)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateContentStatus()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Estado del contenido actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestUpdateModuleTitle_Success(t *testing.T) {
	moduleJSON := `{"course_content_id":1,"module_title":"Updated Module"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/module", strings.NewReader(moduleJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateModuleTitleF: func(courseContentID int, moduleTitle string) error {
			assert.Equal(t, 1, courseContentID)
			assert.Equal(t, "Updated Module", moduleTitle)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateModuleTitle()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Título del módulo actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestUpdateModuleTitle_ValidationError(t *testing.T) {
	moduleJSON := `{"course_content_id":0,"module_title":""}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/module", strings.NewReader(moduleJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	controller := controller.CourseContentController{}
	handler := controller.UpdateModuleTitle()

	err := handler(c)

	// ✅ Validamos que retorne error de validación
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok, "Debe retornar un *echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message string            `json:"message"`
			Body    map[string]string `json:"body"`
		})
		if assert.True(t, ok, "Debe ser un mapa de errores de validación") {
			assert.Equal(t, "This field is required", msgStruct.Body["coursecontentid"])
			assert.Equal(t, "This field is required", msgStruct.Body["moduletitle"])
		}
	}
}

func TestGetCourseContentForStudent_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/student?course_id=1", nil)
	rec := httptest.NewRecorder()
	c := setupContext(e, req, rec, "student-123", "org:student")

	mockContent := []domain.CourseContentWithDetails{
		{
			CourseContentDB: domain.CourseContentDB{
				CourseContentID: 1,
				CourseID:        1,
				ContentID:       "content-001",
				ContentType:     "video",
				Module:          "Module 1",
				SectionIndex:    0,
				ModuleIndex:     0,
				IsActive:        true,
				CreatedAt:       time.Now(),
			},
			Details:  nil,
			StatusID: intPtr(2),
		},
	}

	mockRepo := MockCourseContentRepo{
		GetContentByCourseForStudentF: func(courseID int, onlyActive bool, userID string) ([]domain.CourseContentWithDetails, error) {
			assert.Equal(t, 1, courseID)
			assert.True(t, onlyActive)
			assert.Equal(t, "student-123", userID)
			return mockContent, nil
		},
	}

	mockAssignment := &MockAssignmentRepos{
		GetAssignmentsByStudentAndCourseF: func(studentID string, courseID int) (domain.AssignmentWithCourse, error) {
			assert.Equal(t, "student-123", studentID)
			assert.Equal(t, 1, courseID)
			return domain.AssignmentWithCourse{}, nil
		},
	}

	controller := controller.CourseContentController{
		Repo:          mockRepo,
		RepoAssigment: mockAssignment,
	}

	handler := controller.GetCourseContentForStudent()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status_id":2`)
		assert.Contains(t, rec.Body.String(), `"content_id":"content-001"`)
	}
}

func intPtr(i int) *int {
	return &i
}

func TestUpdateUserContentStatusHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateUserContentStatusF: func(userID, contentID string, statusID int) error {
			assert.Equal(t, "student-123", userID)
			assert.Equal(t, "content-001", contentID)
			assert.True(t, statusID == 2 || statusID == 3)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}

	cases := []struct {
		path         string
		expectedMsg  string
		expectedCode int
	}{
		{"/in-progress", "Contenido marcado como 'en progreso'", http.StatusOK},
		{"/completed", "Contenido marcado como 'completado'", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(strings.Title(strings.TrimPrefix(tc.path, "/")), func(t *testing.T) {
			body := `{"content_id":"content-001"}`
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.Set("user_id", "student-123")

			var handler echo.HandlerFunc
			if tc.path == "/in-progress" {
				handler = controller.UpdateUserContentStatus(2)
			} else {
				handler = controller.UpdateUserContentStatus(3)
			}

			err := handler(c)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedCode, rec.Code)

				var responseBody map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
				assert.NoError(t, err)

				expected := map[string]interface{}{
					"Body": map[string]interface{}{
						"message": tc.expectedMsg,
					},
				}

				assert.Equal(t, expected, responseBody)
			}
		})
	}
}

func TestUpdateUserContentStatusHandler_ValidationError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/course-content/in-progress", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "student-001")
	e.Validator = &CustomValidator{Validator: validator.New()}

	controller := controller.CourseContentController{}
	handler := controller.UpdateUserContentStatus(2)

	err := handler(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		msgStruct, ok := httpErr.Message.(struct {
			Message string            `json:"message"`
			Body    map[string]string `json:"body"`
		})

		if assert.True(t, ok) {
			assert.Equal(t, "This field is required", msgStruct.Body["contentid"])
		}
	}
}

func TestUpdateUserContentStatusHandler_ErrorFromRepo(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/course-content/in-progress", strings.NewReader(`{"content_id": "abc123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "student-001")
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateUserContentStatusF: func(userID, contentID string, statusID int) error {
			return errors.New("failed to update status")
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateUserContentStatus(2)

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"message":"failed to update status"}`, rec.Body.String())
	}
}
