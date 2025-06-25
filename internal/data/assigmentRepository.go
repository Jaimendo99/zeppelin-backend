package data

import (
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type assignmentRepo struct {
	db *gorm.DB
}

func NewAssignmentRepo(db *gorm.DB) domain.AssignmentRepo {
	return &assignmentRepo{db: db}
}

func (r *assignmentRepo) CreateAssignment(userID string, courseID int) error {
	assignment := domain.AssignmentDB{
		UserID:   userID,
		CourseID: courseID,
	}

	result := r.db.Create(&assignment)
	return result.Error
}

func (r *assignmentRepo) GetCourseIDByQRCode(qrCode string) (int, error) {
	var course struct {
		CourseID int
	}
	result := r.db.Table("course").Select("course_id").Where("qr_code = ?", qrCode).First(&course)
	if result.Error != nil {
		return 0, result.Error
	}
	return course.CourseID, nil
}

func (r *assignmentRepo) VerifyAssignment(assignmentID int) error {
	result := r.db.Model(&domain.AssignmentDB{}).
		Where("assignment_id = ?", assignmentID).
		UpdateColumns(map[string]interface{}{
			"is_active": true,
			"is_verify": true,
		})
	return result.Error
}

func (r *assignmentRepo) GetAssignmentsByStudent(userID string) ([]domain.StudentCourseProgress, error) {
	var progresses []domain.StudentCourseProgress
	err := r.db.
		Table(domain.StudentCourseProgress{}.TableName()).
		Where("user_id = ?", userID).
		Scan(&progresses).Error
	return progresses, err
}

func (r *assignmentRepo) GetStudentsByCourse(courseID int) ([]domain.AssignmentWithStudent, error) {
	var assignments []domain.AssignmentWithStudent
	result := r.db.Raw(`
        SELECT a.assignment_id, a.assigned_at, a.is_active, a.is_verify,
               u.user_id, u.name, u.lastname, u.email
        FROM assignment a
        JOIN "user" u ON a.user_id = u.user_id
        WHERE a.course_id = ?`, courseID).Scan(&assignments)
	return assignments, result.Error
}

func (r *assignmentRepo) GetAssignmentsByStudentAndCourse(userID string, courseID int) (domain.AssignmentWithCourse, error) {
	var assignment domain.AssignmentWithCourse
	// Cambié "assignment_with_courses" por "assignment" en la consulta
	err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&assignment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.AssignmentWithCourse{}, nil // No se encontró asignación para este estudiante y curso
		}
		return domain.AssignmentWithCourse{}, err // Error general
	}
	return assignment, nil
}
