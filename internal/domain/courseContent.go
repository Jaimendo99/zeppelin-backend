package domain

import (
	"encoding/json"
	"time"
)

type CourseContentDB struct {
	CourseContentID int       `json:"course_content_id" gorm:"primaryKey"`
	CourseID        int       `json:"course_id" validate:"required"`
	Module          string    `json:"module" validate:"required"`
	ContentType     string    `json:"content_type" validate:"required,oneof=text video quiz"`
	ContentID       string    `json:"content_id" validate:"required"`
	SectionIndex    int       `json:"section_index"`
	ModuleIndex     int       `json:"module_index"`
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at"`
}

type UpdateModuleTitleInput struct {
	CourseContentID int    `json:"course_content_id" validate:"required"`
	ModuleTitle     string `json:"module_title" validate:"required"`
}

type UpdateContentStatusInput struct {
	ContentID string `json:"content_id" validate:"required"`
	IsActive  bool   `json:"is_active"`
}

type VideoContent struct {
	ContentID   string `json:"content_id" validate:"required"`
	Url         string `json:"url" validate:"omitempty,url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type QuizContent struct {
	ContentID   string          `json:"content_id" validate:"required"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	JsonContent json.RawMessage `json:"json_content,omitempty"`
}

type TextContent struct {
	ContentID   string          `json:"content_id" validate:"required"`
	Title       string          `json:"title" validate:"required"`
	JsonContent json.RawMessage `json:"json_content,omitempty"`
	Url         string          `json:"url,omitempty" validate:"omitempty,url"`
}

type AddVideoSectionInput struct {
	Url          string `json:"url" validate:"required"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Module       string `json:"module" validate:"required"`
	SectionIndex int    `json:"section_index"`
	ModuleIndex  int    `json:"module_index"`
}

type AddQuizSectionInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Module       string `json:"module" validate:"required"`
	SectionIndex int    `json:"section_index"`
	ModuleIndex  int    `json:"module_index"`
}

type AddTextSectionInput struct {
	Module       string `json:"module" validate:"required"`
	Title        string `json:"title" validate:"required"`
	SectionIndex int    `json:"section_index"`
	ModuleIndex  int    `json:"module_index"`
}

type UpdateVideoContentInput struct {
	ContentID   string `json:"content_id" validate:"required"`
	Title       string `json:"title,omitempty"`
	Url         string `json:"url,omitempty" validate:"omitempty,url"`
	Description string `json:"description,omitempty"`
}

type UpdateQuizContentInput struct {
	ContentID   string          `json:"content_id" validate:"required"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	JsonContent json.RawMessage `json:"json_content,omitempty"`
}

type UpdateTextContentInput struct {
	ContentID   string          `json:"content_id" validate:"required"`
	Title       string          `json:"title,omitempty"`
	JsonContent json.RawMessage `json:"json_content,omitempty"`
	Url         string          `json:"url,omitempty" validate:"omitempty,url"`
}

type CourseContentWithDetails struct {
	CourseContentDB
	Details interface{}
}

type CourseContentInput struct {
	Module      string `json:"module" validate:"required"`
	ContentType string `json:"content_type" validate:"required,oneof=text video quiz"`
	ContentID   string `json:"content_id" validate:"required"`
	ModuleIndex int    `json:"module_index,omitempty"`
}

type CourseContentRepo interface {
	GetContentByCourse(courseID int) ([]CourseContentWithDetails, error)
	CreateVideo(url, title, description string) (string, error)
	CreateQuiz(title, description string, jsonContent json.RawMessage) (string, error)
	CreateText(title, url string, jsonContent json.RawMessage) (string, error)
	AddVideoSection(courseID int, contentID, module string, sectionIndex, moduleIndex int) error
	AddQuizSection(courseID int, contentID, module string, sectionIndex, moduleIndex int) error
	AddTextSection(courseID int, contentID, module string, sectionIndex, moduleIndex int) error
	UpdateVideo(contentID, title, url, description string) error
	UpdateQuiz(contentID, title, description string, jsonContent json.RawMessage) error
	UpdateText(contentID, title, url string, jsonContent json.RawMessage) error
	UpdateContentStatus(contentID string, isActive bool) error
	UpdateModuleTitle(courseContentID int, moduleTitle string) error
}

func (CourseContentDB) TableName() string {
	return "course_content"
}
