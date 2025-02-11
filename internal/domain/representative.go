package domain

type Representative struct {
	RepresentativeId int    `json:"representative_id" gorm:"primaryKey"`
	Name             string `json:"name"`
	Lastname         string `json:"lastname"`
	Email            string `json:"email"`
	Phone            string `json:"phone_number"`
}

type RepresentativeDb struct {
	Name     string
	Lastname string
	Email    *string
	Phone    *string
}

type RepresentativeInput struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
	Phone    string `json:"phone_number"`
}

type RepresentativeRepo interface {
	CreateRepresentative(representative RepresentativeDb) error
	GetRepresentative(representative_id int) (RepresentativeDb, error)
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
