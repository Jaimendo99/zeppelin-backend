package domain

type CourseDB struct {
	CourseID     int    `json:"id" gorm:"primaryKey"`
	TeacherID    string `json:"teacher_id" gorm:"not null"`
	StartDate    string `json:"start_date" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Description  string `json:"description"`
	QRCode       string `json:"qr_code" gorm:"unique"`
	StudentCount int    `json:"student_count" gorm:"-"`
}

type CourseTeacher struct {
	CourseID             int     `json:"id" gorm:"primaryKey"`
	TeacherID            string  `json:"teacher_id" gorm:"not null"`
	StartDate            string  `json:"start_date" validate:"required"`
	Title                string  `json:"title" validate:"required"`
	Description          string  `json:"description"`
	QRCode               string  `json:"qr_code" gorm:"unique"`
	StudentCount         int     `json:"student_count" gorm:"column:student_count"`
	ModuleCount          int     `json:"module_count" gorm:"column:module_count"`
	VideoCount           int     `json:"video_count" gorm:"column:video_count"`
	TextCount            int     `json:"text_count" gorm:"column:text_count"`
	QuizCount            int     `json:"quiz_count" gorm:"column:quiz_count"`
	CompletionPercentage float64 `json:"completion_percentage" gorm:"column:completion_percentage"`
}

type CourseInput struct {
	StartDate   string `json:"start_date" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type CourseRepo interface {
	CreateCourse(course CourseDB) error
	GetCoursesByTeacher(teacherID string) ([]CourseTeacher, error)
	GetCourseByTeacherAndCourseID(teacherID string, courseID int) (CourseDB, error)
}

func (CourseDB) TableName() string {
	return "course"
}
