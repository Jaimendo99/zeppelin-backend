package domain

import (
	"time"
)

// Modelo para la tabla quiz_answer
type QuizAnswer struct {
	QuizAnswerID  int        `gorm:"column:quiz_answer_id;primaryKey;autoIncrement"`
	ContentID     string     `gorm:"column:content_id"`
	UserID        string     `gorm:"column:user_id"`
	StartTime     time.Time  `gorm:"column:start_time"`
	EndTime       time.Time  `gorm:"column:end_time"`
	Grade         *float64   `gorm:"column:grade"`       // Usar puntero para aceptar NULL
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"` // Usar puntero para aceptar NULL
	QuizURL       string     `gorm:"column:quiz_url"`
	QuizAnswerURL string     `gorm:"column:quiz_answer_url"` // URL del archivo de respuestas del estudiante
	TotalPoints   *int       `gorm:"column:total_points"`    // Usar puntero para aceptar NULL

	// Definir relaciones si las necesitas
	// Content Content `gorm:"foreignKey:ContentID"`
	// User User `gorm:"foreignKey:UserID"`
}

func (QuizAnswer) TableName() string {
	return "quiz_answer"
}

// Struct para el Quiz con respuestas (versión del profesor) - Movido a domain
type TeacherQuiz struct {
	Questions []TeacherQuizQuestion `json:"questions"`
	// ... otros campos del quiz si existen
}

type TeacherQuizQuestion struct {
	ID             string      `json:"id"`
	Type           string      `json:"type"`
	Points         int         `json:"points"`
	CorrectAnswer  interface{} `json:"correctAnswer,omitempty"`
	CorrectAnswers []string    `json:"correctAnswers,omitempty"`
	// ... otros campos de la pregunta
}

// Interfaz para el Repositorio de Quizzes
type QuizRepository interface {
	SaveQuizAttempt(attempt QuizAnswer) error
	GetQuizAnswersByStudent(studentID string) ([]QuizSummary, error)
}

// Struct para la entrada de respuestas del estudiante
type StudentQuizAnswersInput struct {
	ContentID string                 `json:"content_id" validate:"required"`
	StartTime time.Time              `json:"start_time" validate:"required"`
	EndTime   time.Time              `json:"end_time" validate:"required"`
	Answers   map[string]interface{} `json:"answers" validate:"required"` // Mapa de pregunta_id a respuesta(s)
}

// Aquí puedes añadir otras interfaces o structs relacionados con quizzes si son necesarios en domain

type QuizAnswerDbRelation struct {
	Id          int        `gorm:"column:quiz_answer_id;primaryKey;autoIncrement"`
	ContentID   string     `gorm:"column:content_id"`
	Grade       *float64   `gorm:"column:grade"`
	StartDatime time.Time  `gorm:"column:start_time"`
	EndDatime   time.Time  `gorm:"column:end_time"`
	ReviewedAt  *time.Time `gorm:"column:reviewed_at"`
	TotalPoints *int       `gorm:"column:total_points"`
}

type QuizSummary struct {
	ContentID    string     `gorm:"column:content_id"`
	QuizCount    int        `gorm:"column:quiz_count"`
	TotalGrade   *float64   `gorm:"column:total_grade"`
	TotalPoints  *int       `gorm:"column:total_points"`
	LastQuizTime *time.Time `gorm:"column:last_quiz_time"`
}

func (QuizAnswerDbRelation) TableName() string {
	return "quiz_answer"
}

type QuizAnswerOutput struct {
	QuizAnswerID int        `json:"quiz_answer_id"`
	ContentID    string     `json:"content_id"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	Grade        *float64   `json:"grade"`
	ReviewedAt   *time.Time `json:"reviewed_at"`

	Content ContentOutput `json:"content"`
}

// func (q QuizAnswerDbRelation) ToOutput() QuizAnswerOutput {
// 	return QuizAnswerOutput{
// 		QuizAnswerID: q.Id,
// 		ContentID:    q.ContentID,
// 		StartTime:    q.StartDatime,
// 		EndTime:      q.EndDatime,
// 		Grade:        q.Grade,
// 		ReviewedAt:   q.ReviewedAt,
// 		Content: ContentOutput{
// 			ContentID:     q.Content.ContentID,
// 			ContentTypeID: q.Content.ContentTypeID,
// 			Title:         q.Content.Title,
// 			Url:           q.Content.Url,
// 			Description:   q.Content.Description,
// 			SectionIndex:  q.Content.SectionIndex,
// 		},
// 	}
// }
