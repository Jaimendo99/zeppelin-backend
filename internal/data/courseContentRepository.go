package data

import (
	"encoding/json"
	"gorm.io/gorm"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

// courseContentRepo unchanged
type courseContentRepo struct {
	db *gorm.DB
}

func NewCourseContentRepo(db *gorm.DB) domain.CourseContentRepo {
	return &courseContentRepo{db: db}
}

// GetContentByCourse updated to fetch json_content
func (r *courseContentRepo) GetContentByCourse(courseID int) ([]domain.CourseContentWithDetails, error) {
	var contents []domain.CourseContentDB
	err := r.db.Where("course_id = ?", courseID).Order("module, section_index").Find(&contents).Error
	if err != nil {
		return nil, err
	}

	var result []domain.CourseContentWithDetails
	for _, content := range contents {
		contentWithDetails := domain.CourseContentWithDetails{
			CourseContentDB: content,
		}

		switch content.ContentType {
		case "video":
			var video domain.VideoContent
			err = r.db.Table("video").Where("content_id = ?", content.ContentID).First(&video).Error
			if err == nil {
				contentWithDetails.Details = video
			}
		case "quiz":
			var quiz domain.QuizContent
			err = r.db.Table("quiz").Where("content_id = ?", content.ContentID).First(&quiz).Error
			if err == nil {
				contentWithDetails.Details = quiz
			}
		case "text":
			var text domain.TextContent
			err = r.db.Table("text").Where("content_id = ?", content.ContentID).First(&text).Error
			if err == nil {
				contentWithDetails.Details = text
			}
		}

		result = append(result, contentWithDetails)
	}

	return result, nil
}

// CreateVideo unchanged
func (r *courseContentRepo) CreateVideo(url, title, description string) (string, error) {
	contentID := controller.GenerateUID()

	video := map[string]interface{}{
		"content_id":  contentID,
		"url":         url,
		"title":       title,
		"description": description,
	}

	err := r.db.Table("video").Create(video).Error
	return contentID, err
}

// CreateQuiz updated to include json_content
func (r *courseContentRepo) CreateQuiz(title, description string, jsonContent json.RawMessage) (string, error) {
	contentID := controller.GenerateUID()

	quiz := map[string]interface{}{
		"content_id":   contentID,
		"title":        title,
		"description":  description,
		"json_content": jsonContent,
	}

	err := r.db.Table("quiz").Create(quiz).Error
	return contentID, err
}

// CreateText updated to include json_content and url
func (r *courseContentRepo) CreateText(title, url string, jsonContent json.RawMessage) (string, error) {
	contentID := controller.GenerateUID()

	text := map[string]interface{}{
		"content_id":   contentID,
		"title":        title,
		"url":          url,
		"json_content": jsonContent,
	}

	err := r.db.Table("text").Create(text).Error
	return contentID, err
}

// AddVideoSection, AddQuizSection, AddTextSection unchanged
func (r *courseContentRepo) AddVideoSection(courseID int, contentID string, module string, sectionIndex int, moduleIndex int) error {
	newSection := domain.CourseContentDB{
		CourseID:     courseID,
		Module:       module,
		ContentType:  "video",
		ContentID:    contentID,
		SectionIndex: sectionIndex,
		ModuleIndex:  moduleIndex,
	}

	return r.db.Create(&newSection).Error
}

func (r *courseContentRepo) AddQuizSection(courseID int, contentID string, module string, sectionIndex int, moduleIndex int) error {
	newSection := domain.CourseContentDB{
		CourseID:     courseID,
		Module:       module,
		ContentType:  "quiz",
		ContentID:    contentID,
		SectionIndex: sectionIndex,
		ModuleIndex:  moduleIndex,
	}

	return r.db.Create(&newSection).Error
}

func (r *courseContentRepo) AddTextSection(courseID int, contentID string, module string, sectionIndex int, moduleIndex int) error {
	newSection := domain.CourseContentDB{
		CourseID:     courseID,
		Module:       module,
		ContentType:  "text",
		ContentID:    contentID,
		SectionIndex: sectionIndex,
		ModuleIndex:  moduleIndex,
	}

	return r.db.Create(&newSection).Error
}

// New Update methods
func (r *courseContentRepo) UpdateVideo(contentID, title, url, description string) error {
	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if url != "" {
		updates["url"] = url
	}
	if description != "" {
		updates["description"] = description
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	return r.db.Table("video").Where("content_id = ?", contentID).Updates(updates).Error
}

func (r *courseContentRepo) UpdateQuiz(contentID, title, description string, jsonContent json.RawMessage) error {
	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if description != "" {
		updates["description"] = description
	}
	if jsonContent != nil {
		updates["json_content"] = jsonContent
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	return r.db.Table("quiz").Where("content_id = ?", contentID).Updates(updates).Error
}

func (r *courseContentRepo) UpdateText(contentID, title, url string, jsonContent json.RawMessage) error {
	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if url != "" {
		updates["url"] = url
	}
	if jsonContent != nil {
		updates["json_content"] = jsonContent
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	return r.db.Table("text").Where("content_id = ?", contentID).Updates(updates).Error
}

func (r *courseContentRepo) UpdateContentStatus(contentID string, isActive bool) error {
	return r.db.Model(&domain.CourseContentDB{}).
		Where("content_id = ?", contentID).
		Update("is_active", isActive).Error
}

func (r *courseContentRepo) UpdateModuleTitle(courseContentID int, moduleTitle string) error {
	// Actualizar el título del módulo
	return r.db.Model(&domain.CourseContentDB{}).
		Where("course_content_id = ?", courseContentID).
		Update("module", moduleTitle).Error
}
