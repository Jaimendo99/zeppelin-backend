package data

import (
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type courseRepo struct {
	db *gorm.DB
}

func NewCourseRepo(db *gorm.DB) domain.CourseRepo {
	return &courseRepo{db: db}
}

func (r *courseRepo) CreateCourse(course domain.CourseDB) error {
	result := r.db.Create(&course)
	return result.Error
}

func (r *courseRepo) GetCoursesByTeacher(teacherID string) ([]domain.CourseTeacher, error) {
	var courses []domain.CourseTeacher
	result := r.db.Table("course_teacher_view").
		Where("teacher_id = ?", teacherID).
		Find(&courses)
	return courses, result.Error
}

func (r *courseRepo) GetCourseByTeacherAndCourseID(teacherID string, courseID int) (domain.CourseDB, error) {
	var course domain.CourseDB
	err := r.db.Where("teacher_id = ? AND course_id = ?", teacherID, courseID).First(&course).Error
	if err != nil {
		return domain.CourseDB{}, err // Si no encontramos el curso, devuelve un error
	}
	return course, nil // Devuelve el curso si lo encontró
}
