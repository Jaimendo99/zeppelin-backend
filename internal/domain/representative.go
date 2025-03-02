package domain

import "database/sql"

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
	PhoneNumber sql.NullString `gorm:"column:phone"`
}

type RepresentativeInput struct {
	Name        string `json:"name" validate:"required"`
	Lastname    string `json:"lastname" validate:"required"`
	Email       string `json:"email" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,e164" gorm:"column:phone_number"`
}

type RepresentativeRepo interface {
	CreateRepresentative(representative RepresentativeDb) error
	GetRepresentative(representative_id int) (*RepresentativeInput, error)
	GetAllRepresentatives() ([]Representative, error)
	UpdateRepresentative(representative_id int, representative RepresentativeInput) error
	//DeleteRepresentative(representative_id int) error
}

type RepresentativeServiceI interface {
	CreateRepresentative(representative RepresentativeInput) error
	GetRepresentative(representative_id string) (RepresentativeDb, error)
	GetAllRepresentatives() ([]Representative, error)
	UpdateRepresentative(representative_id string, representative RepresentativeInput) error
}

// TableName overrides the default table name.
func (RepresentativeDb) TableName() string {
	return "representatives"
}
func (RepresentativeInput) TableName() string {
	return "representatives"
}
