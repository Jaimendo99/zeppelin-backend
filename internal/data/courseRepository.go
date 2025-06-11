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

func (r *courseRepo) GetCoursesByStudent(studentID string) ([]domain.CourseDB, error) {
	var courses []domain.CourseDB
	result := r.db.Raw(`
		SELECT c.* FROM course c
        JOIN enrollments e ON c.course_id = e.course_id
        WHERE e.student_id = ?`, studentID).Scan(&courses)
	return courses, result.Error
}

func (r *courseRepo) GetCoursesByStudent2(studentID string) ([]domain.CourseDbRelation, error) {
	var assignments []domain.AssignmentDbRelation
	result := r.db.Where("user_id = ?", studentID).
		Preload("Course.Teacher").
		Preload("Course.CourseContent").
		// Preload("Course.CourseContent.Content").
		Find(&assignments)

	if result.Error != nil {
		return nil, result.Error
	}
	var courses []domain.CourseDbRelation
	for _, assignment := range assignments {
		courses = append(courses, assignment.Course)
	}
	return courses, nil
}

func (r *courseRepo) GetCourse(studentID, courseID string) (*domain.CourseDbRelation, error) {
	var assignments domain.AssignmentDbRelation
	result := r.db.Where("user_id = ? AND course_id = ?", studentID, courseID).
		Preload("Course.Teacher").
		Preload("Course.CourseContent").
		Preload("Course.CourseContent.Content").
		First(&assignments)

	if result.Error != nil {
		return &domain.CourseDbRelation{}, result.Error
	}
	return &assignments.Course, nil
}

func (r *courseRepo) GetCourseByTeacherAndCourseID(teacherID string, courseID int) (domain.CourseDB, error) {
	var course domain.CourseDB
	err := r.db.Where("teacher_id = ? AND course_id = ?", teacherID, courseID).First(&course).Error
	if err != nil {
		return domain.CourseDB{}, err // Si no encontramos el curso, devuelve un error
	}
	return course, nil // Devuelve el curso si lo encontró
}
