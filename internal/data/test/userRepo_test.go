package test_test

import (
	"errors"
	"testing"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Assume setupMockDb and quoteSql helpers exist from previous examples

func TestUserRepo_CreateUser(t *testing.T) {
	user := domain.UserDb{
		UserID:   "user-abc-123", // String primary key
		Name:     "Test",
		Lastname: "User",
		Email:    "test.user@example.com",
		TypeID:   3, // Example: Student
	}

	// GORM's Create for models with non-integer/non-auto-increment PKs
	// often just performs a simple INSERT without RETURNING.
	// Therefore, we expect an Exec, not a Query.
	expectedSql := `INSERT INTO "user" ("user_id","name","lastname","email","type_id") VALUES ($1,$2,$3,$4,$5)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		mock.ExpectBegin()
		// Expect an Exec because no RETURNING clause is typically used here
		mock.ExpectExec(quoteSql(expectedSql)).
			WithArgs(user.UserID, user.Name, user.Lastname, user.Email, user.TypeID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1, LastInsertId usually 0 for non-serial PKs
		mock.ExpectCommit()

		err := repo.CreateUser(user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		dbErr := errors.New("db insert error - unique constraint?")

		mock.ExpectBegin()
		mock.ExpectExec(quoteSql(expectedSql)).
			WithArgs(user.UserID, user.Name, user.Lastname, user.Email, user.TypeID).
			WillReturnError(dbErr) // Simulate DB error
		mock.ExpectRollback()

		err := repo.CreateUser(user)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_GetUser(t *testing.T) {
	userID := "user-def-456"
	// GORM's First usually adds ORDER BY primary_key LIMIT 1
	expectedSql := quoteSql(`SELECT * FROM "user" WHERE user_id = $1 ORDER BY "user"."user_id" LIMIT $2`)
	columns := []string{"user_id", "name", "lastname", "email", "type_id"}

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		expectedUser := domain.UserDb{
			UserID:   userID,
			Name:     "Found",
			Lastname: "User",
			Email:    "found@example.com",
			TypeID:   2, // Teacher
		}
		rows := sqlmock.NewRows(columns).
			AddRow(expectedUser.UserID, expectedUser.Name, expectedUser.Lastname, expectedUser.Email, expectedUser.TypeID)

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1). // Arg 1 is userID, Arg 2 is LIMIT 1
			WillReturnRows(rows)

		user, err := repo.GetUser(userID)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser, *user) // Compare the struct content
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound) // Simulate GORM's not found error

		user, err := repo.GetUser(userID)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		dbErr := errors.New("some other db error")

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1).
			WillReturnError(dbErr) // Simulate a generic DB error

		user, err := repo.GetUser(userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_GetAllTeachers(t *testing.T) {
	// GORM's Find doesn't usually add ORDER BY unless specified,
	// matching the behavior fixed in courseRepo tests.
	expectedSql := quoteSql(`SELECT * FROM "user" WHERE type_id = $1`)
	columns := []string{"user_id", "name", "lastname", "email", "type_id"}
	teacherTypeID := 2

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		rows := sqlmock.NewRows(columns).
			AddRow("teacher1", "Alice", "Smith", "alice@example.com", teacherTypeID).
			AddRow("teacher2", "Bob", "Jones", "bob@example.com", teacherTypeID)

		mock.ExpectQuery(expectedSql).
			WithArgs(teacherTypeID).
			WillReturnRows(rows)

		teachers, err := repo.GetAllTeachers()

		assert.NoError(t, err)
		require.NotNil(t, teachers) // Find returns non-nil slice even if empty
		assert.Len(t, teachers, 2)
		assert.Equal(t, "Alice", teachers[0].Name)
		assert.Equal(t, "teacher2", teachers[1].UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		rows := sqlmock.NewRows(columns) // Empty rows

		mock.ExpectQuery(expectedSql).
			WithArgs(teacherTypeID).
			WillReturnRows(rows)

		teachers, err := repo.GetAllTeachers()

		assert.NoError(t, err)
		// Find returns an empty, non-nil slice when no rows match
		require.NotNil(t, teachers)
		assert.Len(t, teachers, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		dbErr := errors.New("db select error for teachers")

		mock.ExpectQuery(expectedSql).
			WithArgs(teacherTypeID).
			WillReturnError(dbErr)

		teachers, err := repo.GetAllTeachers()

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		// On error during Find/Scan, the slice is typically nil
		assert.Nil(t, teachers)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_GetAllStudents(t *testing.T) {
	// Same SQL structure as GetAllTeachers, different type_id
	expectedSql := quoteSql(`SELECT * FROM "user" WHERE type_id = $1`)
	columns := []string{"user_id", "name", "lastname", "email", "type_id"}
	studentTypeID := 3

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		rows := sqlmock.NewRows(columns).
			AddRow("student1", "Charlie", "Brown", "charlie@example.com", studentTypeID).
			AddRow("student2", "Diana", "Prince", "diana@example.com", studentTypeID)

		mock.ExpectQuery(expectedSql).
			WithArgs(studentTypeID).
			WillReturnRows(rows)

		students, err := repo.GetAllStudents()

		assert.NoError(t, err)
		require.NotNil(t, students)
		assert.Len(t, students, 2)
		assert.Equal(t, "Charlie", students[0].Name)
		assert.Equal(t, "student2", students[1].UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		rows := sqlmock.NewRows(columns) // Empty rows

		mock.ExpectQuery(expectedSql).
			WithArgs(studentTypeID).
			WillReturnRows(rows)

		students, err := repo.GetAllStudents()

		assert.NoError(t, err)
		require.NotNil(t, students) // Expect empty, non-nil slice
		assert.Len(t, students, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserRepo(gormDb)

		dbErr := errors.New("db select error for students")

		mock.ExpectQuery(expectedSql).
			WithArgs(studentTypeID).
			WillReturnError(dbErr)

		students, err := repo.GetAllStudents()

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, students) // Expect nil slice on error
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
