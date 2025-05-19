package data

import (
	"fmt"
	"zeppelin/internal/domain" // Importa tu paquete domain

	"gorm.io/gorm"
)

// Implementación del QuizRepository
type quizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) domain.QuizRepository {
	return &quizRepository{
		db: db,
	}
}

// SaveQuizAttempt crea un nuevo registro de intento de quiz
func (r *quizRepository) SaveQuizAttempt(attempt domain.QuizAnswer) error {
	if err := r.db.Create(&attempt).Error; err != nil {
		return fmt.Errorf("error creating new quiz attempt: %w", err)
	}
	return nil
}

func (r *quizRepository) GetQuizAnswersByStudent(userID string) ([]domain.QuizSummary, error) {
	var results []domain.QuizSummary

	err := r.db.Model(&domain.QuizAnswerDbRelation{}).
		Select("content_id, COUNT(*) as quiz_count, SUM(grade) as total_grade, SUM(total_points) as total_points, max(end_time) as last_quiz_time").
		Where("user_id = ?", userID).
		Group("content_id").
		Find(&results).Error

	return results, err
}
