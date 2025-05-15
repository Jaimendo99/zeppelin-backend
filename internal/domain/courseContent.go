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
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	Contents        []Content `json:"contents" gorm:"foreignKey:CourseContentID;references:CourseContentID"`
}

type UserContent struct {
	UserID    string `gorm:"primaryKey"`
	ContentID string `gorm:"primaryKey"`
	StatusID  int
}

func (UserContent) TableName() string {
	return "user_content"
}

func (CourseContentDB) TableName() string {
	return "course_content"
}

type CourseContentStudentResult struct {
	CourseContentDB
	ContentType     string `json:"content_type"`
	StatusID        *int   `json:"status_id"`
	ContentID       string `json:"content_id"`
	CourseContentID int    `json:"course_content_id"`
	ContentTypeID   int    `json:"content_type_id"`
	Title           string `json:"title"`
	Url             string `json:"url"`
	Description     string `json:"description"`
	SectionIndex    int    `json:"section_index"`
}

type ContentWithStatus struct {
	ContentID       string `json:"content_id"`
	CourseContentID int    `json:"course_content_id"`
	ContentTypeID   int    `json:"content_type_id"`
	Title           string `json:"title"`
	Url             string `json:"url"`
	Description     string `json:"description"`
	SectionIndex    int    `json:"section_index"`
	StatusID        *int   `json:"status_id,omitempty"`
	IsActive        bool   `json:"is_active"`
}

type CourseContentWithStudentDetails struct {
	CourseContentID int                 `json:"course_content_id"`
	CourseID        int                 `json:"course_id"`
	Module          string              `json:"module"`
	ModuleIndex     int                 `json:"module_index"`
	CreatedAt       time.Time           `json:"created_at"`
	Details         []ContentWithStatus `json:"details"`
}
type Content struct {
	ContentID       string        `json:"content_id" gorm:"column:content_id;primaryKey" validate:"required"`
	CourseContentID int           `json:"course_content_id" gorm:"column:course_content_id" validate:"required"`
	ContentTypeID   int           `json:"content_type_id" gorm:"column:content_type_id" validate:"required,oneof=1 2 3"` // 1=video, 2=quiz, 3=text
	Title           string        `json:"title" gorm:"column:title" validate:"required,max=100"`
	Url             string        `json:"url" gorm:"column:url" validate:"omitempty,url"`
	Description     string        `json:"description" gorm:"column:description" validate:"omitempty"`
	SectionIndex    int           `json:"section_index" gorm:"column:section_index" validate:"required"`
	UserContent     []UserContent `gorm:"foreignKey:ContentID"`
	IsActive        bool          `json:"is_active"`
}

func (Content) TableName() string {
	return "content"
}

type CourseContentWithDetails struct {
	CourseContentDB
	Details []Content `json:"details"`
}

type AddModuleInput struct {
	CourseID int    `json:"course_id" validate:"required"`
	Module   string `json:"module" validate:"required,max=100"`
}

type AddSectionInput struct {
	CourseContentID int    `json:"course_content_id" validate:"required"`
	ContentTypeID   int    `json:"content_type_id" validate:"required,oneof=1 2 3"`
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
	CourseID    int             `json:"course_id" validate:"required"`
}

type UpdateContentStatusInput struct {
	ContentID string `json:"content_id" validate:"required"`
	IsActive  bool   `json:"is_active"`
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
	GetContentByCourse(courseID int) ([]CourseContentWithDetails, error)
	GetContentByCourseForStudent(courseID int, userID string) ([]CourseContentWithStudentDetails, error)
	CreateContent(input AddSectionInput) (string, error)
	AddSection(input AddSectionInput, userID string) (string, error)
	UpdateContent(input UpdateContentInput) error
	UpdateContentStatus(contentID string, isActive bool) error
	UpdateModuleTitle(courseContentID int, moduleTitle string) error
	UpdateUserContentStatus(userID, contentID string, statusID int) error
	GetContentTypeID(contentID string) (int, error)
	GetUrlByContentID(contentID string) (string, error)
}
