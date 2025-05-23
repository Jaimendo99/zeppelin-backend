package data

import (
	"fmt"
	"gorm.io/gorm"
	"zeppelin/internal/domain" // Importa tu paquete domain
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
