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

// Asumimos que setupMockDb y quoteSql ya existen en tu proyecto

func TestUserFcmTokenRepo_CreateUserFcmToken(t *testing.T) {
	token := domain.UserFcmTokenDb{
		UserID:        "user123",
		FirebaseToken: "firebase_token_abc",
		DeviceType:    "MOBILE",
		DeviceInfo:    "iPhone 14",
	}

	expectedSql := `INSERT INTO "user_fcm_token" ("user_id","firebase_token","device_type","device_info","updated_at") VALUES ($1,$2,$3,$4,$5) RETURNING "token_id"`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectQuery(quoteSql(expectedSql)).
			WithArgs(token.UserID, token.FirebaseToken, token.DeviceType, token.DeviceInfo, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"token_id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.CreateUserFcmToken(token)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectQuery(quoteSql(expectedSql)).
			WithArgs(token.UserID, token.FirebaseToken, token.DeviceType, token.DeviceInfo, sqlmock.AnyArg()).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.CreateUserFcmToken(token)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserFcmTokenRepo_GetUserFcmTokensByUserID(t *testing.T) {
	userID := "user123"
	expectedSql := quoteSql(`SELECT * FROM "user_fcm_token" WHERE user_id = $1`)
	columns := []string{"token_id", "user_id", "firebase_token", "device_type", "device_info", "updated_at"}

	t.Run("Success - Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		rows := sqlmock.NewRows(columns).
			AddRow(1, userID, "token1", "MOBILE", "iPhone 14", "2025-04-27T10:00:00Z").
			AddRow(2, userID, "token2", "WEB", "Chrome", "2025-04-27T11:00:00Z")

		mock.ExpectQuery(expectedSql).
			WithArgs(userID).
			WillReturnRows(rows)

		tokens, err := repo.GetUserFcmTokensByUserID(userID)

		assert.NoError(t, err)
		require.NotNil(t, tokens)
		assert.Len(t, tokens, 2)
		assert.Equal(t, "token1", tokens[0].FirebaseToken)
		assert.Equal(t, "WEB", tokens[1].DeviceType)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		rows := sqlmock.NewRows(columns) // Empty

		mock.ExpectQuery(expectedSql).
			WithArgs(userID).
			WillReturnRows(rows)

		tokens, err := repo.GetUserFcmTokensByUserID(userID)

		assert.NoError(t, err)
		require.NotNil(t, tokens)
		assert.Len(t, tokens, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		dbErr := errors.New("db select error")

		mock.ExpectQuery(expectedSql).
			WithArgs(userID).
			WillReturnError(dbErr)

		tokens, err := repo.GetUserFcmTokensByUserID(userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, tokens)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserFcmTokenRepo_DeleteUserFcmTokenByToken(t *testing.T) {
	firebaseToken := "firebase_token_abc"
	expectedSql := quoteSql(`DELETE FROM "user_fcm_token" WHERE firebase_token = $1`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(firebaseToken).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		mock.ExpectCommit()

		err := repo.DeleteUserFcmTokenByToken(firebaseToken)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		dbErr := errors.New("db delete error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(firebaseToken).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.DeleteUserFcmTokenByToken(firebaseToken)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserFcmTokenRepo_UpdateDeviceInfo(t *testing.T) {
	firebaseToken := "firebase_token_abc"
	deviceInfo := "Updated Device Info"
	expectedSql := quoteSql(`UPDATE "user_fcm_token" SET "device_info"=$1 WHERE firebase_token = $2`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(deviceInfo, firebaseToken).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		mock.ExpectCommit()

		err := repo.UpdateDeviceInfo(firebaseToken, deviceInfo)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(deviceInfo, firebaseToken).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateDeviceInfo(firebaseToken, deviceInfo)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserFcmTokenRepo_UpdateFirebaseToken(t *testing.T) {
	userID := "user123"
	deviceType := "MOBILE"
	newFirebaseToken := "new_firebase_token_xyz"
	expectedSql := quoteSql(`UPDATE "user_fcm_token" SET "firebase_token"=$1 WHERE user_id = $2 AND device_type = $3`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(newFirebaseToken, userID, deviceType).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		mock.ExpectCommit()

		err := repo.UpdateFirebaseToken(userID, deviceType, newFirebaseToken)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(newFirebaseToken, userID, deviceType).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateFirebaseToken(userID, deviceType, newFirebaseToken)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
