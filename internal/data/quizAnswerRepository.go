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

func (r *quizRepository) GetQuizAnswersByStudent(studentID string) ([]domain.QuizAnswerDbRelation, error) {
	var quizAnswers []domain.QuizAnswerDbRelation
	result := r.db.Where("user_id = ?", studentID).
		Preload("Content").
		Find(&quizAnswers)
	if result.Error != nil {
		return nil, result.Error
	}
	return quizAnswers, nil
}
