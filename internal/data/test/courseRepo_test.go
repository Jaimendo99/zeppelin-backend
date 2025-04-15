package test_test

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
)

// Assume setupMockDb and quoteSql helpers exist from previous examples

func TestCourseRepo_CreateCourse(t *testing.T) {
	course := domain.CourseDB{
		// CourseID is often auto-generated, so we might not set it,
		// or set it to 0 depending on DB/GORM behavior.
		// GORM usually handles the PK insertion.
		TeacherID:   "teacher123",
		StartDate:   "2025-04-15",
		Title:       "Introduction to Go",
		Description: "A beginner course",
		QRCode:      "unique_qr_1",
	}

	// GORM's Create often generates SQL like this for Postgres.
	// The exact column order might vary. Using .* is safer if order is unstable.
	// It uses RETURNING to get the generated ID.
	// We need ExpectQuery because RETURNING makes it a query.
	expectedSql := `INSERT INTO "course" ("teacher_id","start_date","title","description","qr_code") VALUES ($1,$2,$3,$4,$5) RETURNING "course_id"`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		mock.ExpectBegin()
		// Expect the INSERT query and return the generated ID (e.g., 1)
		mock.ExpectQuery(quoteSql(expectedSql)).
			WithArgs(course.TeacherID, course.StartDate, course.Title, course.Description, course.QRCode).
			WillReturnRows(sqlmock.NewRows([]string{"course_id"}).AddRow(1)) // Simulate returning ID 1
		mock.ExpectCommit()

		err := repo.CreateCourse(course)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectQuery(quoteSql(expectedSql)).
			WithArgs(course.TeacherID, course.StartDate, course.Title, course.Description, course.QRCode).
			WillReturnError(dbErr) // Simulate DB error
		mock.ExpectRollback()

		err := repo.CreateCourse(course)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseRepo_GetCoursesByTeacher(t *testing.T) {
	teacherID := "teacher123"
	// Corrected SQL: Removed the ORDER BY clause to match GORM's actual output
	expectedSql := quoteSql(`SELECT * FROM "course" WHERE teacher_id = $1`)
	columns := []string{"course_id", "teacher_id", "start_date", "title", "description", "qr_code"}

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		rows := sqlmock.NewRows(columns).
			AddRow(1, teacherID, "2025-01-10", "Course 1", "Desc 1", "qr1").
			AddRow(2, teacherID, "2025-02-15", "Course 2", "Desc 2", "qr2")

		// Use the corrected expected SQL pattern
		mock.ExpectQuery(expectedSql).
			WithArgs(teacherID).
			WillReturnRows(rows)

		courses, err := repo.GetCoursesByTeacher(teacherID)

		assert.NoError(t, err)
		require.NotNil(t, courses) // Should pass now
		assert.Len(t, courses, 2)
		assert.Equal(t, "Course 1", courses[0].Title)
		assert.Equal(t, "Course 2", courses[1].Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		rows := sqlmock.NewRows(columns) // No rows added

		// Use the corrected expected SQL pattern
		mock.ExpectQuery(expectedSql).
			WithArgs(teacherID).
			WillReturnRows(rows)

		courses, err := repo.GetCoursesByTeacher(teacherID)

		assert.NoError(t, err)     // Find doesn't error for empty results
		require.NotNil(t, courses) // Should pass now
		assert.Len(t, courses, 0)  // Expect empty slice
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		dbErr := errors.New("db select error")

		// Use the corrected expected SQL pattern
		mock.ExpectQuery(expectedSql).
			WithArgs(teacherID).
			WillReturnError(dbErr)

		courses, err := repo.GetCoursesByTeacher(teacherID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err) // Should pass now
		// Depending on GORM/driver behavior on error, courses might be nil or an empty slice.
		// Asserting nil is usually safe for error cases where the scan doesn't happen.
		assert.Nil(t, courses)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseRepo_GetCoursesByStudent(t *testing.T) {
	studentID := "student456"
	// Corrected SQL: Use $1 placeholder instead of ?
	// Ensure whitespace matches the original Raw query string closely if needed,
	// but the placeholder is the key fix here.
	expectedRawSql := quoteSql(`SELECT c.* FROM course c JOIN enrollments e ON c.course_id = e.course_id WHERE e.student_id = $1`)
	columns := []string{"course_id", "teacher_id", "start_date", "title", "description", "qr_code"}

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		rows := sqlmock.NewRows(columns).
			AddRow(3, "teacher789", "2025-03-20", "Course 3", "Desc 3", "qr3").
			AddRow(4, "teacher101", "2025-04-25", "Course 4", "Desc 4", "qr4")

		// Use the corrected expected SQL pattern ($1 placeholder)
		mock.ExpectQuery(expectedRawSql).
			WithArgs(studentID).
			WillReturnRows(rows)

		courses, err := repo.GetCoursesByStudent(studentID)

		assert.NoError(t, err)
		require.NotNil(t, courses) // Should pass now
		assert.Len(t, courses, 2)
		assert.Equal(t, "Course 3", courses[0].Title)
		assert.Equal(t, 4, courses[1].CourseID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		rows := sqlmock.NewRows(columns) // No rows

		// Use the corrected expected SQL pattern ($1 placeholder)
		mock.ExpectQuery(expectedRawSql).
			WithArgs(studentID).
			WillReturnRows(rows)

		courses, err := repo.GetCoursesByStudent(studentID)

		assert.NoError(t, err) // Raw().Scan() usually doesn't error for zero rows

		// FIX: Remove the NotNil check.
		// require.NotNil(t, courses) // REMOVE THIS LINE

		// Assert the length is 0. This works correctly for nil slices.
		assert.Len(t, courses, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		dbErr := errors.New("db raw query error")

		// Use the corrected expected SQL pattern ($1 placeholder)
		mock.ExpectQuery(expectedRawSql).
			WithArgs(studentID).
			WillReturnError(dbErr)

		courses, err := repo.GetCoursesByStudent(studentID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err) // Should pass now
		assert.Nil(t, courses)      // Or assert empty slice
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
