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

func TestUserPomodoroRepo_GetByUserID(t *testing.T) {
	userID := "user123"
	// Updated to match actual SQL with pomodoro_id and LIMIT $2
	expectedSql := quoteSql(`SELECT * FROM "user_pomodoro" WHERE user_id = $1 ORDER BY "user_pomodoro"."pomodoro_id" LIMIT $2`)
	columns := []string{"pomodoro_id", "user_id", "active_time", "rest_time", "long_rest_time", "iterations"}

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserPomodoroRepo(gormDb)

		rows := sqlmock.NewRows(columns).
			AddRow(1, userID, 25, 5, 15, 4)

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1). // Add LIMIT parameter
			WillReturnRows(rows)

		pomodoro, err := repo.GetByUserID(userID)

		assert.NoError(t, err)
		require.NotNil(t, pomodoro)
		assert.Equal(t, userID, pomodoro.UserID)
		assert.Equal(t, 25, pomodoro.ActiveTime)
		assert.Equal(t, 5, pomodoro.RestTime)
		assert.Equal(t, 15, pomodoro.LongRestTime)
		assert.Equal(t, 4, pomodoro.Iterations)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserPomodoroRepo(gormDb)

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1). // Add LIMIT parameter
			WillReturnError(gorm.ErrRecordNotFound)

		pomodoro, err := repo.GetByUserID(userID)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.NotNil(t, pomodoro) // Returns empty struct
		assert.Empty(t, pomodoro.UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserPomodoroRepo(gormDb)

		dbErr := errors.New("db select error")

		mock.ExpectQuery(expectedSql).
			WithArgs(userID, 1). // Add LIMIT parameter
			WillReturnError(dbErr)

		pomodoro, err := repo.GetByUserID(userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NotNil(t, pomodoro) // Returns empty struct
		assert.Empty(t, pomodoro.UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserPomodoroRepo_UpdateByUserID(t *testing.T) {
	userID := "user123"
	input := domain.UpdatePomodoroInput{
		ActiveTime:   30,
		RestTime:     10,
		LongRestTime: 20,
		Iterations:   5,
	}
	// Updated to match actual column order
	expectedSql := quoteSql(`UPDATE "user_pomodoro" SET "active_time"=$1,"iterations"=$2,"long_rest_time"=$3,"rest_time"=$4 WHERE user_id = $5`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserPomodoroRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(input.ActiveTime, input.Iterations, input.LongRestTime, input.RestTime, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateByUserID(userID, input)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserPomodoroRepo(gormDb)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(input.ActiveTime, input.Iterations, input.LongRestTime, input.RestTime, userID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateByUserID(userID, input)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No Rows Updated", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserPomodoroRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(input.ActiveTime, input.Iterations, input.LongRestTime, input.RestTime, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateByUserID(userID, input)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
