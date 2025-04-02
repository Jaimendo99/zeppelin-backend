package domain

import "time"

type AssignmentDB struct {
	AssignmentID int        `json:"id" gorm:"primaryKey"`
	UserID       string     `json:"user_id" gorm:"not null"`
	CourseID     int        `json:"course_id" gorm:"not null"`
	AssignedAt   *time.Time `json:"assigned_at"`
	IsActive     *bool      `json:"is_active" gorm:"default:false"`
	IsVerify     *bool      `json:"is_verify" gorm:"default:false"`
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

type AssignmentRepo interface {
	CreateAssignment(userID string, courseID int) error
	VerifyAssignment(assignmentID int) error
	GetAssignmentsByStudent(userID string) ([]AssignmentWithCourse, error)
	GetStudentsByCourse(courseID int) ([]AssignmentWithStudent, error)
	GetCourseIDByQRCode(qrCode string) (int, error)
}

func (AssignmentDB) TableName() string {
	return "assignment"
}
