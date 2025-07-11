package domain

import "github.com/stretchr/testify/mock"

type User struct {
	UserID   string `json:"user_id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
	TypeID   int    `json:"type_id"`
}

type UserDb struct {
	UserID          string `json:"id" gorm:"column:user_id;primaryKey"`
	Name            string
	Lastname        string
	Email           string
	TypeID          int
	Representatives []RepresentativeDb `gorm:"foreignKey:UserID;references:UserID"`
	ParentalConsent *ParentalConsentDb `gorm:"foreignKey:UserID;references:UserID"`
}

type UserInput struct {
	Name           string              `json:"name" validate:"required"`
	Lastname       string              `json:"lastname" validate:"required"`
	Email          string              `json:"email" validate:"required,email"`
	Representative RepresentativeInput `json:"representative" validate:"required"`
}

type UserRepo interface {
	CreateUser(user UserDb) error
	GetUser(userID string) (*UserDb, error)
	GetAllTeachers() ([]UserDb, error)
	GetAllStudents() ([]UserDb, error)
}

func (UserDb) TableName() string {
	return "user"
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) CreateUser(user UserDb) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) GetUser(userID string) (*UserDb, error) {
	args := m.Called(userID)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(*UserDb), args.Error(1)
}

func (m *MockUserRepo) GetAllTeachers() ([]UserDb, error) {
	args := m.Called()
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.([]UserDb), args.Error(1)
}

func (m *MockUserRepo) GetAllStudents() ([]UserDb, error) {
	args := m.Called()
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.([]UserDb), args.Error(1)
}
