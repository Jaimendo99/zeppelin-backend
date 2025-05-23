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
)

func TestSessionRepo_StartSession(t *testing.T) {
	userID := "user_start_123"
	// Adjust the regex to match the exact structure from the error message
	// Use regexp.QuoteMeta for the literal parts, and manually add regex for placeholders and RETURNING
	expectedInsertRegex := regexp.MustCompile(
		regexp.QuoteMeta(`INSERT INTO "session" ("user_id","start","end") VALUES `) + // Quote the literal part
			`\(\$1,\$2,\$3\)` + // Match the ($1,$2,$3) part
			regexp.QuoteMeta(` RETURNING "session_id"`), // Quote the literal part
	)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		expectedSessionID := 1
		mock.ExpectBegin()
		// Use the updated regex
		mock.ExpectQuery(expectedInsertRegex.String()).
			WithArgs(userID, sqlmock.AnyArg(), nil).
			WillReturnRows(sqlmock.NewRows([]string{"session_id"}).AddRow(expectedSessionID))
		mock.ExpectCommit()

		sessionID, err := repo.StartSession(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedSessionID, sessionID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		// Use the updated regex
		mock.ExpectQuery(expectedInsertRegex.String()).
			WithArgs(userID, sqlmock.AnyArg(), nil).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		sessionID, err := repo.StartSession(userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, 0, sessionID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Keep the TestSessionRepo_EndSession and TestSessionRepo_GetActiveSessionByUserID as they were in the last correction,
// as the errors in the most recent output were only for TestSessionRepo_StartSession.
// If they start failing again after this change, we'll revisit their regexes based on new error messages.

func TestSessionRepo_EndSession(t *testing.T) {
	sessionID := 1
	// Correctly escape backslashes in the Go string literal for the regex
	expectedUpdateRegex := regexp.MustCompile(`UPDATE "session" SET "end"=\$1 WHERE session_id = \$2`)

	t.Run("Success - Session Found and Updated", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedUpdateRegex.String()).
			WithArgs(sqlmock.AnyArg(), sessionID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.EndSession(sessionID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Session Not Found or Already Ended", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedUpdateRegex.String()).
			WithArgs(sqlmock.AnyArg(), sessionID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.EndSession(sessionID)

		assert.Error(t, err)
		assert.Equal(t, "session not found or already ended", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedUpdateRegex.String()).
			WithArgs(sqlmock.AnyArg(), sessionID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.EndSession(sessionID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSessionRepo_GetActiveSessionByUserID(t *testing.T) {
	userID := "user_active_456"
	// Correctly escape backslashes and include the table name in ORDER BY
	expectedSelectRegex := regexp.MustCompile(`SELECT \* FROM "session" WHERE user_id = \$1 AND "end" IS NULL ORDER BY "session"."session_id" LIMIT \$2`)

	t.Run("Success - Active Session Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		sessionRows := sqlmock.NewRows([]string{"session_id", "user_id", "start", "end"}).
			AddRow(1, userID, time.Now(), nil)
		mock.ExpectQuery(expectedSelectRegex.String()).
			WithArgs(userID, 1).
			WillReturnRows(sessionRows)

		session, err := repo.GetActiveSessionByUserID(userID)

		assert.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, 1, session.SessionID)
		assert.Equal(t, userID, session.UserID)
		assert.Nil(t, session.End)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - No Active Session Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		mock.ExpectQuery(expectedSelectRegex.String()).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		session, err := repo.GetActiveSessionByUserID(userID)

		assert.NoError(t, err)
		assert.Nil(t, session)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewSessionRepo(gormDb)

		dbErr := errors.New("db select error")
		mock.ExpectQuery(expectedSelectRegex.String()).
			WithArgs(userID, 1).
			WillReturnError(dbErr)

		session, err := repo.GetActiveSessionByUserID(userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, session)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
