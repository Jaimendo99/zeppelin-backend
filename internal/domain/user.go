package domain

type User struct {
	UserID   string `json:"user_id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
	TypeID   int    `json:"type_id"`
}

type UserDb struct {
	UserID   string
	Name     string
	Lastname string
	Email    string
	TypeID   int
}

type UserInput struct {
	Name     string `json:"name" validate:"required"`
	Lastname string `json:"lastname" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type UserRepo interface {
	CreateUser(user UserDb) error
	GetUser(userID string) (*UserDb, error)
	GetAllUsers() ([]UserDb, error)
}

func (UserDb) TableName() string {
	return "user"
}
