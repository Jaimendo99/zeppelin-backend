package domain

import (
	"time"
)

type Session struct {
	SessionID int        `json:"session_id" gorm:"column:session_id;primaryKey;autoIncrement"`
	UserID    string     `json:"user_id" gorm:"column:user_id" validate:"required"`
	Start     time.Time  `json:"start" gorm:"column:start"`
	End       *time.Time `json:"end" gorm:"column:end"` // Usamos *time.Time para permitir NULL
}

func (Session) TableName() string {
	return "session"
}

type SessionCourseContent struct {
	SessionCourseContentID int       `json:"session_course_content_id" gorm:"column:session_course_content_id;primaryKey;autoIncrement"`
	SessionID              int       `json:"session_id" gorm:"column:session_id"`
	CourseContentID        int       `json:"course_content_id" gorm:"column:course_content_id"`
	AccessedAt             time.Time `json:"accessed_at" gorm:"column:accessed_at;autoCreateTime"`

	Session       Session         `gorm:"foreignKey:SessionID;references:SessionID"`
	CourseContent CourseContentDB `gorm:"foreignKey:CourseContentID;references:CourseContentID"`
}

func (SessionCourseContent) TableName() string {
	return "session_course_content"
}

type SessionRepo interface {
	StartSession(userID string) (int, error)
	EndSession(sessionID int) error
	GetActiveSessionByUserID(userID string) (*Session, error)
}
