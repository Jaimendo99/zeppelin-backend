package data

import (
	"gorm.io/gorm"
	"zeppelin/internal/domain"
)

type userPomodoroRepo struct {
	db *gorm.DB
}

func NewUserPomodoroRepo(db *gorm.DB) domain.UserPomodoroRepo {
	return &userPomodoroRepo{db: db}
}

func (r *userPomodoroRepo) GetByUserID(userID string) (*domain.UserPomodoro, error) {
	var p domain.UserPomodoro
	err := r.db.Where("user_id = ?", userID).First(&p).Error
	return &p, err
}

func (r *userPomodoroRepo) UpdateByUserID(userID string, input domain.UpdatePomodoroInput) error {
	return r.db.Model(&domain.UserPomodoro{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"active_time":    input.ActiveTime,
			"rest_time":      input.RestTime,
			"long_rest_time": input.LongRestTime,
			"iterations":     input.Iterations,
		}).Error
}
