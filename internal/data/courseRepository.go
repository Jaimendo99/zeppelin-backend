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

func (r *courseRepo) GetCoursesByTeacher(teacherID string) ([]domain.CourseDB, error) {
	var courses []domain.CourseDB
	result := r.db.Where("teacher_id = ?", teacherID).Find(&courses)
	return courses, result.Error
}

func (r *courseRepo) GetCoursesByStudent(studentID string) ([]domain.CourseDB, error) {
	var courses []domain.CourseDB
	result := r.db.Raw(`
        SELECT c.* FROM course c
        JOIN enrollments e ON c.course_id = e.course_id
        WHERE e.student_id = ?`, studentID).Scan(&courses)
	return courses, result.Error
}
