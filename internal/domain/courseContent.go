package domain

import (
	"encoding/json"
	"time"
)

type CourseContentDB struct {
	CourseContentID int       `json:"course_content_id" gorm:"column:course_content_id;primaryKey;autoIncrement"`
	CourseID        int       `json:"course_id" gorm:"column:course_id" validate:"required"`
	Module          string    `json:"module" gorm:"column:module" validate:"required,max=100"`
	ModuleIndex     int       `json:"module_index" gorm:"column:module_index"`
	IsActive        bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (CourseContentDB) TableName() string {
	return "course_content"
}

type Content struct {
	ContentID       string `json:"content_id" gorm:"column:content_id;primaryKey" validate:"required"`
	CourseContentID int    `json:"course_content_id" gorm:"column:course_content_id" validate:"required"`
	ContentTypeID   int    `json:"content_type_id" gorm:"column:content_type_id" validate:"required,oneof=1 2 3"`
	Title           string `json:"title" gorm:"column:title" validate:"required,max=100"`
	Url             string `json:"url" gorm:"column:url" validate:"omitempty,url"`
	Description     string `json:"description" gorm:"column:description" validate:"omitempty"`
	SectionIndex    int    `json:"section_index" gorm:"column:section_index" validate:"required"`
}

func (Content) TableName() string {
	return "content"
}

type CourseContentWithDetails struct {
	CourseContentDB
	Details  []Content `json:"details"`
	StatusID *int      `json:"status_id,omitempty"`
}

type AddModuleInput struct {
	CourseID int    `json:"course_id" validate:"required"`
	Module   string `json:"module" validate:"required,max=100"`
}

type AddSectionInput struct {
	CourseContentID int    `json:"course_content_id" validate:"required"`
	ContentTypeID   int    `json:"content_type_id" validate:"required,oneof=1 2 3"` // 1=video, 2=quiz, 3=text
	Title           string `json:"title" validate:"required,max=100"`
	Description     string `json:"description" validate:"omitempty"`
}

type UpdateContentInput struct {
	ContentID   string          `json:"content_id" validate:"required"`
	Title       string          `json:"title" validate:"max=100"`
	Url         string          `json:"url" validate:"omitempty,url"`
	Description string          `json:"description" validate:"omitempty"`
	JsonData    json.RawMessage `json:"json_data" validate:"omitempty,json"`
	VideoID     string          `json:"video_id" validate:"omitempty"`
}

type UpdateContentStatusInput struct {
	CourseContentID int  `json:"course_content_id" validate:"required"`
	IsActive        bool `json:"is_active"`
}

type UpdateModuleTitleInput struct {
	CourseContentID int    `json:"course_content_id" validate:"required"`
	ModuleTitle     string `json:"module_title" validate:"required,max=100"`
}

type UpdateUserContentStatusInput struct {
	ContentID string `json:"content_id" validate:"required"`
}

type CourseContentInput struct {
	Module      string `json:"module" validate:"required,max=100"`
	ContentType string `json:"content_type" validate:"required,oneof=text video quiz"`
	ModuleIndex int    `json:"module_index"`
}

type CourseContentWithStatus struct {
	CourseContentDB
	StatusID *int `json:"status_id"`
}

type CourseContentRepo interface {
	AddModule(courseID int, module string, userID string) (int, error)
	VerifyModuleOwnership(courseContentID int, userID string) error
	GetContentByCourse(courseID int, isActive bool) ([]CourseContentWithDetails, error)
	GetContentByCourseForStudent(courseID int, isActive bool, userID string) ([]CourseContentWithDetails, error)
	CreateContent(input AddSectionInput) (string, error)
	AddSection(input AddSectionInput, userID string) (string, error)
	UpdateContent(input UpdateContentInput) error
	UpdateContentStatus(courseContentID int, isActive bool) error
	UpdateModuleTitle(courseContentID int, moduleTitle string) error
	UpdateUserContentStatus(userID, contentID string, statusID int) error
}
