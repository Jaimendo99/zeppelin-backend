package domain

import (
	"time"
)

// QuizAnswer model
type QuizAnswer struct {
	QuizAnswerID  int        `gorm:"column:quiz_answer_id;primaryKey;autoIncrement"`
	ContentID     string     `gorm:"column:content_id"`
	UserID        string     `gorm:"column:user_id"`
	StartTime     time.Time  `gorm:"column:start_time"`
	EndTime       time.Time  `gorm:"column:end_time"`
	Grade         *float64   `gorm:"column:grade"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"`
	QuizURL       string     `gorm:"column:quiz_url"`
	QuizAnswerURL string     `gorm:"column:quiz_answer_url"`
	TotalPoints   *int       `gorm:"column:total_points"`
}

func (QuizAnswer) TableName() string {
	return "quiz_answer"
}

// TeacherQuiz structure
type TeacherQuiz struct {
	Questions   []TeacherQuizQuestion `json:"questions"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
}

// TeacherQuizQuestion structure
type TeacherQuizQuestion struct {
	ID             string      `json:"id"`
	Type           string      `json:"type"`
	Points         int         `json:"points"`
	CorrectAnswer  interface{} `json:"correctAnswer,omitempty"`
	CorrectAnswers []string    `json:"correctAnswers,omitempty"`
	Question       string      `json:"question"`
}

// StudentQuizAnswersInput structure
type StudentQuizAnswersInput struct {
	ContentID  string                 `json:"content_id" validate:"required"`
	StartTime  time.Time              `json:"start_time" validate:"required"`
	EndTime    time.Time              `json:"end_time" validate:"required"`
	Answers    map[string]interface{} `json:"answers" validate:"required"`
	ReviewedAt *time.Time             `json:"reviewed_at,omitempty"`
}

// TextAnswerReviewInput structure
type TextAnswerReviewInput struct {
	QuizAnswerID  int     `json:"quiz_answer_id" validate:"required"`
	QuestionID    string  `json:"question_id" validate:"required"`
	IsCorrect     bool    `json:"is_correct"`
	PointsAwarded float64 `json:"points_awarded" validate:"gte=0"`
}

// QuizWithAttempts structure
type QuizWithAttempts struct {
	ContentID   string           `json:"content_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	QuizURL     string           `json:"quiz_url"`
	Attempts    []StudentAttempt `json:"attempts"`
}

// StudentAttempt structure
type StudentAttempt struct {
	QuizAnswerID    int                    `json:"quiz_answer_id"`
	UserID          string                 `json:"user_id"`
	StudentName     string                 `json:"student_name"`
	StudentLastname string                 `json:"student_lastname"`
	StudentEmail    string                 `json:"student_email"`
	Grade           *float64               `json:"grade"`
	TotalPoints     *int                   `json:"total_points"`
	NeedsReview     bool                   `json:"needs_review"`
	ReviewedAt      *time.Time             `json:"reviewed_at"`
	QuizAnswerURL   string                 `json:"quiz_answer_url"`
	Answers         map[string]interface{} `json:"answers"`
}

// StudentQuizResponse for student endpoint
type StudentQuizResponse struct {
	UserID       string              `json:"user_id"`
	Name         string              `json:"name"`
	Lastname     string              `json:"lastname"`
	Email        string              `json:"email"`
	QuizAttempts []QuizAttemptDetail `json:"quiz_attempts"`
}

// QuizAttemptDetail structure
type QuizAttemptDetail struct {
	QuizAnswerID  int                    `json:"quiz_answer_id"`
	ContentID     string                 `json:"content_id"`
	CourseID      int                    `json:"course_id"`
	CourseTitle   string                 `json:"course_title"`
	QuizTitle     string                 `json:"quiz_title"`
	Description   string                 `json:"description"`
	Grade         *float64               `json:"grade"`
	TotalPoints   *int                   `json:"total_points"`
	NeedsReview   bool                   `json:"needs_review"`
	ReviewedAt    *time.Time             `json:"reviewed_at"`
	QuizURL       string                 `json:"quiz_url"`
	QuizAnswerURL string                 `json:"quiz_answer_url"`
	Answers       map[string]interface{} `json:"answers"`
}

