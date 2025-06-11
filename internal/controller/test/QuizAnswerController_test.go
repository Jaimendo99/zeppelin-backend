package controller_test

import (
	"bytes"
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
	SaveQuizAttemptFn            func(input domain.QuizAnswer) error
	GetQuizAttemptsByStudentFunc func(userID string) ([]domain.QuizAttemptView, error)
	GetQuizAttemptsByCourseFn    func(courseID int) ([]domain.QuizAttemptView, error)
	FindQuizAttemptByIDMock      func(quizAnswerID int) (domain.QuizAnswer, error)
	UpdateQuizAttemptMock        func(attempt domain.QuizAnswer) error
}

func (m mockQuizRepo) FindQuizAttemptByID(quizAnswerID int) (domain.QuizAnswer, error) {
	return m.FindQuizAttemptByIDMock(quizAnswerID)
}

func (m mockQuizRepo) UpdateQuizAttempt(attempt domain.QuizAnswer) error {
	return m.UpdateQuizAttemptMock(attempt)
}

func (m mockQuizRepo) FindQuizAttemptsByCourse(courseID int) ([]domain.QuizAnswer, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockQuizRepo) FindQuizAttemptsByUser(userID string) ([]domain.QuizAnswer, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockQuizRepo) GetQuizAttemptsByCourse(courseID int) ([]domain.QuizAttemptView, error) {
	if m.GetQuizAttemptsByCourseFn != nil {
		return m.GetQuizAttemptsByCourseFn(courseID)
	}
	return nil, nil
}

func (m mockQuizRepo) GetQuizAttemptsByStudent(userID string) ([]domain.QuizAttemptView, error) {
	return m.GetQuizAttemptsByStudentFunc(userID)
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
func (m mockAssignmentRepo) GetAssignmentsByStudent(string) ([]domain.StudentCourseProgress, error) {
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

	mockQuiz := domain.TeacherQuiz{
		Questions: []domain.TeacherQuizQuestion{
			{ID: "q1", Type: "text", Points: 10, CorrectAnswer: "A"},
		},
	}

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			SaveQuizAttemptFn: func(input domain.QuizAnswer) error {
				input.QuizAnswerID = 1
				assert.Equal(t, contentID, input.ContentID)
				assert.Equal(t, userID, input.UserID)
				assert.Equal(t, 0.0, *input.Grade)
				assert.Equal(t, 10, *input.TotalPoints)
				return nil
			},
		},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(uid string, cid int) (domain.AssignmentWithCourse, error) {
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
		UploadStudentAnswers:  mockUploadToR2(t),
		GetTeacherQuizContent: mockGetFromR2(t, mockQuiz),
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"score":0`)
	assert.Contains(t, rec.Body.String(), `"total_points":10`)
	assert.Contains(t, rec.Body.String(), `"quiz_answer_id":`)
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
			"q1": "B",
			"q2": []interface{}{"A", "C"},
			"q3": true,
			"q4": "respuesta abierta",
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
			{ID: "q1", Type: "multiple", Points: 10, CorrectAnswer: "B"},
			{ID: "q2", Type: "checkbox", Points: 10, CorrectAnswers: []string{"A", "C"}},
			{ID: "q3", Type: "boolean", Points: 5, CorrectAnswer: true},
			{ID: "q4", Type: "text", Points: 5, CorrectAnswer: "respuesta abierta"},
		},
	}

	quizCtrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			SaveQuizAttemptFn: func(input domain.QuizAnswer) error {
				input.QuizAnswerID = 1
				assert.Equal(t, contentID, input.ContentID)
				assert.Equal(t, userID, input.UserID)
				assert.Equal(t, 25.0, *input.Grade) // Solo se evalúan 3 preguntas automáticamente
				assert.Equal(t, 30, *input.TotalPoints)
				return nil
			},
		},
		AssignmentRepo: mockAssignmentRepo{
			GetAssignmentsByStudentAndCourseFn: func(uid string, cid int) (domain.AssignmentWithCourse, error) {
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
			return nil
		},
		GetTeacherQuizContent: mockGetFromR2(t, mockQuiz),
	}

	handler := quizCtrl.SubmitQuiz()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"score":25`)
	assert.Contains(t, rec.Body.String(), `"total_points":30`)
	assert.Contains(t, rec.Body.String(), `"quiz_answer_id":`)
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

func floatPointer(f float64) *float64 {
	return &f
}

func intPointer(i int) *int {
	return &i
}

func TestQuizController_GetQuizzesByStudent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/quizzes/student", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "student-123")

	attempts := []domain.QuizAttemptView{
		{
			QuizAnswerID:      1,
			UserID:            "student-123",
			StudentName:       "Juan",
			StudentLastname:   "Pérez",
			StudentEmail:      "juan@example.com",
			Grade:             floatPointer(7.5),
			TotalPoints:       intPointer(10),
			NeedsReview:       false,
			ReviewedAt:        nil,
			QuizAnswerURL:     "https://test-account.r2.cloudflarestorage.com/path/to/answer.json",
			StartTime:         time.Now().Add(-time.Hour),
			EndTime:           time.Now(),
			CourseID:          101,
			CourseTitle:       "Matemáticas",
			CourseDescription: "Curso de matemáticas básicas",
			ContentID:         "quiz-123",
			QuizURL:           "https://test-account.r2.cloudflarestorage.com/path/to/quiz.json",
			QuizTitle:         "Álgebra",
			QuizDescription:   "Evaluación de álgebra",
			CourseContentID:   123,
			Module:            "Módulo 1",
			ModuleIndex:       intPointer(1),
			TotalQuizzes:      3,
		},
	}

	os.Setenv("R2_ACCOUNT_ID", "test-account")

	ctrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			GetQuizAttemptsByStudentFunc: func(userID string) ([]domain.QuizAttemptView, error) {
				return attempts, nil
			},
		},
		GeneratePresignedURL: func(bucket, key string) (string, error) {
			return "https://signed.url/" + key, nil
		},
	}

	handler := ctrl.GetQuizzesByStudent()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"user_id":"student-123"`)
	assert.Contains(t, rec.Body.String(), `"course_id":101`)
	assert.Contains(t, rec.Body.String(), `"quiz_title":"Álgebra"`)
	assert.Contains(t, rec.Body.String(), `"quiz_url":"https://signed.url/path/to/quiz.json"`)
	assert.Contains(t, rec.Body.String(), `"quiz_answer_url":"https://signed.url/path/to/answer.json"`)

}

func mustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(fmt.Sprintf("invalid time %q: %v", value, err))
	}
	return t
}

func TestQuizController_GetQuizzesByCourse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/quizzes/course/101", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("101")

	os.Setenv("R2_ACCOUNT_ID", "test-account")

	attempts := []domain.QuizAttemptView{
		{
			QuizAnswerID:    1,
			UserID:          "student-123",
			StudentName:     "Juan",
			StudentLastname: "Pérez",
			StudentEmail:    "juan@example.com",
			Grade:           floatPointer(7.5),
			TotalPoints:     intPointer(10),
			NeedsReview:     false,
			ReviewedAt:      nil,
			QuizAnswerURL:   "https://test-account.r2.cloudflarestorage.com/path/to/answer.json",
			StartTime:       mustParseTime(time.RFC3339Nano, "2025-06-10T03:02:24.952822Z"),
			EndTime:         mustParseTime(time.RFC3339Nano, "2025-06-10T04:02:24.952822Z"),

			CourseID:          101,
			CourseTitle:       "Matemáticas",
			CourseDescription: "Curso de matemáticas básicas",
			ContentID:         "quiz-123",
			QuizURL:           "https://test-account.r2.cloudflarestorage.com/path/to/quiz.json",
			QuizTitle:         "Álgebra",
			QuizDescription:   "Evaluación de álgebra",
			CourseContentID:   123,
			Module:            "Módulo 1",
			ModuleIndex:       intPointer(1),
			TotalQuizzes:      3,
		},
	}

	ctrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			GetQuizAttemptsByCourseFn: func(courseID int) ([]domain.QuizAttemptView, error) {
				assert.Equal(t, 101, courseID)
				return attempts, nil
			},
		},
		GeneratePresignedURL: func(bucket, key string) (string, error) {
			return "https://signed.url/" + key, nil
		},
	}

	handler := ctrl.GetQuizzesByCourse()
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"course_id":101`)
	assert.Contains(t, rec.Body.String(), `"quiz_title":"Álgebra"`)
	assert.Contains(t, rec.Body.String(), `"quiz_url":"https://signed.url/path/to/quiz.json"`)
	assert.Contains(t, rec.Body.String(), `"quiz_answer_url":"https://signed.url/path/to/answer.json"`)

}

