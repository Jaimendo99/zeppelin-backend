package data

import (
	"errors"
	"fmt"
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

func (r *courseContentRepo) AddModule(courseID int, module string, userID string) (int, error) {
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

func (r *courseContentRepo) GetContentTypeID(contentID string) (int, error) {
	var content domain.Content
	err := r.db.Table("content").
		Select("content_type_id").
		Where("content_id = ?", contentID).
		First(&content).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New("content not found")
		}
		return 0, err
	}
	return content.ContentTypeID, nil
}

func (r *courseContentRepo) GetContentByCourse(courseID int) ([]domain.CourseContentWithDetails, error) {
	var courseContents []domain.CourseContentDB
	// Dentro de tu repositorio (en GetContentByCourse)
	fmt.Printf("Conexión DB: %+v\n", r.db)

	query := r.db.
		Where("course_id = ?", courseID).
		Order("module_index")

	err := query.Preload("Contents", func(db *gorm.DB) *gorm.DB {
		return db.Order("section_index")
	}).Find(&courseContents).Error

	if err != nil {
		return nil, err
	}

	var result []domain.CourseContentWithDetails
	for _, cc := range courseContents {
		result = append(result, domain.CourseContentWithDetails{
			CourseContentDB: cc,
			Details:         cc.Contents,
		})
	}
	return result, nil
}

func (r *courseContentRepo) GetContentByCourseForStudent(courseID int, userID string) ([]domain.CourseContentWithStudentDetails, error) {
	var courseContents []domain.CourseContentDB

	query := r.db.
		Where("course_id = ?", courseID).
		Order("module_index")

	err := query.Preload("Contents", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true).Order("section_index").Preload("UserContent", "user_id = ?", userID)
	}).Find(&courseContents).Error

	if err != nil {
		return nil, err
	}

	var finalResult []domain.CourseContentWithStudentDetails
	for _, cc := range courseContents {
		var details []domain.ContentWithStatus

		for _, content := range cc.Contents {
			if len(content.UserContent) == 0 {
				continue
			}

			contentWithStatus := domain.ContentWithStatus{
				ContentID:       content.ContentID,
				CourseContentID: content.CourseContentID,
				ContentTypeID:   content.ContentTypeID,
				Title:           content.Title,
				Url:             content.Url,
				Description:     content.Description,
				SectionIndex:    content.SectionIndex,
				IsActive:        content.IsActive,
			}

			if len(content.UserContent) > 0 {
				statusID := content.UserContent[0].StatusID
				contentWithStatus.StatusID = &statusID
			}

			details = append(details, contentWithStatus)
		}

		if len(details) > 0 {
			finalResult = append(finalResult, domain.CourseContentWithStudentDetails{
				CourseContentID: cc.CourseContentID,
				CourseID:        cc.CourseID,
				Module:          cc.Module,
				ModuleIndex:     cc.ModuleIndex,
				CreatedAt:       cc.CreatedAt,
				Details:         details, // Only append if there are valid details
			})
		}
	}

	return finalResult, nil
}

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
func (r *courseContentRepo) UpdateContentStatus(contentID string, isActive bool) error {
	return r.db.Model(&domain.Content{}).
		Where("content_id = ?", contentID).
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

func (r *courseContentRepo) GetUrlByContentID(contentID string) (string, error) {
	var content domain.Content
	err := r.db.Table("content").
		Select("url").
		Where("content_id = ?", contentID).
		First(&content).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("content not found")
		}
		return "", err
	}
	return content.Url, nil
}
