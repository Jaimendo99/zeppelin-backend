package data

import (
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) domain.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) CreateUser(user domain.UserDb) error {
	result := r.db.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *userRepo) GetUser(userID string) (*domain.UserDb, error) {
	var user domain.UserDb
	result := r.db.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepo) GetAllUsers() ([]domain.UserDb, error) {
	var users []domain.UserDb
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
