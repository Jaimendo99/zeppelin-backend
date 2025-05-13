package data

import (
	"errors"
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

// AddModule
func (r *courseContentRepo) AddModule(courseID int, module string, userID string) (int, error) {
	// Verify course ownership
	var course domain.CourseDB
	err := r.db.Table("course").
		Where("course_id = ? AND teacher_id = ?", courseID, userID).
		First(&course).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New("course does not belong to the teacher")
		}
		return 0, err
	}

	var courseContent domain.CourseContentDB
	err = r.db.Table("course_content").
		Where("course_id = ? AND module = ?", courseID, module).
		First(&courseContent).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	if err == gorm.ErrRecordNotFound {
		newModule := domain.CourseContentDB{
			CourseID: courseID,
			Module:   module,
			IsActive: true,
		}
		if err := r.db.Create(&newModule).Error; err != nil {
			return 0, err
		}
		return newModule.CourseContentID, nil
	}

	return courseContent.CourseContentID, nil
}

// VerifyModuleOwnership
func (r *courseContentRepo) VerifyModuleOwnership(courseContentID int, userID string) error {
	var courseContent domain.CourseContentDB
	err := r.db.Table("course_content").
		Joins("JOIN course ON course_content.course_id = course.course_id").
		Where("course_content.course_content_id = ? AND course.teacher_id = ?", courseContentID, userID).
		First(&courseContent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("module does not belong to the teacher's course")
		}
		return err
	}
	return nil
}

// CreateContent
func (r *courseContentRepo) CreateContent(input domain.AddSectionInput) (string, error) {
	var count int64
	if err := r.db.Table("course_content").Where("course_content_id = ?", input.CourseContentID).Count(&count).Error; err != nil {
		return "", err
	}
	if count == 0 {
		return "", gorm.ErrRecordNotFound
	}

	contentID := r.generateUID()
	content := domain.Content{
		ContentID:       contentID,
		CourseContentID: input.CourseContentID,
		ContentTypeID:   input.ContentTypeID,
		Title:           input.Title,
		Description:     input.Description,
	}

	if err := r.db.Table("content").Create(&content).Error; err != nil {
		return "", err
	}
	return contentID, nil
}

// AddSection
func (r *courseContentRepo) AddSection(input domain.AddSectionInput, userID string) (string, error) {
	// Verify module ownership
	if err := r.VerifyModuleOwnership(input.CourseContentID, userID); err != nil {
		return "", err
	}

	if input.ContentTypeID < 1 || input.ContentTypeID > 3 {
		return "", errors.New("invalid content_type_id")
	}

	return r.CreateContent(input)
}

// GetContentByCourse
func (r *courseContentRepo) GetContentByCourse(courseID int, isActive bool) ([]domain.CourseContentWithDetails, error) {
	var courseContents []domain.CourseContentDB
	query := r.db.Table("course_content").
		Where("course_id = ?", courseID).
		Order("module_index")

	if err := query.Find(&courseContents).Error; err != nil {
		return nil, err
	}

	var result []domain.CourseContentWithDetails
	for _, cc := range courseContents {
		var contents []struct {
			domain.Content
			ContentType string `json:"content_type"`
		}
		err := r.db.Table("content").
			Select("content.*, content_type.name AS content_type").
			Joins("LEFT JOIN content_type ON content.content_type_id = content_type.content_type_id").
			Where("content.course_content_id = ?", cc.CourseContentID).
			Order("content.section_index").
			Find(&contents).Error
		if err != nil {
			return nil, err
		}

		contentDetails := make([]domain.Content, len(contents))
		for i, c := range contents {
			contentDetails[i] = c.Content
		}

		result = append(result, domain.CourseContentWithDetails{
			CourseContentDB: cc,
			Details:         contentDetails,
		})
	}

	return result, nil
}

