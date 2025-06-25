// data/assignment_repo_test.go (continued)
package test_test

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
)

func TestAssignmentRepo_CreateAssignment(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewAssignmentRepo(gormDb)

	userID := "user-123"
	courseID := 101

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin() // Expect transaction start

		expectedSql := quoteSql(`INSERT INTO "assignment" ("user_id","course_id","assigned_at","is_active","is_verify") VALUES ($1,$2,$3,$4,$5) RETURNING "assignment_id"`)

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, courseID, sqlmock.AnyArg(), false, false).          // Match args, use AnyArg for time, false for bool defaults
			WillReturnRows(sqlmock.NewRows([]string{"assignment_id"}).AddRow(1)) // Simulate returning the new ID

		mock.ExpectCommit()

		err := repo.CreateAssignment(userID, courseID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure", func(t *testing.T) {
		mock.ExpectBegin() // Expect transaction start

		expectedSql := quoteSql(`INSERT INTO "assignment" ("user_id","course_id","assigned_at","is_active","is_verify") VALUES ($1,$2,$3,$4,$5) RETURNING "assignment_id"`)
		expectedErr := errors.New("db error on insert")

		// Expect Query to fail
		mock.ExpectQuery(expectedSql).
			WithArgs(userID, courseID, sqlmock.AnyArg(), false, false).
			WillReturnError(expectedErr)

		mock.ExpectRollback() // Expect transaction rollback on error

		err := repo.CreateAssignment(userID, courseID)

		assert.Error(t, err)
		// GORM might wrap the error, check if it contains the original
		assert.Contains(t, err.Error(), expectedErr.Error())
		// Or assert specific GORM error type if applicable
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// data/assignment_repo_test.go

// Remove the quoteSql helper function

// ... setupMockDb ...

func TestAssignmentRepo_GetCourseIDByQRCode(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewAssignmentRepo(gormDb)

	qrCode := "qr-abc-123"
	expectedCourseID := 55

	// Use the actual SQL query format that GORM generates
	expectedSql := `SELECT "course_id" FROM "course" WHERE qr_code = \$1 ORDER BY "course"."course_id" LIMIT \$2`

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"course_id"}).AddRow(expectedCourseID)

		// PostgreSQL uses $1, $2, etc. for parameters (not $0)
		// GORM's First() method uses LIMIT 1, not LIMIT 55
		mock.ExpectQuery(expectedSql).
			WithArgs(qrCode, 1).
			WillReturnRows(rows)

		courseID, err := repo.GetCourseIDByQRCode(qrCode)

		assert.NoError(t, err)
		assert.Equal(t, expectedCourseID, courseID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		// Use the exact expectedSql string
		mock.ExpectQuery(expectedSql).
			WithArgs(qrCode, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		courseID, err := repo.GetCourseIDByQRCode(qrCode)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Equal(t, 0, courseID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		expectedErr := errors.New("connection failed")

		// Use the exact expectedSql string
		mock.ExpectQuery(expectedSql).
			WithArgs(qrCode, 1).
			WillReturnError(expectedErr)

		courseID, err := repo.GetCourseIDByQRCode(qrCode)

		assert.Error(t, err)
		// Check if the error contains the expected message, as GORM might wrap it
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.Equal(t, 0, courseID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ... rest of the tests ...

func TestAssignmentRepo_VerifyAssignment(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewAssignmentRepo(gormDb)

	assignmentID := 999

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		// UpdateColumns generates an UPDATE statement
		expectedSql := quoteSql(`UPDATE "assignment" SET "is_active"=$1,"is_verify"=$2 WHERE assignment_id = $3`)
		mock.ExpectExec(expectedSql).
			WithArgs(true, true, assignmentID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 insert id, 1 row affected
		mock.ExpectCommit()

		err := repo.VerifyAssignment(assignmentID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure", func(t *testing.T) {
		mock.ExpectBegin()
		expectedSql := quoteSql(`UPDATE "assignment" SET "is_active"=$1,"is_verify"=$2 WHERE assignment_id = $3`)
		expectedErr := errors.New("update constraint violation")
		mock.ExpectExec(expectedSql).
			WithArgs(true, true, assignmentID).
			WillReturnError(expectedErr)
		mock.ExpectRollback()

		err := repo.VerifyAssignment(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound (Update affects 0 rows)", func(t *testing.T) {
		// Simulate the case where the WHERE clause doesn't match any rows
		mock.ExpectBegin()
		expectedSql := quoteSql(`UPDATE "assignment" SET "is_active"=$1,"is_verify"=$2 WHERE assignment_id = $3`)
		mock.ExpectExec(expectedSql).
			WithArgs(true, true, assignmentID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		mock.ExpectCommit() // Update might still commit even if 0 rows affected

		err := repo.VerifyAssignment(assignmentID)

		// GORM update doesn't typically return ErrRecordNotFound if 0 rows affected,
		// it usually returns nil error. Check if your logic expects an error here.
		assert.NoError(t, err)
		// If you *want* an error when no rows are updated, you need to check
		// result.RowsAffected in the repository method itself.
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAssignmentRepo_GetAssignmentsByStudent(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewAssignmentRepo(gormDb)

	userID := "student-456"
	now := time.Now()

	expectedQuery := regexp.QuoteMeta(
		`SELECT * FROM "student_course_progress_view" WHERE user_id = $1`)
	// OJO: GORM genera SELECT * FROM "student_course_progress_view"..,
	// y para Scan genera una sola consulta

	// Simular las filas devueltas
	rows := sqlmock.NewRows([]string{
		"user_id", "course_id", "teacher_id", "start_date",
		"title", "description", "qr_code",
		"module_count", "video_count", "text_count", "quiz_count", "completion_percentage",
	}).
		AddRow(userID, 101, "teacher-1", now.Format(time.RFC3339),
			"Course 101", "Desc 1", "qr1", 5, 10, 3, 2, 50.5).
		AddRow(userID, 102, "teacher-2", now.Add(-time.Hour).Format(time.RFC3339),
			"Course 102", "Desc 2", "qr2", 2, 4, 1, 1, 25.0)

	mock.ExpectQuery(expectedQuery).
		WithArgs(userID).
		WillReturnRows(rows)

	progresses, err := repo.GetAssignmentsByStudent(userID)
	assert.NoError(t, err)
	require.Len(t, progresses, 2)

	p1 := progresses[0]
	assert.Equal(t, "student-456", p1.UserID)
	assert.Equal(t, 101, p1.CourseID)
	assert.Equal(t, "Course 101", p1.Title)
	assert.Equal(t, 5, int(p1.ModuleCount))
	assert.InDelta(t, 50.5, p1.CompletionPercentage, 1e-6)

	p2 := progresses[1]
	assert.Equal(t, 102, p2.CourseID)
	assert.Equal(t, "qr2", p2.QRCode)
	assert.InDelta(t, 25.0, p2.CompletionPercentage, 1e-6)

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestAssignmentRepo_GetStudentsByCourse(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewAssignmentRepo(gormDb)

	courseID := 202
	now := time.Now()

	// The actual SQL query that matches what your repository method generates
	expectedSql := `SELECT a.assignment_id, a.assigned_at, a.is_active, a.is_verify, u.user_id, u.name, u.lastname, u.email FROM assignment a JOIN "user" u ON a.user_id = u.user_id WHERE a.course_id = \$1`

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"assignment_id", "assigned_at", "is_active", "is_verify",
			"user_id", "name", "lastname", "email",
		}).
			AddRow(10, now, true, true, "student-1", "Alice", "A", "alice@test.com").
			AddRow(11, now.Add(-time.Hour), true, false, "student-2", "Bob", "B", "bob@test.com")

		mock.ExpectQuery(expectedSql).
			WithArgs(courseID).
			WillReturnRows(rows)

		assignments, err := repo.GetStudentsByCourse(courseID)

		assert.NoError(t, err)
		require.NotNil(t, assignments)
		assert.Len(t, assignments, 2)
		assert.Equal(t, 10, assignments[0].AssignmentID)
		assert.Equal(t, "student-1", assignments[0].UserID)
		assert.Equal(t, "Bob", assignments[1].Name)
		assert.Equal(t, false, assignments[1].IsVerify)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success_Empty", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"assignment_id", "assigned_at", "is_active", "is_verify",
			"user_id", "name", "lastname", "email",
		})

		mock.ExpectQuery(expectedSql).
			WithArgs(courseID).
			WillReturnRows(rows)

		assignments, err := repo.GetStudentsByCourse(courseID)

		assert.NoError(t, err)
		assert.Len(t, assignments, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("raw query failed again")
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID).
			WillReturnError(expectedErr)

		assignments, err := repo.GetStudentsByCourse(courseID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, assignments)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAssignmentRepo_GetAssignmentsByStudentAndCourse(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewAssignmentRepo(gormDb)

	userID := "user-123"
	courseID := 100

	t.Run("Success", func(t *testing.T) {
		expected := domain.AssignmentWithCourse{
			AssignmentID: 1,
			AssignedAt:   "2025-01-01T00:00:00Z",
			IsActive:     true,
			IsVerify:     false,
			CourseID:     courseID,
			TeacherID:    "teacher-1",
			StartDate:    "2024-12-01T00:00:00Z",
			Title:        "Go Basics",
			Description:  "Intro to Go",
			QRCode:       "qr-123",
		}

		rows := sqlmock.NewRows([]string{
			"assignment_id", "assigned_at", "is_active", "is_verify",
			"course_id", "teacher_id", "start_date", "title", "description", "qr_code",
		}).AddRow(
			expected.AssignmentID, expected.AssignedAt, expected.IsActive, expected.IsVerify,
			expected.CourseID, expected.TeacherID, expected.StartDate, expected.Title, expected.Description, expected.QRCode,
		)

		mock.ExpectQuery(`SELECT \* FROM "assignment" WHERE user_id = \$1 AND course_id = \$2 ORDER BY "assignment"\."assignment_id" LIMIT \$3`).
			WithArgs(userID, courseID, 1).
			WillReturnRows(rows)

		result, err := repo.GetAssignmentsByStudentAndCourse(userID, courseID)

		assert.NoError(t, err)
		assert.Equal(t, expected.AssignmentID, result.AssignmentID)
		assert.Equal(t, expected.CourseID, result.CourseID)
		assert.Equal(t, expected.Title, result.Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "assignment" WHERE user_id = \$1 AND course_id = \$2 ORDER BY "assignment"\."assignment_id" LIMIT \$3`).
			WithArgs(userID, courseID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.GetAssignmentsByStudentAndCourse(userID, courseID)

		assert.NoError(t, err)
		assert.Equal(t, domain.AssignmentWithCourse{}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		dbErr := errors.New("db connection failed")

		mock.ExpectQuery(`SELECT \* FROM "assignment" WHERE user_id = \$1 AND course_id = \$2 ORDER BY "assignment"\."assignment_id" LIMIT \$3`).
			WithArgs(userID, courseID, 1).
			WillReturnError(dbErr)

		result, err := repo.GetAssignmentsByStudentAndCourse(userID, courseID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, domain.AssignmentWithCourse{}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
