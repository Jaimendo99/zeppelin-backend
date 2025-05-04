package data

import (
	"encoding/json"
	"gorm.io/gorm"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
)

type courseContentRepo struct {
	db          *gorm.DB
	generateUID func() string
}

func NewCourseContentRepo(db *gorm.DB, generateUID func() string) domain.CourseContentRepo {
	if generateUID == nil {
		generateUID = controller.GenerateUID
	}
	return &courseContentRepo{
		db:          db,
		generateUID: generateUID,
	}
}

// CreateVideo corregido
func (r *courseContentRepo) CreateVideo(url, title, description string) (string, error) {
	contentID := r.generateUID()
	video := map[string]interface{}{
		"content_id":  contentID,
		"url":         url,
		"title":       title,
		"description": description,
	}

	if err := r.db.Table("video").Create(video).Error; err != nil {
		return "", err // Devolver cadena vacía en caso de error
	}
	return contentID, nil
}

// CreateQuiz corregido
func (r *courseContentRepo) CreateQuiz(title, url, description string, jsonContent json.RawMessage) (string, error) {
	contentID := r.generateUID()
	quiz := map[string]interface{}{
		"content_id":   contentID,
		"title":        title,
		"description":  description,
		"json_content": jsonContent,
		"url":          url,
	}

	if err := r.db.Table("quiz").Create(quiz).Error; err != nil {
		return "", err // Devolver cadena vacía en caso de error
	}
	return contentID, nil
}

// CreateText corregido
func (r *courseContentRepo) CreateText(title, url string, jsonContent json.RawMessage) (string, error) {
	contentID := r.generateUID()
	text := map[string]interface{}{
		"content_id":   contentID,
		"title":        title,
		"url":          url,
		"json_content": jsonContent,
	}

	if err := r.db.Table("text").Create(text).Error; err != nil {
		return "", err // Devolver cadena vacía en caso de error
	}
	return contentID, nil
}

func (r *courseContentRepo) GetContentByCourse(courseID int, isActive bool) ([]domain.CourseContentWithDetails, error) {
	var contents []domain.CourseContentDB
	// Filtramos según el valor de `isActive`
	query := r.db.Table("course_content").Where("course_id = ?", courseID)
	if isActive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("module, section_index").Scan(&contents).Error
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

func (r *courseContentRepo) GetContentByCourseForStudent(courseID int, isActive bool, userID string) ([]domain.CourseContentWithDetails, error) {

	var contents []domain.CourseContentWithStatus

	// Base query
	query := r.db.Table("course_content cc").Select("cc.*")
	// Si es estudiante, hacemos join con user_content if userID != nil {
	query = query.Select("cc.*, uc.status_id").
		Joins("LEFT JOIN user_content uc ON cc.content_id = uc.content_id AND uc.user_id = ?", userID)

	query = query.Where("cc.course_id = ?", courseID).Order("cc.module, cc.section_index")

	if isActive {
		query = query.Where("cc.is_active = ?", true)
	}

	if err := query.Scan(&contents).Error; err != nil {
		return nil, err
	}

	// Mapear a la respuesta esperada
	var result []domain.CourseContentWithDetails
	for _, c := range contents {
		contentWithDetails := domain.CourseContentWithDetails{
			CourseContentDB: c.CourseContentDB,
			StatusID:        c.StatusID, // Agregalo en tu struct si aún no existe
		}

		// Cargar detalles (video, quiz, text)
		switch c.ContentType {
		case "video":
			var video domain.VideoContent
			if err := r.db.Table("video").Where("content_id = ?", c.ContentID).First(&video).Error; err == nil {
				contentWithDetails.Details = video
			}
		case "quiz":
			var quiz domain.QuizContent
			if err := r.db.Table("quiz").Where("content_id = ?", c.ContentID).First(&quiz).Error; err == nil {
				contentWithDetails.Details = quiz
			}
		case "text":
			var text domain.TextContent
			if err := r.db.Table("text").Where("content_id = ?", c.ContentID).First(&text).Error; err == nil {
				contentWithDetails.Details = text
			}
		}

		result = append(result, contentWithDetails)
	}

	return result, nil
}
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
		return nil
	}

	return r.db.Table("video").Where("content_id = ?", contentID).Updates(updates).Error
}

func (r *courseContentRepo) UpdateQuiz(contentID, title, url, description string, jsonContent json.RawMessage) error {
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
	if jsonContent != nil {
		updates["json_content"] = jsonContent
	}

	if len(updates) == 0 {
		return nil
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
		return nil
	}

	return r.db.Table("text").Where("content_id = ?", contentID).Updates(updates).Error
}

func (r *courseContentRepo) UpdateContentStatus(contentID string, isActive bool) error {
	return r.db.Model(&domain.CourseContentDB{}).
		Where("content_id = ?", contentID).
		Update("is_active", isActive).Error
}

func (r *courseContentRepo) UpdateModuleTitle(courseContentID int, moduleTitle string) error {
	return r.db.Model(&domain.CourseContentDB{}).
		Where("course_content_id = ?", courseContentID).
		Update("module", moduleTitle).Error
}

func (r *courseContentRepo) UpdateUserContentStatus(userID, contentID string, statusID int) error {
	return r.db.Table("user_content").
		Where("user_id = ? AND content_id = ?", userID, contentID).
		Update("status_id", statusID).Error
}