// GetContentByCourseForStudent
func (r *courseContentRepo) GetContentByCourseForStudent(courseID int, isActive bool, userID string) ([]domain.CourseContentWithDetails, error) {
	sqlQuery := `
		SELECT 
			cc.*, 
			ct.name AS content_type,
			uc.status_id, 
			c.content_id, c.course_content_id, c.content_type_id, c.title, c.url, c.description, c.section_index
		FROM 
			course_content cc
		LEFT JOIN content c ON cc.course_content_id = c.course_content_id
		LEFT JOIN content_type ct ON c.content_type_id = ct.content_type_id
		LEFT JOIN user_content uc ON c.content_id = uc.content_id AND uc.user_id = ?
		WHERE 
			cc.course_id = ? AND cc.is_active = ?
		ORDER BY 
			cc.module_index, c.section_index
	`

	var result []struct {
		domain.CourseContentDB
		ContentType     string `json:"content_type"`
		StatusID        int    `json:"status_id"`
		ContentID       string `json:"content_id"`
		CourseContentID int    `json:"course_content_id"`
		ContentTypeID   int    `json:"content_type_id"`
		Title           string `json:"title"`
		Url             string `json:"url"`
		Description     string `json:"description"`
		SectionIndex    int    `json:"section_index"`
	}

	err := r.db.Raw(sqlQuery, userID, courseID, isActive).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	// Group by course_content_id
	contentsByModule := make(map[int][]struct {
		ContentType string
		Details     domain.Content
		StatusID    *int
	})
	for _, r := range result {
		content := domain.Content{
			ContentID:       r.ContentID,
			CourseContentID: r.CourseContentID,
			ContentTypeID:   r.ContentTypeID,
			Title:           r.Title,
			Url:             r.Url,
			Description:     r.Description,
			SectionIndex:    r.SectionIndex,
		}
		contentsByModule[r.CourseContentID] = append(contentsByModule[r.CourseContentID], struct {
			ContentType string
			Details     domain.Content
			StatusID    *int
		}{
			ContentType: r.ContentType,
			Details:     content,
			StatusID:    &r.StatusID,
		})
	}

	var finalResult []domain.CourseContentWithDetails
	var courseContents []domain.CourseContentDB
	if err := r.db.Table("course_content").
		Where("course_id = ? AND is_active = ?", courseID, isActive).
		Order("module_index").
		Find(&courseContents).Error; err != nil {
		return nil, err
	}

	for _, cc := range courseContents {
		contents := contentsByModule[cc.CourseContentID]
		details := make([]domain.Content, len(contents))
		for i, c := range contents {
			details[i] = c.Details
		}
		var statusID *int
		if len(contents) > 0 {
			statusID = contents[0].StatusID
		}
		finalResult = append(finalResult, domain.CourseContentWithDetails{
			CourseContentDB: cc,
			Details:         details,
			StatusID:        statusID,
		})
	}

	return finalResult, nil
}

// UpdateContent
func (r *courseContentRepo) UpdateContent(input domain.UpdateContentInput) error {
	updates := map[string]interface{}{}
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Url != "" {
		updates["url"] = input.Url
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Table("content").Where("content_id = ?", input.ContentID).Updates(updates).Error
}

// UpdateContentStatus
func (r *courseContentRepo) UpdateContentStatus(courseContentID int, isActive bool) error {
	return r.db.Model(&domain.CourseContentDB{}).
		Where("course_content_id = ?", courseContentID).
		Update("is_active", isActive).Error
}

// UpdateModuleTitle
func (r *courseContentRepo) UpdateModuleTitle(courseContentID int, moduleTitle string) error {
	return r.db.Model(&domain.CourseContentDB{}).
		Where("course_content_id = ?", courseContentID).
		Update("module", moduleTitle).Error
}

// UpdateUserContentStatus
func (r *courseContentRepo) UpdateUserContentStatus(userID, contentID string, statusID int) error {
	return r.db.Table("user_content").
		Where("user_id = ? AND content_id = ?", userID, contentID).
		Update("status_id", statusID).Error
}
