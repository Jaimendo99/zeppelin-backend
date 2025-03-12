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

func (r *userRepo) GetAllTeachers() ([]domain.UserDb, error) {
	var teachers []domain.UserDb
	result := r.db.Where("type_id = ?", 2).Find(&teachers)
	if result.Error != nil {
		return nil, result.Error
	}
	return teachers, nil
}

func (r *userRepo) GetAllStudents() ([]domain.UserDb, error) {
	var students []domain.UserDb
	result := r.db.Where("type_id = ?", 3).Find(&students)
	if result.Error != nil {
		return nil, result.Error
	}
	return students, nil
}
