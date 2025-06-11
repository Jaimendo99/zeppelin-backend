package data

import (
	"fmt"
	"gorm.io/gorm"
	"zeppelin/internal/domain"
)

type quizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) domain.QuizRepository {
	return &quizRepository{
		db: db,
	}
}

func (r *quizRepository) SaveQuizAttempt(attempt domain.QuizAnswer) error {
	if err := r.db.Create(&attempt).Error; err != nil {
		return fmt.Errorf("error creating new quiz attempt: %w", err)
	}
	return nil
}

func (r *quizRepository) UpdateQuizAttempt(attempt domain.QuizAnswer) error {
	if err := r.db.Save(&attempt).Error; err != nil {
		return fmt.Errorf("error updating quiz attempt: %w", err)
	}
	return nil
}

func (r *quizRepository) FindQuizAttemptByID(quizAnswerID int) (domain.QuizAnswer, error) {
	var attempt domain.QuizAnswer
	if err := r.db.Where("quiz_answer_id = ?", quizAnswerID).First(&attempt).Error; err != nil {
		return domain.QuizAnswer{}, fmt.Errorf("error finding quiz attempt: %w", err)
	}
	return attempt, nil
}

func (r *quizRepository) FindQuizAttemptsByCourse(courseID int) ([]domain.QuizAnswer, error) {
	var attempts []domain.QuizAnswer
	err := r.db.Joins("JOIN content ON quiz_answer.content_id = content.content_id").
		Joins("JOIN course_content ON content.course_content_id = course_content.course_content_id").
		Where("course_content.course_id = ?", courseID).
		Find(&attempts).Error
	if err != nil {
		return nil, fmt.Errorf("error finding quiz attempts by course: %w", err)
	}
	return attempts, nil
}

func (r *quizRepository) FindQuizAttemptsByUser(userID string) ([]domain.QuizAnswer, error) {
	var attempts []domain.QuizAnswer
	err := r.db.Where("user_id = ?", userID).Find(&attempts).Error
	if err != nil {
		return nil, fmt.Errorf("error finding quiz attempts by user: %w", err)
	}
	return attempts, nil
}

// GetQuizAttemptsByCourse obtiene los intentos de quiz para un curso desde la vista
func (r *quizRepository) GetQuizAttemptsByCourse(courseID int) ([]domain.QuizAttemptView, error) {
	var attempts []domain.QuizAttemptView
	err := r.db.Where("course_id = ?", courseID).Find(&attempts).Error
	if err != nil {
		return nil, err
	}
	return attempts, nil
}

// GetQuizAttemptsByStudent obtiene los intentos de quiz para un estudiante desde la vista
func (r *quizRepository) GetQuizAttemptsByStudent(userID string) ([]domain.QuizAttemptView, error) {
	var attempts []domain.QuizAttemptView
	err := r.db.Where("user_id = ?", userID).Find(&attempts).Error
	if err != nil {
		return nil, err
	}
	return attempts, nil
}
