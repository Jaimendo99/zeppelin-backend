package test_test

import (
	"errors"
	"testing"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			TypeID:   2,
		}
		rows := sqlmock.NewRows(columns).
			AddRow(expectedUser.UserID, expectedUser.Name, expectedUser.Lastname, expectedUser.Email, expectedUser.TypeID)

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1).
			WillReturnRows(rows)

		mock.ExpectQuery(`SELECT \* FROM "parental_consents" WHERE "parental_consents"."user_id" = \$1`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"consent_id", "user_id", "token", "status", "ip_address", "user_agent", "responded_at", "created_at"}))

		mock.ExpectQuery(`SELECT \* FROM "representatives" WHERE "representatives"."user_id" = \$1`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"representative_id", "name", "lastname", "email", "phone_number", "user_id"}))

		user, err := repo.GetUser(userID)

		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser.UserID, user.UserID)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Email, user.Email)
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

		mock.ExpectQuery(`SELECT \* FROM "parental_consents" WHERE "parental_consents"."user_id" IN \(\$1,\$2\)`).
			WithArgs("student1", "student2").
			WillReturnRows(sqlmock.NewRows([]string{
				"consent_id", "user_id", "status", "responded_at",
			}))

		mock.ExpectQuery(`SELECT \* FROM "representatives" WHERE "representatives"."user_id" IN \(\$1,\$2\)`).
			WithArgs("student1", "student2").
			WillReturnRows(sqlmock.NewRows([]string{
				"representative_id", "name", "lastname", "email", "phone_number", "user_id",
			}))

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

		rows := sqlmock.NewRows(columns)

		mock.ExpectQuery(expectedSql).
			WithArgs(studentTypeID).
			WillReturnRows(rows)

		students, err := repo.GetAllStudents()

		assert.NoError(t, err)
		require.NotNil(t, students)
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
		assert.Nil(t, students)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
