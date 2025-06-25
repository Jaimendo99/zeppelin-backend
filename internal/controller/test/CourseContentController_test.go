package controller_test

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

// Mocks
type MockCourseContentRepo struct {
	AddModuleT                    func(courseID int, module string, userID string) (int, error)
	GetContentByCourseT           func(courseID int) ([]domain.CourseContentWithDetails, error)
	GetContentByCourseForStudentT func(courseID int, userID string) ([]domain.CourseContentWithStudentDetails, error)
	AddSectionT                   func(input domain.AddSectionInput, userID string) (string, error)
	UpdateContentT                func(input domain.UpdateContentInput) error
	UpdateContentStatusT          func(contentID string, isActive bool) error
	UpdateModuleTitleT            func(courseContentID int, moduleTitle string) error
	UpdateUserContentStatusT      func(userID, contentID string, statusID int) error
	GetContentTypeIDT             func(contentID string) (int, error)
}

func (m MockCourseContentRepo) AddModule(courseID int, module string, userID string) (int, error) {
	if m.AddModuleT != nil {
		return m.AddModuleT(courseID, module, userID)
	}
	return 0, errors.New("AddModule not implemented")
}

func (m MockCourseContentRepo) GetContentByCourse(courseID int) ([]domain.CourseContentWithDetails, error) {
	if m.GetContentByCourseT != nil {
		return m.GetContentByCourseT(courseID)
	}
	return nil, errors.New("GetContentByCourse not implemented")
}

func (m MockCourseContentRepo) GetContentByCourseForStudent(courseID int, userID string) ([]domain.CourseContentWithStudentDetails, error) {
	if m.GetContentByCourseForStudentT != nil {
		return m.GetContentByCourseForStudentT(courseID, userID)
	}
	return nil, errors.New("GetContentByCourseForStudent not implemented")
}

func (m MockCourseContentRepo) AddSection(input domain.AddSectionInput, userID string) (string, error) {
	if m.AddSectionT != nil {
		return m.AddSectionT(input, userID)
	}
	return "", errors.New("AddSection not implemented")
}

func (m MockCourseContentRepo) UpdateContent(input domain.UpdateContentInput) error {
	if m.UpdateContentT != nil {
		return m.UpdateContentT(input)
	}
	return errors.New("UpdateContent not implemented")
}

func (m MockCourseContentRepo) UpdateContentStatus(contentID string, isActive bool) error {
	if m.UpdateContentStatusT != nil {
		return m.UpdateContentStatusT(contentID, isActive)
	}
	return errors.New("UpdateContentStatus not implemented")
}

func (m MockCourseContentRepo) UpdateModuleTitle(courseContentID int, moduleTitle string) error {
	if m.UpdateModuleTitleT != nil {
		return m.UpdateModuleTitleT(courseContentID, moduleTitle)
	}
	return errors.New("UpdateModuleTitle not implemented")
}

func (m MockCourseContentRepo) UpdateUserContentStatus(userID, contentID string, statusID int) error {
	if m.UpdateUserContentStatusT != nil {
		return m.UpdateUserContentStatusT(userID, contentID, statusID)
	}
	return errors.New("UpdateUserContentStatus not implemented")
}

func (m MockCourseContentRepo) GetContentTypeID(contentID string) (int, error) {
	if m.GetContentTypeIDT != nil {
		return m.GetContentTypeIDT(contentID)
	}
	return 0, errors.New("GetContentTypeID not implemented")
}

func (m MockCourseContentRepo) VerifyModuleOwnership(courseContentID int, userID string) error {
	return errors.New("VerifyModuleOwnership not implemented")
}

func (m MockCourseContentRepo) CreateContent(input domain.AddSectionInput) (string, error) {
	return "", errors.New("CreateContent not implemented")
}

