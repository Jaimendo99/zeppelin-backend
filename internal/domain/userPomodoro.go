package domain

type UserPomodoro struct {
	PomodoroID   int    `json:"pomodoro_id" gorm:"primaryKey"`
	UserID       string `json:"user_id" gorm:"not null"`
	ActiveTime   int    `json:"active_time"`
	RestTime     int    `json:"rest_time"`
	LongRestTime int    `json:"long_rest_time"`
	Iterations   int    `json:"iterations"`
}

type UpdatePomodoroInput struct {
	ActiveTime   int `json:"active_time"`
	RestTime     int `json:"rest_time"`
	LongRestTime int `json:"long_rest_time"`
	Iterations   int `json:"iterations"`
}

type UserPomodoroRepo interface {
	GetByUserID(userID string) (*UserPomodoro, error)
	UpdateByUserID(userID string, input UpdatePomodoroInput) error
}

func (UserPomodoro) TableName() string {
	return "user_pomodoro"
}