func TestQuizController_ReviewTextAnswer_Success(t *testing.T) {
	e := echo.New()
	e.Validator = &CustomValidator{Validator: validator.New()}

	input := domain.TextAnswerReviewInput{
		QuizAnswerID:  1,
		QuestionID:    "q4",
		PointsAwarded: 5,
		IsCorrect:     true,
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/quiz/review-text", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	os.Setenv("R2_ACCOUNT_ID", "test-account")

	quizAttempt := domain.QuizAnswer{
		QuizAnswerID:  1,
		UserID:        "student-123",
		Grade:         floatPointer(10),
		TotalPoints:   intPointer(0),
		QuizAnswerURL: "https://test-account.r2.cloudflarestorage.com/folder/answer.json",
		QuizURL:       "https://test-account.r2.cloudflarestorage.com/folder/quiz.json",
	}

	studentAnswers := map[string]interface{}{
		"q4": "respuesta abierta",
	}

	teacherQuiz := domain.TeacherQuiz{
		Questions: []domain.TeacherQuizQuestion{
			{ID: "q4", Type: "text", Points: 5},
		},
	}

	ctrl := controller.QuizController{
		QuizRepo: mockQuizRepo{
			FindQuizAttemptByIDMock: func(id int) (domain.QuizAnswer, error) {
				assert.Equal(t, 1, id)
				return quizAttempt, nil
			},
			UpdateQuizAttemptMock: func(attempt domain.QuizAnswer) error {
				assert.Equal(t, 1, attempt.QuizAnswerID)
				assert.Equal(t, 15.0, *attempt.Grade)
				assert.Equal(t, 5, *attempt.TotalPoints)
				assert.NotNil(t, attempt.ReviewedAt)
				return nil
			},
		},
		GetTeacherQuizContent: func(bucket, key string) ([]byte, error) {
			if strings.Contains(key, "answer.json") {
				return json.Marshal(studentAnswers)
			}
			if strings.Contains(key, "quiz.json") {
				return json.Marshal(teacherQuiz)
			}
			return nil, fmt.Errorf("unexpected key %s", key)
		},
		UploadStudentAnswers: func(key string, data []byte) error {
			var m map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &m))
			assert.Contains(t, m, "extra_review")
			return nil
		},
	}

	handler := ctrl.ReviewTextAnswer()
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"score":15`)
	assert.Contains(t, rec.Body.String(), `"total_points":5`)
	assert.Contains(t, rec.Body.String(), `"message":"Respuesta de texto revisada exitosamente"`)
}