func (m MockCourseContentRepo) GetUrlByContentID(contentID string) (string, error) {
	return "", errors.New("GetUrlByContentID not implemented")
}

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
	createdAt := time.Now().Format(time.RFC3339Nano) // Match actual format
	mockContent := []domain.CourseContentWithDetails{
		{
			CourseContentDB: domain.CourseContentDB{
				CourseContentID: 1,
				CourseID:        courseID,
				Module:          "Module 1",
				ModuleIndex:     1,
				CreatedAt:       time.Now(),
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
		// Include all fields from actual response
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
		]`, createdAt)
		assert.JSONEq(t, expectedJSON, rec.Body.String())
	}
}

func TestCourseContentController_GetCourseContentTeacher_InvalidCourseID(t *testing.T) {
	userID := "teacher-123"

	// Test case 1: Missing course_id query parameter
	t.Run("Missing CourseID", func(t *testing.T) {
		// Assuming 'e', 'c', and 'rec' are created here by your test setup
		e := echo.New() // Replace with your actual setup call if needed
		req := httptest.NewRequest(http.MethodGet, "/content", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.Validator = &CustomValidator{Validator: validator.New()} // Assuming these are available
		e.HTTPErrorHandler = testHTTPErrorHandler                  // Assuming this is available

		c.Set("user_id", userID) // Set user ID

		// Create the controller with minimal mocks (they won't be called in this error path)
		controller := controller.CourseContentController{
			Repo:       MockCourseContentRepo{}, // Use zero value or mock
			RepoCourse: MockCourseRepo{},
			// GeneratePresignedURL not strictly needed for this error path
		}
		handler := controller.GetCourseContentTeacher()

		err := handler(c)

		// Assert that an HTTPError is returned
		require.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)

		// Assert the status code
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		// *** Corrected Assertion for the error message ***
		// The error message is a struct with a 'Message' field
		// We need to assert that the Message field of the HTTPError's Message (which is a struct)
		// contains the expected string.
		errMsgStruct, ok := httpErr.Message.(struct {
			Message interface{} "json:\"message\""
		})
		require.True(t, ok, "Expected HTTPError.Message to be a struct with a Message field")
		assert.Equal(t, "course_id inválido", errMsgStruct.Message)

		// Assert the response recorder status code if your error handler sets it
		// assert.Equal(t, http.StatusBadRequest, rec.Code) // Uncomment if your error handler sets rec.Code
	})

	// Test case 2: Non-integer course_id query parameter
	t.Run("Non-Integer CourseID", func(t *testing.T) {
		// Assuming 'e', 'c', and 'rec' are created here by your test setup
		e := echo.New() // Replace with your actual setup call if needed
		req := httptest.NewRequest(http.MethodGet, "/content?course_id=abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.Validator = &CustomValidator{Validator: validator.New()} // Assuming these are available
		e.HTTPErrorHandler = testHTTPErrorHandler                  // Assuming this is available

		c.Set("user_id", userID) // Set user ID

		// Create the controller with minimal mocks
		controller := controller.CourseContentController{
			Repo:       MockCourseContentRepo{},
			RepoCourse: MockCourseRepo{},
		}
		handler := controller.GetCourseContentTeacher()

		err := handler(c)

		// Assert that an HTTPError is returned
		require.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)

		// Assert the status code
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)

		// *** Corrected Assertion for the error message ***
		// The error message is a struct with a 'Message' field
		// We need to assert that the Message field of the HTTPError's Message (which is a struct)
		// contains the expected string.
		errMsgStruct, ok := httpErr.Message.(struct {
			Message interface{} "json:\"message\""
		})
		require.True(t, ok, "Expected HTTPError.Message to be a struct with a Message field")
		assert.Equal(t, "course_id inválido", errMsgStruct.Message)

		// Assert the response recorder status code if your error handler sets it
		// assert.Equal(t, http.StatusBadRequest, rec.Code) // Uncomment if your error handler sets rec.Code
	})

	// Add more cases if other invalid formats are possible
}

func mockGeneratePresignedURLWithError(bucket, key string) (string, error) {
	// Simulate an error when a specific key is requested
	if key == "focused/1/text/teacher/content-2.json" { // Example key that will fail
		return "", errors.New("mock R2 presigned URL error")
	}
	// Default success for other keys
	return "https://mock-signed-url.com/" + key, nil
}

func TestCourseContentController_GetCourseContentTeacher_Forbidden(t *testing.T) {
	userID := "intruder-123" // A user who is NOT the teacher
	courseID := 1

	// Assuming 'e', 'c', and 'rec' are created here by your test setup
	e := echo.New() // Replace with your actual setup call if needed
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/content?course_id=%d", courseID), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}
	e.HTTPErrorHandler = testHTTPErrorHandler

	c.Set("user_id", userID) // Set the intruder's user ID

	// Configure MockCourseRepo to indicate the course does NOT belong to the intruder
	mockCourseRepo := MockCourseRepo{
		GetCourseByTeacherAndCourseIDT: func(inputTeacherID string, cid int) (domain.CourseDB, error) {
			assert.Equal(t, userID, inputTeacherID) // The controller will pass the intruder's ID
			assert.Equal(t, courseID, cid)
			// Simulate that the course is not found for this teacher, indicating forbidden access
			return domain.CourseDB{}, errors.New("course does not belong to the teacher") // Match the expected error string
		},
	}

	// MockContentRepo and GeneratePresignedURL won't be called in this error path
	mockContentRepo := MockCourseContentRepo{}

	controller := controller.CourseContentController{
		Repo:       mockContentRepo,
		RepoCourse: mockCourseRepo,
		// GeneratePresignedURL doesn't matter for this test
	}
	handler := controller.GetCourseContentTeacher()

	err := handler(c)

	// Assert that an HTTPError is returned due to forbidden access
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)

	// Assert the status code
	assert.Equal(t, http.StatusForbidden, httpErr.Code)

	// Assert the error message structure
	errMsgStruct, ok := httpErr.Message.(struct {
		Message interface{} "json:\"message\""
	})
	require.True(t, ok, "Expected HTTPError.Message to be a struct with a Message field")
	assert.Equal(t, "Este curso no le pertenece al profesor", errMsgStruct.Message) // Match the exact error message from the controller

}
func TestCourseContentController_GetCourseContentTeacher_GeneratePresignedURLError(t *testing.T) {
	userID := "teacher-123"
	courseID := 1
	// Define mock content that will trigger a GeneratePresignedURL call
	mockContent := []domain.CourseContentWithDetails{
		{
			CourseContentDB: domain.CourseContentDB{
				CourseContentID: 1,
				CourseID:        courseID,
				Module:          "Module 1",
				ModuleIndex:     1,
				CreatedAt:       time.Now(),
			},
			Details: []domain.Content{
				{
					ContentID:       "content-1",
					CourseContentID: 1,
					ContentTypeID:   1, // Video (should not need presigned URL)
					Title:           "Video 1",
					Url:             "http://original-video.com",
					Description:     "Video desc",
					SectionIndex:    1,
					IsActive:        true,
				},
				{
					ContentID:       "content-2",
					CourseContentID: 1,
					ContentTypeID:   2, // Quiz (should need presigned URL) - This one will error
					Title:           "Quiz 1",
					Description:     "Quiz desc",
					SectionIndex:    2,
					IsActive:        true,
				},
			},
		},
	}

	// Assuming 'e', 'c', and 'rec' are created here by your test setup
	e := echo.New() // Replace with your actual setup call if needed
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/content?course_id=%d", courseID), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()} // Assuming these are available
	e.HTTPErrorHandler = testHTTPErrorHandler                  // Assuming this is available

	c.Set("user_id", userID) // Set user ID

	// Configure mock repositories to return data that leads to the error
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
		// Other methods of MockCourseContentRepo are not expected to be called in this specific test
	}

	controller := controller.CourseContentController{
		Repo:                 mockContentRepo,
		RepoCourse:           mockCourseRepo,
		GeneratePresignedURL: mockGeneratePresignedURLWithError, // Inject the mock that returns an error
	}
	handler := controller.GetCourseContentTeacher()

	err := handler(c)

	// Assert that an HTTPError is returned due to the GeneratePresignedURL error
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)

	// Assert the status code
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)

	// Assert the error message structure returned by ReturnReadResponse
	errMsgStruct, ok := httpErr.Message.(struct {
		Message interface{} "json:\"message\""
	})
	require.True(t, ok, "Expected HTTPError.Message to be a struct with a Message field")

	// Check that the error message contains the expected formatted string
	expectedErrMsg := fmt.Sprintf("error al generar URL firmada para content_type_id %d", 2) // ContentTypeID that failed
	assert.Equal(t, expectedErrMsg, errMsgStruct.Message)

}

func TestCourseContentController_UpdateContent_Success_Text(t *testing.T) {
	inputJSON := `{
		"course_id": 1,
		"content_id": "content-123",
		"json_data": {"key":"value"}
	}`

	os.Setenv("R2_ACCOUNT_ID", "test-account") // Simula variable de entorno

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/course-content", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		GetContentTypeIDT: func(contentID string) (int, error) {
			assert.Equal(t, "content-123", contentID)
			return 2, nil // Texto
		},
		UpdateContentT: func(input domain.UpdateContentInput) error {
			assert.Equal(t, "https://test-account.r2.cloudflarestorage.com/focused/1/text/teacher/content-123.json", input.Url)
			return nil
		},
	}

	mockUpload := func(courseID, contentID string, json []byte) error {
		assert.Equal(t, "1", courseID)
		assert.Equal(t, "content-123", contentID)
		assert.JSONEq(t, `{"key":"value"}`, string(json))
		return nil
	}

	controller := controller.CourseContentController{
		Repo:           mockRepo,
		UploadTextFunc: mockUpload,
	}

	handler := controller.UpdateContent()

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	expected := `{"Body":{"message":"Contenido actualizado"}}`
	assert.JSONEq(t, expected, rec.Body.String())
}

func TestCourseContentController_UpdateContent_Success_Quiz(t *testing.T) {
	inputJSON := `{
		"course_id": 1,
		"content_id": "content-123",
		"json_data": {"key":"value"}
	}`

	os.Setenv("R2_ACCOUNT_ID", "test-account") // Simula variable de entorno

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/course-content", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		GetContentTypeIDT: func(contentID string) (int, error) {
			assert.Equal(t, "content-123", contentID)
			return 3, nil
		},
		UpdateContentT: func(input domain.UpdateContentInput) error {
			assert.Equal(t, "https://test-account.r2.cloudflarestorage.com/focused/1/quiz/teacher/content-123.json", input.Url)
			return nil
		},
	}

	mockUploadQuiz := func(courseID, contentID string, json []byte) error {
		assert.Equal(t, "1", courseID)
		assert.Equal(t, "content-123", contentID)
		assert.JSONEq(t, `{"key":"value"}`, string(json))
		return nil
	}

	controller := controller.CourseContentController{
		Repo:           mockRepo,
		UploadQuizFunc: mockUploadQuiz,
	}

	handler := controller.UpdateContent()

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	expected := `{"Body":{"message":"Contenido actualizado"}}`
	assert.JSONEq(t, expected, rec.Body.String())
}

func TestCourseContentController_UpdateContent_Success_Video(t *testing.T) {
	inputJSON := `{
		"course_id": 1,
		"content_id": "video-789",
		"video_id": "https://video.url/123"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/course-content", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		GetContentTypeIDT: func(contentID string) (int, error) {
			return 1, nil // Video
		},
		UpdateContentT: func(input domain.UpdateContentInput) error {
			assert.Equal(t, "https://video.url/123", input.Url)
			return nil
		},
	}

	controller := controller.CourseContentController{
		Repo: mockRepo,
	}

	handler := controller.UpdateContent()

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Contenido actualizado")
}

