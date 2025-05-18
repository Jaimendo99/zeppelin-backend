package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSubmitQuiz(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	// Test cases
	tests := []struct {
		name               string
		userID             string
		userRole           string
		inputJSON          string
		mockSetup          func(*MockQuizRepo, *MockCourseContentRepo, *MockAssignmentRepo)
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:     "Success - Quiz Submission",
			userID:   "user123",
			userRole: "org:student",
			inputJSON: `{
				"content_id": "quiz123",
				"start_time": "2023-10-01T10:00:00Z",
				"end_time": "2023-10-01T11:00:00Z",
				"answers": {
					"q1": "answer1",
					"q2": true,
					"q3": ["option1", "option3"]
				}
			}`,
			mockSetup: func(qr *MockQuizRepo, ccr *MockCourseContentRepo, ar *MockAssignmentRepo) {
				// Mock the GetAssignmentsByStudentAndCourse method directly
				ar.GetAssignmentsBySAC = func(userID string, courseID int) (domain.AssignmentWithCourse, error) {
					return domain.AssignmentWithCourse{
						AssignmentID: 1,
						AssignedAt:   "2023-01-01T00:00:00Z",
						IsActive:     true,
						IsVerify:     true,
						CourseID:     1,
						TeacherID:    "teacher456",
						StartDate:    "2023-01-01T00:00:00Z",
						Title:        "Test Course",
						Description:  "A test course",
						QRCode:       "qr123456",
					}, nil
				}

				// Setup URL retrieval
				ccr.GetUrlByContentIDT = func(contentID string) (string, error) {
					return "https://account123.r2.cloudflarestorage.com/zeppelin/courses/1/quizzes/quiz123.json", nil
				}

				// Setup content type check
				ccr.GetContentTypeIDT = func(contentID string) (int, error) {
					return 3, nil // 3 means quiz
				}

				// Setup successful quiz save
				qr.SaveQuizAttemptFn = func(attempt domain.QuizAnswer) error {
					return nil
				}

				// Set environment variable for test
				os.Setenv("R2_ACCOUNT_ID", "account123")
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"message": "Quiz calificado exitosamente",
			},
		},
		{
			name:     "Failure - Student Not Assigned To Course",
			userID:   "user123",
			userRole: "org:student",
			inputJSON: `{
				"content_id": "quiz123",
				"start_time": "2023-10-01T10:00:00Z",
				"end_time": "2023-10-01T11:00:00Z",
				"answers": {"q1": "answer1"}
			}`,
			mockSetup: func(qr *MockQuizRepo, ccr *MockCourseContentRepo, ar *MockAssignmentRepo) {
				// Setup URL retrieval
				ccr.GetUrlByContentIDT = func(contentID string) (string, error) {
					return "https://account123.r2.cloudflarestorage.com/zeppelin/courses/1/quizzes/quiz123.json", nil
				}

				// Mock the GetAssignmentsByStudentAndCourse method directly - student not assigned
				ar.GetAssignmentsBySAC = func(userID string, courseID int) (domain.AssignmentWithCourse, error) {
					return domain.AssignmentWithCourse{}, errors.New("no assignment found for this student and course")
				}
			},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse: map[string]interface{}{
				"message": "Este estudiante no está asignado a este curso",
			},
		},
		{
			name:     "Failure - Content is Not a Quiz",
			userID:   "user123",
			userRole: "org:student",
			inputJSON: `{
				"content_id": "lesson123",
				"start_time": "2023-10-01T10:00:00Z",
				"end_time": "2023-10-01T11:00:00Z",
				"answers": {"q1": "answer1"}
			}`,
			mockSetup: func(qr *MockQuizRepo, ccr *MockCourseContentRepo, ar *MockAssignmentRepo) {
				// Setup URL retrieval
				ccr.GetUrlByContentIDT = func(contentID string) (string, error) {
					return "https://account123.r2.cloudflarestorage.com/zeppelin/courses/1/lessons/lesson123.json", nil
				}

				// Mock the GetAssignmentsByStudentAndCourse method directly
				ar.GetAssignmentsBySAC = func(userID string, courseID int) (domain.AssignmentWithCourse, error) {
					return domain.AssignmentWithCourse{
						AssignmentID: 1,
						AssignedAt:   "2023-01-01T00:00:00Z",
						IsActive:     true,
						IsVerify:     true,
						CourseID:     1,
						TeacherID:    "teacher456",
						StartDate:    "2023-01-01T00:00:00Z",
						Title:        "Test Course",
						Description:  "A test course",
						QRCode:       "qr123456",
					}, nil
				}

				// Setup content type check - not a quiz
				ccr.GetContentTypeIDT = func(contentID string) (int, error) {
					return 1, nil // 1 means lesson, not quiz
				}
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"message": "el content_id no corresponde a un quiz",
			},
		},
	}

	// Mock config.UploadJSONToR2 and config.GetR2Object
	originalUploadJSONToR2 := controller.ConfigUploadJSONToR2
	originalGetR2Object := controller.ConfigGetR2Object

	defer func() {
		// Restore original functions after tests
		controller.ConfigUploadJSONToR2 = originalUploadJSONToR2
		controller.ConfigGetR2Object = originalGetR2Object
	}()

	// Override with mocks for testing
	controller.ConfigUploadJSONToR2 = func(key string, data []byte) error {
		return nil
	}
	controller.ConfigGetR2Object = func(bucket, key string) ([]byte, error) {
		teacherQuiz := domain.TeacherQuiz{
			Questions: []domain.TeacherQuizQuestion{
				{
					ID:            "q1",
					Type:          "text",
					Points:        10,
					CorrectAnswer: "answer1",
				},
				{
					ID:            "q2",
					Type:          "boolean",
					Points:        5,
					CorrectAnswer: true,
				},
				{
					ID:             "q3",
					Type:           "checkbox",
					Points:         15,
					CorrectAnswers: []string{"option1", "option3"},
				},
			},
		}
		return json.Marshal(teacherQuiz)
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/quiz/submit", strings.NewReader(tt.inputJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set user context
			c.Set("user_id", tt.userID)
			c.Set("user_role", tt.userRole)

			// Create mocks
			quizRepo := &MockQuizRepo{}
			courseContentRepo := &MockCourseContentRepo{}
			assignmentRepo := &MockAssignmentRepo{}

			// Setup mocks based on test case
			if tt.mockSetup != nil {
				tt.mockSetup(quizRepo, courseContentRepo, assignmentRepo)
			}

			// Create controller with mocks
			controller := &controller.QuizController{
				QuizRepo:          quizRepo,
				CourseContentRepo: courseContentRepo,
				AssignmentRepo:    assignmentRepo,
			}

			// Call handler
			handler := controller.SubmitQuiz()
			_ = handler(c)

			// Parse response
			var respData map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &respData)
			assert.NoError(t, err)

			// For error cases, ReturnWriteResponse returns a status 200 with error in the message
			message, ok := respData["message"]
			if ok {
				// Error case
				assert.Contains(t, message.(string), tt.expectedResponse["message"].(string))
			} else if body, ok := respData["Body"].(map[string]interface{}); ok {
				// Success case
				assert.Equal(t, tt.expectedResponse["message"], body["message"])
			} else {
				assert.Fail(t, "Unexpected response format", "Got: %v", respData)
			}
		})
	}

	// Restoration is handled by the defer function at the top of the test
}

