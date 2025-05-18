package controller_test

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

// Mocks
func TestCourseContentController_AddModule_Success(t *testing.T) {
	userID := "teacher-123"
	inputJSON := `{"course_id":1,"module":"New Module"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/module", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockContentRepo := MockCourseContentRepo{
		AddModuleT: func(courseID int, module string, uid string) (int, error) {
			assert.Equal(t, 1, courseID)
			assert.Equal(t, "New Module", module)
			assert.Equal(t, userID, uid)
			return 1, nil
		},
	}

	controller := controller.CourseContentController{Repo: mockContentRepo}
	handler := controller.AddModule()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Módulo creado","course_content_id":1,"module":"New Module"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestCourseContentController_AddModule_InvalidInput(t *testing.T) {
	userID := "teacher-123"
	inputJSON := `{"course_id":0,"module":""}` // Invalid: course_id and module required

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/module", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	e.Validator = &CustomValidator{Validator: validator.New()}
	e.HTTPErrorHandler = testHTTPErrorHandler

	mockContentRepo := MockCourseContentRepo{}

	controller := controller.CourseContentController{Repo: mockContentRepo}
	handler := controller.AddModule()

	err := handler(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	}
}

func TestCourseContentController_AddSection_Success(t *testing.T) {
	userID := "teacher-123"
	inputJSON := `{"course_content_id":1,"content_type_id":1,"title":"New Section","description":"Section desc"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/section", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockContentRepo := MockCourseContentRepo{
		AddSectionT: func(input domain.AddSectionInput, uid string) (string, error) {
			assert.Equal(t, 1, input.CourseContentID)
			assert.Equal(t, 1, input.ContentTypeID)
			assert.Equal(t, "New Section", input.Title)
			assert.Equal(t, "Section desc", input.Description)
			assert.Equal(t, userID, uid)
			return "content-123", nil
		},
	}

	controller := controller.CourseContentController{Repo: mockContentRepo}
	handler := controller.AddSection()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Sección agregada","content_id":"content-123","course_content_id":1,"content_type_id":1}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestCourseContentController_UpdateContentStatus_Success(t *testing.T) {
	inputJSON := `{"content_id":"content-123","is_active":true}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/status", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockContentRepo := MockCourseContentRepo{
		UpdateContentStatusT: func(contentID string, isActive bool) error {
			assert.Equal(t, "content-123", contentID)
			assert.True(t, isActive)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockContentRepo}
	handler := controller.UpdateContentStatus()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Estado del contenido actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestCourseContentController_UpdateContentStatus_InvalidInput(t *testing.T) {
	inputJSON := `{"content_id":""}` // Missing content_id

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/content/status", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}
	e.HTTPErrorHandler = testHTTPErrorHandler

	mockContentRepo := MockCourseContentRepo{}

	controller := controller.CourseContentController{Repo: mockContentRepo}
	handler := controller.UpdateContentStatus()

	err := handler(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	}
}

func TestCourseContentController_UpdateModuleTitle_Success(t *testing.T) {
	inputJSON := `{"course_content_id":1,"module_title":"Updated Module"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/module", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockContentRepo := MockCourseContentRepo{
		UpdateModuleTitleT: func(courseContentID int, moduleTitle string) error {
			assert.Equal(t, 1, courseContentID)
			assert.Equal(t, "Updated Module", moduleTitle)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockContentRepo}
	handler := controller.UpdateModuleTitle()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := `{"Body":{"message":"Título del módulo actualizado"}}`
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

// Mock GeneratePresignedURL
func mockGeneratePresignedURL(bucket, key string) (string, error) {
	return "https://mock-signed-url.com/" + key, nil
}

// Tests
func TestCourseContentController_GetCourseContentTeacher_Success(t *testing.T) {
	userID := "teacher-123"
	courseID := 1
	createdAt := time.Date(2025, 5, 17, 19, 25, 54, 894897000, time.FixedZone("-05:00", -5*60*60)) // Fixed timestamp

	mockContent := []domain.CourseContentWithDetails{
		{
			CourseContentDB: domain.CourseContentDB{
				CourseContentID: 1,
				CourseID:        courseID,
				Module:          "Module 1",
				ModuleIndex:     1,
				CreatedAt:       createdAt,
			},
			Details: []domain.Content{
				{
					ContentID:       "content-1",
					CourseContentID: 1,
					ContentTypeID:   1, // Video
					Title:           "Video 1",
					Url:             "http://original-video.com",
					Description:     "Video desc",
					SectionIndex:    1,
					IsActive:        true,
					UserContent:     nil, // Match actual nil value
				},
				{
					ContentID:       "content-2",
					CourseContentID: 1,
					ContentTypeID:   2, // Quiz
					Title:           "Quiz 1",
					Description:     "Quiz desc",
					SectionIndex:    2,
					IsActive:        true,
					UserContent:     nil, // Match actual nil value
					Url:             "https://mock-signed-url.com/focused/1/text/teacher/content-2.json",
				},
			},
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content?course_id=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)

	mockCourseRepo := MockCourseRepo{
		GetCourseByTeacherAndCourseIDT: func(teacherID string, cid int) (domain.CourseDB, error) {
			assert.Equal(t, userID, teacherID)
			assert.Equal(t, courseID, cid)
			return domain.CourseDB{CourseID: courseID, TeacherID: userID}, nil
		},
	}

	mockContentRepo := MockCourseContentRepo{
		GetContentByCourseT: func(cid int) ([]domain.CourseContentWithDetails, error) {
			assert.Equal(t, courseID, cid)
			return mockContent, nil
		},
	}

	controller := controller.CourseContentController{
		Repo:                 mockContentRepo,
		RepoCourse:           mockCourseRepo,
		GeneratePresignedURL: mockGeneratePresignedURL, // Inject mock
	}
	handler := controller.GetCourseContentTeacher()

	err := handler(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON := fmt.Sprintf(`[
			{
				"contents": null,
				"course_content_id": 1,
				"course_id": 1,
				"created_at": "%s",
				"module": "Module 1",
				"module_index": 1,
				"details": [
					{
						"UserContent": null,
						"content_id": "content-1",
						"course_content_id": 1,
						"content_type_id": 1,
						"title": "Video 1",
						"url": "http://original-video.com",
						"description": "Video desc",
						"section_index": 1,
						"is_active": true
					},
					{
						"UserContent": null,
						"content_id": "content-2",
						"course_content_id": 1,
						"content_type_id": 2,
						"title": "Quiz 1",
						"url": "https://mock-signed-url.com/focused/1/text/teacher/content-2.json",
						"description": "Quiz desc",
						"section_index": 2,
						"is_active": true
					}
				]
			}
		]`, createdAt.Format(time.RFC3339Nano))
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}
