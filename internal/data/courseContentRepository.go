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
	// Consulta SQL pura para obtener los datos combinados
	sqlQuery := `
		SELECT 
			cc.*, 
			uc.status_id, 
			v.content_id AS video_content_id, v.url AS video_url, v.title AS video_title, v.description AS video_description, 
			q.content_id AS quiz_content_id, q.title AS quiz_title, q.description AS quiz_description, q.json_content AS quiz_json_content, q.url AS quiz_url,         
			t.content_id AS text_content_id, t.title AS text_title, t.json_content AS text_json_content, t.url AS text_url 
		FROM 
			course_content cc
		LEFT JOIN user_content uc ON cc.content_id = uc.content_id AND uc.user_id = ?
		LEFT JOIN video v ON cc.content_id = v.content_id AND cc.content_type = 'video'
		LEFT JOIN quiz q ON cc.content_id = q.content_id AND cc.content_type = 'quiz'
		LEFT JOIN text t ON cc.content_id = t.content_id AND cc.content_type = 'text'
		WHERE 
			cc.course_id = ? AND cc.is_active = ?
		ORDER BY 
			cc.module, cc.section_index
	`

	var result []struct {
		domain.CourseContentDB
		StatusID         int             `json:"status_id"` // Cambio a tipo simple
		VideoContentID   string          `json:"video_content_id"`
		VideoUrl         string          `json:"video_url"`
		VideoTitle       string          `json:"video_title"`
		VideoDescription string          `json:"video_description"`
		QuizContentID    string          `json:"quiz_content_id"`
		QuizTitle        string          `json:"quiz_title"`
		QuizDescription  string          `json:"quiz_description"`
		QuizJsonContent  json.RawMessage `json:"quiz_json_content"`
		QuizUrl          string          `json:"quiz_url"`
		TextContentID    string          `json:"text_content_id"`
		TextTitle        string          `json:"text_title"`
		TextJsonContent  json.RawMessage `json:"text_json_content"`
		TextUrl          string          `json:"text_url"`
	}

	// Ejecutar la consulta SQL en bruto
	err := r.db.Raw(sqlQuery, userID, courseID, isActive).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	var finalResult []domain.CourseContentWithDetails
	for _, content := range result {
		contentWithDetails := domain.CourseContentWithDetails{
			CourseContentDB: content.CourseContentDB,
			StatusID:        &content.StatusID, // Aquí convertimos a puntero solo para la estructura final
		}

		// Asignar detalles según el tipo de contenido
		switch content.ContentType {
		case "video":
			// Verificamos si los campos de video no están vacíos
			if content.VideoContentID != "" && content.VideoUrl != "" && content.VideoTitle != "" && content.VideoDescription != "" {
				contentWithDetails.Details = domain.VideoContent{
					ContentID:   content.VideoContentID,
					Url:         content.VideoUrl,
					Title:       content.VideoTitle,
					Description: content.VideoDescription,
				}
			}
		case "quiz":
			// Verificamos si los campos de quiz no están vacíos
			if content.QuizContentID != "" && content.QuizTitle != "" && content.QuizDescription != "" {
				contentWithDetails.Details = domain.QuizContent{
					ContentID:   content.QuizContentID,
					Title:       content.QuizTitle,
					Description: content.QuizDescription,
					JsonContent: content.QuizJsonContent,
					Url:         content.QuizUrl,
				}
			}
		case "text":
			// Verificamos si los campos de texto no están vacíos
			if content.TextContentID != "" && content.TextTitle != "" && content.TextJsonContent != nil {
				contentWithDetails.Details = domain.TextContent{
					ContentID:   content.TextContentID,
					Title:       content.TextTitle,
					JsonContent: content.TextJsonContent,
					Url:         content.TextUrl,
				}
			} else {
				// Si el contenido de texto está vacío, asignamos un valor vacío a Details
				contentWithDetails.Details = domain.TextContent{}
			}
		}

		finalResult = append(finalResult, contentWithDetails)
	}

	return finalResult, nil
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
