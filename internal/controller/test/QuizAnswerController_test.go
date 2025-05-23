package controller_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/stretchr/testify/assert"
)

// --- Mocks locales ---

type mockQuizRepo struct {
	SaveQuizAttemptFn func(input domain.QuizAnswer) error
}

func (m mockQuizRepo) SaveQuizAttempt(input domain.QuizAnswer) error {
	if m.SaveQuizAttemptFn != nil {
		return m.SaveQuizAttemptFn(input)
	}
	return nil
}

type mockAssignmentRepo struct {
	GetAssignmentsByStudentAndCourseFn func(userID string, courseID int) (domain.AssignmentWithCourse, error)
}

func (m mockAssignmentRepo) GetAssignmentsByStudentAndCourse(userID string, courseID int) (domain.AssignmentWithCourse, error) {
	if m.GetAssignmentsByStudentAndCourseFn != nil {
		return m.GetAssignmentsByStudentAndCourseFn(userID, courseID)
	}
	return domain.AssignmentWithCourse{}, nil
}

// Métodos dummy
func (m mockAssignmentRepo) CreateAssignment(string, int) error { return nil }
func (m mockAssignmentRepo) VerifyAssignment(int) error         { return nil }
func (m mockAssignmentRepo) GetAssignmentsByStudent(string) ([]domain.AssignmentWithCourse, error) {
	return nil, nil
}
func (m mockAssignmentRepo) GetStudentsByCourse(int) ([]domain.AssignmentWithStudent, error) {
	return nil, nil
}
func (m mockAssignmentRepo) GetCourseIDByQRCode(string) (int, error) { return 0, nil }

type mockCourseContentRepo struct {
	GetUrlByContentIDFn func(contentID string) (string, error)
	GetContentTypeIDFn  func(contentID string) (int, error)
}

func (m mockCourseContentRepo) GetUrlByContentID(contentID string) (string, error) {
	if m.GetUrlByContentIDFn != nil {
		return m.GetUrlByContentIDFn(contentID)
	}
	return "", errors.New("not implemented")
}

func (m mockCourseContentRepo) GetContentTypeID(contentID string) (int, error) {
	if m.GetContentTypeIDFn != nil {
		return m.GetContentTypeIDFn(contentID)
	}
	return 0, errors.New("not implemented")
}

// Dummy methods
func (m mockCourseContentRepo) AddModule(int, string, string) (int, error) { return 0, nil }
func (m mockCourseContentRepo) VerifyModuleOwnership(int, string) error    { return nil }
func (m mockCourseContentRepo) GetContentByCourse(int) ([]domain.CourseContentWithDetails, error) {
	return nil, nil
}
func (m mockCourseContentRepo) GetContentByCourseForStudent(int, string) ([]domain.CourseContentWithStudentDetails, error) {
	return nil, nil
}
func (m mockCourseContentRepo) CreateContent(domain.AddSectionInput) (string, error) { return "", nil }
func (m mockCourseContentRepo) AddSection(domain.AddSectionInput, string) (string, error) {
	return "", nil
}
func (m mockCourseContentRepo) UpdateContent(domain.UpdateContentInput) error     { return nil }
func (m mockCourseContentRepo) UpdateContentStatus(string, bool) error            { return nil }
func (m mockCourseContentRepo) UpdateModuleTitle(int, string) error               { return nil }
func (m mockCourseContentRepo) UpdateUserContentStatus(string, string, int) error { return nil }

// --- Funciones de mock para R2 ---

func mockUploadToR2(t *testing.T) func(string, []byte) error {
	return func(key string, data []byte) error {
		assert.Contains(t, key, "quiz/answer/student-123/content-quiz-1.json")
		assert.JSONEq(t, `{"q1":"A"}`, string(data))
		return nil
	}
}

func mockGetFromR2(t *testing.T, quiz domain.TeacherQuiz) func(bucket string, key string) ([]byte, error) {
	return func(bucket, key string) ([]byte, error) {
		assert.Equal(t, "zeppelin", bucket)
		return json.Marshal(quiz)
	}
}

