package domain

import (
	"database/sql"
	"github.com/stretchr/testify/mock"
)

type Representative struct {
	RepresentativeId int    `json:"representative_id" gorm:"primaryKey"`
	Name             string `json:"name"`
	Lastname         string `json:"lastname"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phone_number"`
}

type RepresentativeDb struct {
	Name        string
	Lastname    string
	Email       sql.NullString
	PhoneNumber sql.NullString `gorm:"column:phone_number"`
}

type RepresentativeInput struct {
	Name        string `json:"name" validate:"required"`
	Lastname    string `json:"lastname" validate:"required"`
	Email       string `json:"email" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,e164" gorm:"column:phone_number"`
}

type RepresentativeRepo interface {
	CreateRepresentative(representative RepresentativeDb) error
	GetRepresentative(representativeId int) (*Representative, error)
	GetAllRepresentatives() ([]Representative, error)
	UpdateRepresentative(representativeId int, representative RepresentativeInput) error
}

type RepresentativeServiceI interface {
	CreateRepresentative(representative RepresentativeInput) error
	GetRepresentative(representativeId string) (RepresentativeDb, error)
	GetAllRepresentatives() ([]Representative, error)
	UpdateRepresentative(representativeId string, representative RepresentativeInput) error
}

// TableName overrides the default table name.
func (RepresentativeDb) TableName() string {
	return "representatives"
}
func (RepresentativeInput) TableName() string {
	return "representatives"
}

type MockRepresentativeRepo struct {
	mock.Mock
}

func (m *MockRepresentativeRepo) CreateRepresentative(representative RepresentativeDb) error {
	args := m.Called(representative)
	return args.Error(0)
}

func (m *MockRepresentativeRepo) GetRepresentative(representative_id int) (*Representative, error) {
	args := m.Called(representative_id)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(*Representative), args.Error(1)
}

func (m *MockRepresentativeRepo) GetAllRepresentatives() ([]Representative, error) {
	args := m.Called()
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.([]Representative), args.Error(1)
}

func (m *MockRepresentativeRepo) UpdateRepresentative(representative_id int, representative RepresentativeInput) error {
	args := m.Called(representative_id, representative)
	return args.Error(0)
}