type MockAssignmentRepoLocal struct {
	GetAssignmentsByStudentAndCourseFn func(userID string, courseID int) (domain.AssignmentWithCourse, error)
}

func (m MockAssignmentRepoLocal) CreateAssignment(userID string, courseID int) error {
	//TODO implement me
	panic("implement me")
}

func (m MockAssignmentRepoLocal) VerifyAssignment(assignmentID int) error {
	//TODO implement me
	panic("implement me")
}

func (m MockAssignmentRepoLocal) GetAssignmentsByStudent(userID string) ([]domain.StudentCourseProgress, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAssignmentRepoLocal) GetStudentsByCourse(courseID int) ([]domain.AssignmentWithStudent, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAssignmentRepoLocal) GetCourseIDByQRCode(qrCode string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAssignmentRepoLocal) GetAssignmentsByStudentAndCourse(userID string, courseID int) (domain.AssignmentWithCourse, error) {
	if m.GetAssignmentsByStudentAndCourseFn != nil {
		return m.GetAssignmentsByStudentAndCourseFn(userID, courseID)
	}
	return domain.AssignmentWithCourse{}, errors.New("not implemented")
}

func TestCourseContentController_GetCourseContentForStudent_Success(t *testing.T) {
	userID := "student-123"
	courseID := 1
	role := "org:student"

	mockContent := []domain.CourseContentWithStudentDetails{
		{
			CourseContentID: 1,
			CourseID:        courseID,
			Module:          "Módulo 1",
			ModuleIndex:     0,
			CreatedAt:       time.Now(),
			Details: []domain.ContentWithStatus{
				{
					ContentID:       "content-1",
					CourseContentID: 1,
					ContentTypeID:   2, // Quiz → debería generar URL con /text/teacher/
					Title:           "Quiz 1",
					Description:     "Desc",
					SectionIndex:    1,
					IsActive:        true,
				},
				{
					ContentID:       "content-2",
					CourseContentID: 1,
					ContentTypeID:   3, // Text → debería generar URL con /quiz/student/
					Title:           "Texto 1",
					Description:     "Desc",
					SectionIndex:    2,
					IsActive:        true,
				},
			},
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/course-content/student?course_id=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	c.Set("user_role", role)
	os.Setenv("BOURBON", "0") // para que solo 'org:student' tenga acceso

	mockRepo := MockCourseContentRepo{
		GetContentByCourseForStudentT: func(cid int, uid string) ([]domain.CourseContentWithStudentDetails, error) {
			assert.Equal(t, courseID, cid)
			assert.Equal(t, userID, uid)
			return mockContent, nil
		},
	}

	mockAssignmentRepo := MockAssignmentRepoLocal{
		GetAssignmentsByStudentAndCourseFn: func(uid string, cid int) (domain.AssignmentWithCourse, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, courseID, cid)
			return domain.AssignmentWithCourse{}, nil // simulamos asignación válida
		},
	}

	mockPresigned := func(bucket, key string) (string, error) {
		return "https://signed-url/" + key, nil
	}

	controller := controller.CourseContentController{
		Repo:                 mockRepo,
		RepoAssigment:        mockAssignmentRepo,
		GeneratePresignedURL: mockPresigned,
	}

	handler := controller.GetCourseContentForStudent()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Validamos que el cuerpo tenga las URLs firmadas correctamente
	assert.Contains(t, rec.Body.String(), "https://signed-url/focused/1/text/teacher/content-1.json")
	assert.Contains(t, rec.Body.String(), "https://signed-url/focused/1/quiz/student/content-2.json")
}

func TestCourseContentController_UpdateUserContentStatus_Success(t *testing.T) {
	userID := "student-123"
	inputJSON := `{"content_id":"content-abc"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/user-content-status", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateUserContentStatusT: func(uid, contentID string, statusID int) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, "content-abc", contentID)
			assert.Equal(t, 3, statusID)
			return nil
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateUserContentStatus(3) // statusID = 3 (completado)

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	expected := `{"Body":{"message":"Contenido marcado como 'completado'"}}`
	assert.JSONEq(t, expected, rec.Body.String())
}

func TestCourseContentController_UpdateUserContentStatus_InvalidInput(t *testing.T) {
	inputJSON := `{}` // Falta content_id

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/user-content-status", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "student-123")
	e.Validator = &CustomValidator{Validator: validator.New()}
	e.HTTPErrorHandler = testHTTPErrorHandler // para capturar errores personalizados

	mockRepo := MockCourseContentRepo{}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateUserContentStatus(2) // estado 'en progreso'

	err := handler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestCourseContentController_UpdateUserContentStatus_RepoError(t *testing.T) {
	inputJSON := `{"content_id":"content-err"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/user-content-status", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "student-123")
	e.Validator = &CustomValidator{Validator: validator.New()}

	mockRepo := MockCourseContentRepo{
		UpdateUserContentStatusT: func(uid, cid string, statusID int) error {
			return errors.New("falló el repo")
		},
	}

	controller := controller.CourseContentController{Repo: mockRepo}
	handler := controller.UpdateUserContentStatus(3) // estado 'completado'

	err := handler(c)

	require.NoError(t, err) // porque ReturnWriteResponse maneja el error internamente
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), `"message":"falló el repo"`)
}

func TestCourseContentController_UpdateUserContentStatus_AllCases(t *testing.T) {
	tests := []struct {
		name        string
		statusID    int
		expectedMsg string
	}{
		{"Case 2 - en progreso", 2, "Contenido marcado como 'en progreso'"},
		{"Case 3 - completado", 3, "Contenido marcado como 'completado'"},
		{"Default case", 9, "Estado actualizado"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputJSON := `{"content_id":"content-123"}`
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/user-content-status", strings.NewReader(inputJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "student-123")
			e.Validator = &CustomValidator{Validator: validator.New()}

			mockRepo := MockCourseContentRepo{
				UpdateUserContentStatusT: func(uid, cid string, statusID int) error {
					assert.Equal(t, "student-123", uid)
					assert.Equal(t, "content-123", cid)
					assert.Equal(t, tt.statusID, statusID)
					return nil
				},
			}

			controller := controller.CourseContentController{Repo: mockRepo}
			handler := controller.UpdateUserContentStatus(tt.statusID)

			err := handler(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedMsg)
		})
	}
}