func TestQuizController_SubmitQuiz_Success(t *testing.T) {
	userID := "student-123"
	contentID := "content-quiz-1"
	accountID := "test-account"
	os.Setenv("R2_ACCOUNT_ID", accountID)

	// JSON de entrada simulado
	input := domain.StudentQuizAnswersInput{
		ContentID: contentID,
		StartTime: time.Now().Add(-3 * time.Minute),
		EndTime:   time.Now(),
		Answers: map[string]interface{}{
			"q1": "A",
		},
	}
	body, _ := json.Marshal(input)

	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(http.MethodPost, "/quiz/submit", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)

	// Simulamos un quiz con una pregunta correcta
	mockQuiz := domain.TeacherQuiz{
		Questions: []domain.TeacherQuizQuestion{
			{
				ID:            "q1",
				Type:          "text",
				Points:        10,
				CorrectAnswer: "A",
			},
		},
	}

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			SaveQuizAttemptFn: func(input domain.QuizAnswer) error {
				assert.Equal(t, contentID, input.ContentID)
				assert.Equal(t, userID, input.UserID)
				assert.NotNil(t, input.Grade)
				assert.Equal(t, 10.0, *input.Grade)
				assert.Equal(t, 10, *input.TotalPoints)
				assert.NotEmpty(t, input.QuizAnswerURL)
				return nil
			},
		},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(uid string, cid int) (domain.AssignmentWithCourse, error) {
				assert.Equal(t, userID, uid)
				assert.Equal(t, 123, cid) // extraído de URL en mock
				return domain.AssignmentWithCourse{}, nil
			},
		},
		CourseContentRepo: mockCourseContentRepo{
			GetUrlByContentIDFn: func(id string) (string, error) {
				assert.Equal(t, contentID, id)
				// Simulamos URL de R2 bien formada
				return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/focused/123/quiz/teacher/%s.json", accountID, id), nil
			},
			GetContentTypeIDFn: func(id string) (int, error) {
				assert.Equal(t, contentID, id)
				return 3, nil // tipo quiz
			},
		},
		UploadStudentAnswers:  mockUploadToR2(t),
		GetTeacherQuizContent: mockGetFromR2(t, mockQuiz),
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"score":10`)
	assert.Contains(t, rec.Body.String(), `"total_points":10`)
	assert.Contains(t, rec.Body.String(), `"quiz_answer_id"`)
	assert.Contains(t, rec.Body.String(), `"student_answers_url"`)
}

func TestQuizController_SubmitQuiz_AllTypesSuccess(t *testing.T) {
	userID := "student-123"
	contentID := "content-quiz-1"
	accountID := "test-account"
	os.Setenv("R2_ACCOUNT_ID", accountID)

	input := domain.StudentQuizAnswersInput{
		ContentID: contentID,
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		Answers: map[string]interface{}{
			"q1": "B",                     // multiple
			"q2": []interface{}{"A", "C"}, // checkbox
			"q3": true,                    // boolean
			"q4": "respuesta abierta",     // text
		},
	}
	body, _ := json.Marshal(input)

	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(http.MethodPost, "/quiz/submit", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)

	mockQuiz := domain.TeacherQuiz{
		Questions: []domain.TeacherQuizQuestion{
			{
				ID:            "q1",
				Type:          "multiple",
				Points:        10,
				CorrectAnswer: "B",
			},
			{
				ID:             "q2",
				Type:           "checkbox",
				Points:         10,
				CorrectAnswers: []string{"A", "C"},
			},
			{
				ID:            "q3",
				Type:          "boolean",
				Points:        5,
				CorrectAnswer: true,
			},
			{
				ID:            "q4",
				Type:          "text",
				Points:        5,
				CorrectAnswer: "respuesta abierta",
			},
		},
	}

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			SaveQuizAttemptFn: func(input domain.QuizAnswer) error {
				assert.Equal(t, contentID, input.ContentID)
				assert.Equal(t, userID, input.UserID)
				assert.NotNil(t, input.Grade)
				assert.Equal(t, 30.0, *input.Grade)
				assert.Equal(t, 30, *input.TotalPoints)
				return nil
			},
		},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(uid string, cid int) (domain.AssignmentWithCourse, error) {
				assert.Equal(t, userID, uid)
				assert.Equal(t, 123, cid)
				return domain.AssignmentWithCourse{}, nil
			},
		},
		CourseContentRepo: mockCourseContentRepo{
			GetUrlByContentIDFn: func(id string) (string, error) {
				return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/focused/123/quiz/teacher/%s.json", accountID, id), nil
			},
			GetContentTypeIDFn: func(id string) (int, error) {
				return 3, nil
			},
		},
		UploadStudentAnswers: func(key string, data []byte) error {
			assert.Contains(t, key, userID)
			assert.Contains(t, key, contentID)
			return nil
		},
		GetTeacherQuizContent: mockGetFromR2(t, mockQuiz),
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"score":30`)
	assert.Contains(t, rec.Body.String(), `"total_points":30`)
	assert.Contains(t, rec.Body.String(), `"student_answers_url"`)
}