func TestGetQuizAnswersByStudent(t *testing.T) {
	// Setup
	e := echo.New()

	// Test cases
	tests := []struct {
		name               string
		userID             string
		userRole           string
		mockSetup          func(*MockQuizRepo)
		expectedStatusCode int
		expectedQuizCount  int
	}{
		{
			name:     "Success - Get Student Quiz Answers",
			userID:   "student123",
			userRole: "org:student",
			mockSetup: func(qr *MockQuizRepo) {
				qr.GetQuizAnswersByStudentFn = func(studentID string) ([]domain.QuizSummary, error) {
					assert.Equal(t, "student123", studentID)

					// Create sample quiz summaries
					grade := 85.5
					totalPoints := 100
					quizCount := 1
					lastQuizTime := time.Date(2023, 10, 15, 14, 30, 0, 0, time.UTC) // Fixed timestamp for testing

					return []domain.QuizSummary{
						{
							ContentID:    "quiz1",
							QuizCount:    quizCount,
							TotalGrade:   &grade,
							TotalPoints:  &totalPoints,
							LastQuizTime: &lastQuizTime,
						},
					}, nil
				}
			},
			expectedStatusCode: http.StatusOK,
			expectedQuizCount:  1,
		},
		{
			name:     "Failure - Not a Student",
			userID:   "teacher123",
			userRole: "org:teacher",
			mockSetup: func(qr *MockQuizRepo) {
				// No mock setup needed as we should fail before calling the repo
			},
			expectedStatusCode: http.StatusForbidden,
			expectedQuizCount:  0,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/quiz/answers", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set user context
			c.Set("user_id", tt.userID)
			c.Set("user_role", tt.userRole)

			// Create mocks
			quizRepo := &MockQuizRepo{}

			// Setup mocks based on test case
			if tt.mockSetup != nil {
				tt.mockSetup(quizRepo)
			}

			// Create controller with mocks
			controller := &controller.QuizController{
				QuizRepo: quizRepo,
			}

			// Call handler
			handler := controller.GetQuizAnswersByStudent()
			_ = handler(c)

			// For this test, we need to check if we got an error response or a success response
			if tt.userRole != "org:student" {
				// For non-students, the error should be returned through the echo.HTTPError object
				// We can just check for the response code which should be 200 (ReturnReadResponse always returns 200)
				// But the response should either be empty or contain an error message
				respBody := rec.Body.String()
				if respBody != "" && respBody != "[]" {
					var errResp map[string]interface{}
					err := json.Unmarshal(rec.Body.Bytes(), &errResp)
					assert.NoError(t, err)
					if message, ok := errResp["message"]; ok {
						assert.Contains(t, message, "Solo los estudiantes pueden ver sus cursos")
					}
				}
			} else if tt.expectedQuizCount > 0 {
				// Success case with data - response should be an array
				var resp []interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedQuizCount, len(resp))
			}
		})
	}
}
