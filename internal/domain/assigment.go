package domain

import (
	"time"
)

type AssignmentDB struct {
	AssignmentID int        `json:"id" gorm:"primaryKey"`
	UserID       string     `json:"user_id" gorm:"not null"`
	CourseID     int        `json:"course_id" gorm:"not null"`
	AssignedAt   *time.Time `json:"assigned_at"`
	IsActive     *bool      `json:"is_active" gorm:"default:false"`
	IsVerify     *bool      `json:"is_verify" gorm:"default:false"`
}

type AssignmentDbRelation struct {
	AssignmentID uint      `gorm:"primaryKey;autoIncrement;column:assignment_id"`
	UserID       string    `gorm:"column:user_id;not null"`
	CourseID     uint      `gorm:"column:course_id;not null"`
	AssignedAt   time.Time `gorm:"column:assigned_at;autoCreateTime"`
	IsActive     bool      `gorm:"column:is_active;default:false"`
	IsVerify     bool      `gorm:"column:is_verify;default:false"`

	User   UserDbRelation   `gorm:"foreignKey:UserID;references:UserID"`
	Course CourseDbRelation `gorm:"foreignKey:CourseID;references:CourseID"`
}

func (AssignmentDbRelation) TableName() string {
	return "assignment"
}

type AssignmentWithCourse struct {
	AssignmentID int    `json:"assignment_id"`
	AssignedAt   string `json:"assigned_at"`
	IsActive     bool   `json:"is_active"`
	IsVerify     bool   `json:"is_verify"`
	CourseID     int    `json:"course_id"`
	TeacherID    string `json:"teacher_id"`
	StartDate    string `json:"start_date"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	QRCode       string `json:"qr_code"`
}

type AssignmentWithStudent struct {
	AssignmentID int    `json:"id"`
	AssignedAt   string `json:"assigned_at"`
	IsActive     bool   `json:"is_active"`
	IsVerify     bool   `json:"is_verify"`
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	Lastname     string `json:"lastname"`
	Email        string `json:"email"`
}

type StudentCourseProgress struct {
	UserID               string  `json:"user_id"`
	CourseID             int     `json:"course_id"`
	TeacherID            string  `json:"teacher_id"`
	StartDate            string  `json:"start_date"`
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	QRCode               string  `json:"qr_code"`
	ModuleCount          int64   `json:"module_count"`
	VideoCount           int64   `json:"video_count"`
	TextCount            int64   `json:"text_count"`
	QuizCount            int64   `json:"quiz_count"`
	CompletionPercentage float64 `json:"completion_percentage"`
}

// Indicamos que GORM use la vista en vez de una tabla
func (StudentCourseProgress) TableName() string {
	return "student_course_progress_view"
}

type AssignmentRepo interface {
	CreateAssignment(userID string, courseID int) error
	VerifyAssignment(assignmentID int) error
	GetAssignmentsByStudent(userID string) ([]StudentCourseProgress, error)
	GetStudentsByCourse(courseID int) ([]AssignmentWithStudent, error)
	GetCourseIDByQRCode(qrCode string) (int, error)
	GetAssignmentsByStudentAndCourse(userID string, courseID int) (AssignmentWithCourse, error)
}

func (AssignmentDB) TableName() string {
	return "assignment"
}

func (AssignmentWithCourse) TableName() string {
	return "assignment"
}
