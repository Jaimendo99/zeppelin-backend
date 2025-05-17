package test_test

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

func TestCourseRepo_GetCourseByTeacherAndCourseID(t *testing.T) {
	teacherID := "teacher123"
	courseID := 2

	// CORREGIDO: No escapes innecesarios de $ o .
	expectedSql := `SELECT \* FROM "course" WHERE teacher_id = \$1 AND course_id = \$2 ORDER BY "course"\."course_id" LIMIT \$3`
	columns := []string{"course_id", "teacher_id", "start_date", "title", "description", "qr_code"}

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		expectedCourse := domain.CourseDB{
			CourseID:    courseID,
			TeacherID:   teacherID,
			StartDate:   "2025-01-10",
			Title:       "Course 1",
			Description: "Desc 1",
			QRCode:      "qr1",
		}

		rows := sqlmock.NewRows(columns).
			AddRow(expectedCourse.CourseID, expectedCourse.TeacherID, expectedCourse.StartDate, expectedCourse.Title, expectedCourse.Description, expectedCourse.QRCode)

		mock.ExpectQuery(expectedSql).
			WithArgs(teacherID, courseID, 1).
			WillReturnRows(rows)

		course, err := repo.GetCourseByTeacherAndCourseID(teacherID, courseID)

		assert.NoError(t, err)
		assert.Equal(t, expectedCourse, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		mock.ExpectQuery(expectedSql).
			WithArgs(teacherID, courseID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		course, err := repo.GetCourseByTeacherAndCourseID(teacherID, courseID)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, domain.CourseDB{}, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		dbErr := errors.New("db select error")

		mock.ExpectQuery(expectedSql).
			WithArgs(teacherID, courseID, 1).
			WillReturnError(dbErr)

		course, err := repo.GetCourseByTeacherAndCourseID(teacherID, courseID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, domain.CourseDB{}, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseRepo_GetCourse(t *testing.T) {
	studentID := "student123"
	courseID := 456

	t.Run("Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		// GORM with PrepareStmt=true will emit LIMIT $3 and bind a 3rd arg=1
		sql := regexp.QuoteMeta(
			`SELECT * FROM "assignment" WHERE user_id = $1 AND course_id = $2 ` +
				`ORDER BY "assignment"."assignment_id" LIMIT $3`,
		)
		mock.ExpectQuery(sql).
			WithArgs(studentID, strconv.Itoa(courseID), 1). // Corrected line
			WillReturnError(gorm.ErrRecordNotFound)

		course, err := repo.GetCourse(studentID, strconv.Itoa(courseID))
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, &domain.CourseDbRelation{}, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		dbErr := errors.New("unexpected db error")

		sql := regexp.QuoteMeta(
			`SELECT * FROM "assignment" WHERE user_id = $1 AND course_id = $2 ` +
				`ORDER BY "assignment"."assignment_id" LIMIT $3`,
		)
		mock.ExpectQuery(sql).
			WithArgs(studentID, strconv.Itoa(courseID), 1). // Corrected line
			WillReturnError(dbErr)

		course, err := repo.GetCourse(studentID, strconv.Itoa(courseID))
		assert.Equal(t, dbErr, err)
		assert.Equal(t, &domain.CourseDbRelation{}, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestCourseRepo_GetCoursesByStudent2(t *testing.T) {
	studentID := "student123"

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		now := time.Now().Truncate(time.Second) // Truncate for consistent time comparison

		// Mock data
		mockAssignment1 := domain.AssignmentDbRelation{AssignmentID: 1, UserID: studentID, CourseID: 101, AssignedAt: now}
		mockAssignment2 := domain.AssignmentDbRelation{AssignmentID: 2, UserID: studentID, CourseID: 102, AssignedAt: now}

		mockCourse1 := domain.CourseDbRelation{CourseID: 101, TeacherID: "teacherA", Title: "Course 101", StartDate: now, QrCode: "qr101"}
		mockCourse2 := domain.CourseDbRelation{CourseID: 102, TeacherID: "teacherB", Title: "Course 102", StartDate: now, QrCode: "qr102"}

		mockTeacherA := domain.UserDbRelation{UserID: "teacherA", Name: "Alice", Lastname: "Smith", Email: "alice@example.com"}
		mockTeacherB := domain.UserDbRelation{UserID: "teacherB", Name: "Bob", Lastname: "Johnson", Email: "bob@example.com"}

		mockCourseContent1 := domain.CourseContentDb{CourseContentID: 201, CourseID: 101, Module: "Module 1 for Course 101", ModuleIndex: 1, CreatedAt: now, IsActive: true}
		mockCourseContent2 := domain.CourseContentDb{CourseContentID: 202, CourseID: 102, Module: "Module 1 for Course 102", ModuleIndex: 1, CreatedAt: now, IsActive: true}

		// 1. Mock query for assignments
		assignmentsQuery := quoteSql(`SELECT * FROM "assignment" WHERE user_id = $1`)
		assignmentsRows := sqlmock.NewRows([]string{"assignment_id", "user_id", "course_id", "assigned_at", "is_active", "is_verify"}).
			AddRow(mockAssignment1.AssignmentID, mockAssignment1.UserID, mockAssignment1.CourseID, mockAssignment1.AssignedAt, mockAssignment1.IsActive, mockAssignment1.IsVerify).
			AddRow(mockAssignment2.AssignmentID, mockAssignment2.UserID, mockAssignment2.CourseID, mockAssignment2.AssignedAt, mockAssignment2.IsActive, mockAssignment2.IsVerify)
		mock.ExpectQuery(assignmentsQuery).WithArgs(studentID).WillReturnRows(assignmentsRows)

		// 2. Mock query for courses (related to assignments)
		// GORM typically sorts IDs for IN clauses. CourseID is uint in AssignmentDbRelation, int in CourseDbRelation.
		// We pass uint to WithArgs as that's what GORM will use from the parent struct.
		coursesQuery := quoteSql(`SELECT * FROM "course" WHERE "course"."course_id" IN ($1,$2)`)
		courseRows := sqlmock.NewRows([]string{"course_id", "teacher_id", "start_date", "title", "description", "qr_code"}).
			AddRow(mockCourse1.CourseID, mockCourse1.TeacherID, mockCourse1.StartDate, mockCourse1.Title, mockCourse1.Description, mockCourse1.QrCode).
			AddRow(mockCourse2.CourseID, mockCourse2.TeacherID, mockCourse2.StartDate, mockCourse2.Title, mockCourse2.Description, mockCourse2.QrCode)
		mock.ExpectQuery(coursesQuery).WithArgs(mockAssignment1.CourseID, mockAssignment2.CourseID).WillReturnRows(courseRows)

		// 3. Mock query for course_content (related to courses)
		courseContentsQuery := quoteSql(`SELECT * FROM "course_content" WHERE "course_content"."course_id" IN ($1,$2)`)
		courseContentRows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at", "is_active"}).
			AddRow(mockCourseContent1.CourseContentID, mockCourseContent1.CourseID, mockCourseContent1.Module, mockCourseContent1.ModuleIndex, mockCourseContent1.CreatedAt, mockCourseContent1.IsActive).
			AddRow(mockCourseContent2.CourseContentID, mockCourseContent2.CourseID, mockCourseContent2.Module, mockCourseContent2.ModuleIndex, mockCourseContent2.CreatedAt, mockCourseContent2.IsActive)
		mock.ExpectQuery(courseContentsQuery).WithArgs(mockCourse1.CourseID, mockCourse2.CourseID).WillReturnRows(courseContentRows)

		// 5. Mock query for teachers (related to courses)
		teachersQuery := quoteSql(`SELECT * FROM "user" WHERE "user"."user_id" IN ($1,$2)`) // Assuming TeacherIDs are sorted by GORM if different
		teacherRows := sqlmock.NewRows([]string{"user_id", "name", "lastname", "email", "type_id"}).
			AddRow(mockTeacherA.UserID, mockTeacherA.Name, mockTeacherA.Lastname, mockTeacherA.Email, mockTeacherA.TypeID).
			AddRow(mockTeacherB.UserID, mockTeacherB.Name, mockTeacherB.Lastname, mockTeacherB.Email, mockTeacherB.TypeID)
		// The order of TeacherID in IN clause depends on the order of courses, assuming teacherA, teacherB for course1, course2
		mock.ExpectQuery(teachersQuery).WithArgs(mockCourse1.TeacherID, mockCourse2.TeacherID).WillReturnRows(teacherRows)

		courses, err := repo.GetCoursesByStudent2(studentID)

		assert.NoError(t, err)
		require.Len(t, courses, 2)

		// Assert Course 1
		assert.Equal(t, mockCourse1.Title, courses[0].Title)
		assert.Equal(t, mockCourse1.CourseID, courses[0].CourseID)
		assert.Equal(t, mockTeacherA.Name, courses[0].Teacher.Name)
		require.Len(t, courses[0].CourseContent, 1)
		assert.Equal(t, mockCourseContent1.Module, courses[0].CourseContent[0].Module)

		// Assert Course 2
		assert.Equal(t, mockCourse2.Title, courses[1].Title)
		assert.Equal(t, mockCourse2.CourseID, courses[1].CourseID)
		assert.Equal(t, mockTeacherB.Name, courses[1].Teacher.Name)
		require.Len(t, courses[1].CourseContent, 1)
		assert.Equal(t, mockCourseContent2.Module, courses[1].CourseContent[0].Module)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found (No Assignments)", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		assignmentsQuery := quoteSql(`SELECT * FROM "assignment" WHERE user_id = $1`)
		assignmentsRows := sqlmock.NewRows([]string{"assignment_id", "user_id", "course_id", "assigned_at", "is_active", "is_verify"}) // No rows
		mock.ExpectQuery(assignmentsQuery).WithArgs(studentID).WillReturnRows(assignmentsRows)

		courses, err := repo.GetCoursesByStudent2(studentID)

		assert.NoError(t, err)
		assert.Len(t, courses, 0) // Expect empty slice
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error on Assignments Query", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)

		dbErr := errors.New("db error on assignments")
		assignmentsQuery := quoteSql(`SELECT * FROM "assignment" WHERE user_id = $1`)
		mock.ExpectQuery(assignmentsQuery).WithArgs(studentID).WillReturnError(dbErr)

		courses, err := repo.GetCoursesByStudent2(studentID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, courses)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error on Courses Preload", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseRepo(gormDb)
		now := time.Now()

		mockAssignment1 := domain.AssignmentDbRelation{AssignmentID: 1, UserID: studentID, CourseID: 101, AssignedAt: now}

		assignmentsQuery := quoteSql(`SELECT * FROM "assignment" WHERE user_id = $1`)
		assignmentsRows := sqlmock.NewRows([]string{"assignment_id", "user_id", "course_id", "assigned_at", "is_active", "is_verify"}).
			AddRow(mockAssignment1.AssignmentID, mockAssignment1.UserID, mockAssignment1.CourseID, mockAssignment1.AssignedAt, false, false)
		mock.ExpectQuery(assignmentsQuery).WithArgs(studentID).WillReturnRows(assignmentsRows)

		dbErr := errors.New("db error on courses preload")
		// GORM optimizes IN clause with single argument to =
		coursesQuery := quoteSql(`SELECT * FROM "course" WHERE "course"."course_id" = $1`)
		mock.ExpectQuery(coursesQuery).WithArgs(mockAssignment1.CourseID).WillReturnError(dbErr)

		courses, err := repo.GetCoursesByStudent2(studentID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, courses) // Or empty, GORM behavior might vary, but error should be primary check
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
