package domain

type CourseDB struct {
	CourseID    int    `json:"id" gorm:"primaryKey"`
	TeacherID   string `json:"teacher_id" gorm:"not null"`
	StartDate   string `json:"start_date" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	QRCode      string `json:"qr_code" gorm:"unique"`
}

type CourseInput struct {
	StartDate   string `json:"start_date" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type CourseRepo interface {
	CreateCourse(course CourseDB) error
	GetCoursesByTeacher(teacherID string) ([]CourseDB, error)
	GetCoursesByStudent(studentID string) ([]CourseDB, error)
	GetCourseByTeacherAndCourseID(teacherID string, courseID int) (CourseDB, error)
}

func (CourseDB) TableName() string {
	return "course"
}
