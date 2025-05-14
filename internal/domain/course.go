package domain

import (
	"time"
)

type CourseDB struct {
	CourseID    int    `json:"id" gorm:"primaryKey"`
	TeacherID   string `json:"teacher_id" gorm:"not null"`
	StartDate   string `json:"start_date" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	QRCode      string `json:"qr_code" gorm:"unique"`
}

type CourseDbRelation struct {
	CourseID    int       `gorm:"primaryKey;autoIncrement;column:course_id"`
	TeacherID   string    `gorm:"column:teacher_id;not null"`
	StartDate   time.Time `gorm:"column:start_date;not null"`
	Title       string    `gorm:"column:title;size:100;not null"`
	Description string    `gorm:"column:description"`
	QrCode      string    `gorm:"column:qr_code;unique"`

	CourseContent []CourseContentDb      `gorm:"foreignKey:CourseID;references:CourseID"`
	Assignments   []AssignmentDbRelation `gorm:"foreignKey:CourseID;references:CourseID"`
	Teacher       UserDbRelation         `gorm:"foreignKey:TeacherID;references:UserID"`
}

type CourseContentDb struct {
	CourseContentID int       `gorm:"primaryKey;autoIncrement;column:course_content_id"`
	CourseID        int       `gorm:"column:course_id;not null"`
	Module          string    `gorm:"column:module;size:100;not null"`
	ModuleIndex     int       `gorm:"column:module_index;not null"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	IsActive        bool      `gorm:"column:is_active;default:true"`

	Content []ContentDb `gorm:"foreignKey:CourseContentID;references:CourseContentID"`
}

func (CourseContentDb) TableName() string {
	return "course_content"
}

type ContentDb struct {
	ContentID       string `json:"content_id" gorm:"primaryKey"`
	CourseContentID int    `json:"course_content_id" gorm:"column:course_content_id;not null"`
	ContentTypeID   int    `json:"content_type_id" gorm:"not null"`
	Title           string `json:"title" gorm:"size:100;not null"`
	Url             string `json:"url" gorm:"size:255"`
	Description     string `json:"description" gorm:"size:255"`
	SectionIndex    int    `json:"section_index" gorm:"not null"`
}

func (ContentDb) TableName() string {
	return "content"
}

type TeacherOutput struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
}

type ContentOutput struct {
	ContentID     string `json:"content_id"`
	ContentTypeID int    `json:"content_type_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Url           string `json:"url"`
	SectionIndex  int    `json:"section_index"`
}

type ModuleOutput struct {
	ModuleID    int             `json:"module_id"`
	ModuleName  string          `json:"module_name"`
	ModuleIndex int             `json:"module_index"`
	Contents    []ContentOutput `json:"contents"`
}

type CourseOutput struct {
	CourseID    int            `json:"id"`
	Title       string         `json:"title"`
	StartDate   time.Time      `json:"start_date"`
	Description string         `json:"description"`
	Teacher     TeacherOutput  `json:"teacher"`
	Modules     []ModuleOutput `json:"modules"`
}

func (c *CourseDbRelation) ToCourseOutput() CourseOutput {
	output := CourseOutput{
		CourseID:    c.CourseID,
		Title:       c.Title,
		StartDate:   c.StartDate,
		Description: c.Description,
		Teacher: TeacherOutput{
			UserID:   c.Teacher.UserID,
			Name:     c.Teacher.Name,
			Lastname: c.Teacher.Lastname,
			Email:    c.Teacher.Email,
		},
		Modules: []ModuleOutput{}, // Initialize the slice
	}

	// Map CourseContent to Modules
	for _, cc := range c.CourseContent {
		module := ModuleOutput{
			ModuleID:    cc.CourseContentID,
			ModuleName:  cc.Module,
			ModuleIndex: cc.ModuleIndex,
			Contents:    []ContentOutput{}, // Initialize the slice
		}

		// Map Content to Contents within the module
		for _, content := range cc.Content {
			contentOutput := ContentOutput{
				ContentID:     content.ContentID,
				ContentTypeID: content.ContentTypeID,
				Title:         content.Title,
				Description:   content.Description,
				Url:           content.Url,
				SectionIndex:  content.SectionIndex,
			}
			module.Contents = append(module.Contents, contentOutput)
		}

		output.Modules = append(output.Modules, module)
	}

	return output
}

func (CourseDbRelation) TableName() string {
	return "course"
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
	GetCoursesByStudent2(studentID string) ([]CourseDbRelation, error)
	GetCourseByTeacherAndCourseID(teacherID string, courseID int) (CourseDB, error)
}

func (CourseDB) TableName() string {
	return "course"
}