func TestQuizController_gradeQuiz_BooleanCases(t *testing.T) {
	ctrl := controller.QuizController{}
	questions := []domain.TeacherQuizQuestion{
		{
			ID:            "q1",
			Type:          "boolean",
			Points:        10,
			CorrectAnswer: "verdadero", // string válida
		},
		{
			ID:            "q2",
			Type:          "boolean",
			Points:        5,
			CorrectAnswer: "quizás", // string inválida
		},
		{
			ID:            "q3",
			Type:          "boolean",
			Points:        5,
			CorrectAnswer: true, // correcto, pero respuesta mal parseada
		},
		{
			ID:            "q4",
			Type:          "boolean",
			Points:        3,
			CorrectAnswer: "falso",
		},
	}
	studentAnswers := map[string]interface{}{
		"q1": "Verdadero",  // ✅
		"q2": "no sé",      // ❌
		"q3": "incorrecto", // ❌
		"q4": "falso",      // ✅
	}
	quiz := domain.TeacherQuiz{Questions: questions}

	score, total := ctrl.GradeQuiz(quiz, studentAnswers)

	assert.Equal(t, 13.0, score)
	assert.Equal(t, 23, total)
}

func TestQuizController_SubmitQuiz_Error_GetUrlByContentID(t *testing.T) {
	userID := "student-123"
	contentID := "content-id"

	input := domain.StudentQuizAnswersInput{
		ContentID: contentID,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Answers:   map[string]interface{}{"q1": "a"},
	}
	body, _ := json.Marshal(input)

	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(http.MethodPost, "/quiz/submit", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(userID string, courseID int) (domain.AssignmentWithCourse, error) {
				return domain.AssignmentWithCourse{}, nil
			},
		},
		CourseContentRepo: mockCourseContentRepo{
			GetUrlByContentIDFn: func(id string) (string, error) {
				return "", errors.New("contenido no encontrado") // <- forzamos error aquí
			},
			GetContentTypeIDFn: func(id string) (int, error) {
				return 3, nil
			},
		},
		UploadStudentAnswers:  func(string, []byte) error { return nil },
		GetTeacherQuizContent: func(bucket, key string) ([]byte, error) { return []byte("{}"), nil },
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	// ⚠️ No se espera un error retornado, se espera que la respuesta HTTP tenga código 404
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "contenido con ID")
}

func TestQuizController_SubmitQuiz_Error_StudentNotAssigned(t *testing.T) {
	userID := "student-123"
	contentID := "content-id"

	input := domain.StudentQuizAnswersInput{
		ContentID: contentID,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Answers:   map[string]interface{}{"q1": "a"},
	}
	body, _ := json.Marshal(input)

	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(http.MethodPost, "/quiz/submit", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(userID string, courseID int) (domain.AssignmentWithCourse, error) {
				return domain.AssignmentWithCourse{}, errors.New("no asignado") // fuerza el error
			},
		},
		CourseContentRepo: mockCourseContentRepo{
			GetUrlByContentIDFn: func(id string) (string, error) {
				return "https://test-account.r2.cloudflarestorage.com/focused/123/quiz/teacher/content-id.json", nil
			},
			GetContentTypeIDFn: func(id string) (int, error) {
				return 3, nil
			},
		},
		UploadStudentAnswers:  func(string, []byte) error { return nil },
		GetTeacherQuizContent: func(bucket, key string) ([]byte, error) { return []byte("{}"), nil },
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "no está asignado")
}

func TestQuizController_SubmitQuiz_Error_GetContentTypeID(t *testing.T) {
	userID := "student-123"
	contentID := "content-id"

	input := domain.StudentQuizAnswersInput{
		ContentID: contentID,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Answers:   map[string]interface{}{"q1": "a"},
	}
	body, _ := json.Marshal(input)

	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}
	req := httptest.NewRequest(http.MethodPost, "/quiz/submit", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", userID)

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(userID string, courseID int) (domain.AssignmentWithCourse, error) {
				return domain.AssignmentWithCourse{}, nil
			},
		},
		CourseContentRepo: mockCourseContentRepo{
			GetUrlByContentIDFn: func(id string) (string, error) {
				return "https://test-account.r2.cloudflarestorage.com/focused/123/quiz/teacher/content-id.json", nil
			},
			GetContentTypeIDFn: func(id string) (int, error) {
				return 0, errors.New("fallo al obtener tipo") // provoca error
			},
		},
		UploadStudentAnswers:  func(string, []byte) error { return nil },
		GetTeacherQuizContent: func(bucket, key string) ([]byte, error) { return []byte("{}"), nil },
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "error al obtener tipo de contenido")
}