// QuizRepository interface
type QuizRepository interface {
	SaveQuizAttempt(attempt QuizAnswer) error
	UpdateQuizAttempt(attempt QuizAnswer) error
	FindQuizAttemptByID(quizAnswerID int) (QuizAnswer, error)
	FindQuizAttemptsByCourse(courseID int) ([]QuizAnswer, error)
	FindQuizAttemptsByUser(userID string) ([]QuizAnswer, error)
	GetQuizAttemptsByCourse(courseID int) ([]QuizAttemptView, error)
	GetQuizAttemptsByStudent(userID string) ([]QuizAttemptView, error)
}

type QuizAttemptView struct {
	QuizAnswerID      int        `gorm:"column:quiz_answer_id"`
	ContentID         string     `gorm:"column:content_id"`
	UserID            string     `gorm:"column:user_id"`
	StartTime         time.Time  `gorm:"column:start_time"`
	EndTime           time.Time  `gorm:"column:end_time"`
	Grade             *float64   `gorm:"column:grade"`
	ReviewedAt        *time.Time `gorm:"column:reviewed_at"`
	QuizURL           string     `gorm:"column:quiz_url"`
	QuizAnswerURL     string     `gorm:"column:quiz_answer_url"`
	TotalPoints       *int       `gorm:"column:total_points"`
	StudentName       string     `gorm:"column:student_name"`
	StudentLastname   string     `gorm:"column:student_lastname"`
	StudentEmail      string     `gorm:"column:student_email"`
	QuizTitle         string     `gorm:"column:quiz_title"`
	QuizDescription   string     `gorm:"column:quiz_description"`
	CourseContentID   int        `gorm:"column:course_content_id"`
	CourseID          int        `gorm:"column:course_id"`
	Module            string     `gorm:"column:module"`
	ModuleIndex       *int       `gorm:"column:module_index"`
	CourseTitle       string     `gorm:"column:course_title"`
	CourseDescription string     `gorm:"column:course_description"`
	TeacherID         string     `gorm:"column:teacher_id"`
	NeedsReview       bool       `gorm:"column:needs_review"`
	TotalQuizzes      int        `gorm:"column:total_quizzes"` // Nuevo campo
}

func (QuizAttemptView) TableName() string {
	return "quiz_attempts_view"
}

// Quiz representa un quiz con sus intentos
type Quiz struct {
	ContentID       string        `json:"content_id"`
	QuizURL         string        `json:"quiz_url"`
	QuizTitle       string        `json:"quiz_title"`
	QuizDescription string        `json:"quiz_description"`
	CourseContentID int           `json:"course_content_id"`
	Module          string        `json:"module"`
	ModuleIndex     *int          `json:"module_index"`
	Attempts        []QuizAttempt `json:"attempts"`
}

// CourseQuizResponse representa la respuesta del endpoint
type CourseQuizResponse struct {
	CourseID          int    `json:"course_id"`
	CourseTitle       string `json:"course_title"`
	CourseDescription string `json:"course_description"`
	TotalQuizzes      int    `json:"total_quizzes"` // Nuevo campo
	Quizzes           []Quiz `json:"quizzes"`
}

type QuizAttempt struct {
	QuizAnswerID    int        `json:"quiz_answer_id"`
	UserID          string     `json:"user_id"`
	StudentName     string     `json:"student_name"`
	StudentLastname string     `json:"student_lastname"`
	StudentEmail    string     `json:"student_email"`
	Grade           *float64   `json:"grade"`
	TotalPoints     *int       `json:"total_points"`
	NeedsReview     bool       `json:"needs_review"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	QuizAnswerURL   string     `json:"quiz_answer_url"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
}

// StudentCoursesQuizResponse representa la respuesta del endpoint para estudiantes, agrupando por curso
type StudentCoursesQuizResponse struct {
	UserID   string               `json:"user_id"`
	Name     string               `json:"name"`
	Lastname string               `json:"lastname"`
	Email    string               `json:"email"`
	Courses  []CourseQuizResponse `json:"courses"`
}
