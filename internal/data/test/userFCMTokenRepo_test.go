package test_test

import (
	"errors"
	"testing"
	"time"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFcmTokenRepo_CreateUserFcmToken(t *testing.T) {
	now := time.Now()

	token := domain.UserFcmTokenDb{
		UserID:        "user123",
		FirebaseToken: "firebase_token_abc",
		DeviceType:    "MOBILE",
		DeviceInfo:    "iPhone 14",
		UpdatedAt:     now,
	}

	expectedSql := quoteSql(`INSERT INTO "user_fcm_token" ("user_id","firebase_token","device_type","device_info","updated_at") VALUES ($1,$2,$3,$4,$5) RETURNING "updated_at","token_id"`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(token.UserID, token.FirebaseToken, token.DeviceType, token.DeviceInfo, token.UpdatedAt).
			WillReturnRows(sqlmock.NewRows([]string{"updated_at", "token_id"}).AddRow(now, 1))
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
		mock.ExpectQuery(expectedSql).
			WithArgs(token.UserID, token.FirebaseToken, token.DeviceType, token.DeviceInfo, token.UpdatedAt).
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
			AddRow(1, userID, "token1", "MOBILE", "iPhone 14", time.Now()).
			AddRow(2, userID, "token2", "WEB", "Chrome", time.Now())

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

		rows := sqlmock.NewRows(columns)

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
			WillReturnResult(sqlmock.NewResult(0, 1))
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

	t.Run("No Rows Affected", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(firebaseToken).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteUserFcmTokenByToken(firebaseToken)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserFcmTokenRepo_UpdateDeviceInfo(t *testing.T) {
	firebaseToken := "firebase_token_abc"
	deviceInfo := "Updated Device Info"
	expectedSql := quoteSql(`UPDATE "user_fcm_token" SET "device_info"=$1,"updated_at"=$2 WHERE firebase_token = $3`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(deviceInfo, sqlmock.AnyArg(), firebaseToken).
			WillReturnResult(sqlmock.NewResult(0, 1))
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
			WithArgs(deviceInfo, sqlmock.AnyArg(), firebaseToken).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateDeviceInfo(firebaseToken, deviceInfo)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No Rows Updated", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(deviceInfo, sqlmock.AnyArg(), firebaseToken).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateDeviceInfo(firebaseToken, deviceInfo)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserFcmTokenRepo_UpdateFirebaseToken(t *testing.T) {
	userID := "user123"
	deviceType := "MOBILE"
	newFirebaseToken := "new_firebase_token_xyz"
	expectedSql := quoteSql(`UPDATE "user_fcm_token" SET "firebase_token"=$1,"updated_at"=$2 WHERE user_id = $3 AND device_type = $4`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(newFirebaseToken, sqlmock.AnyArg(), userID, deviceType).
			WillReturnResult(sqlmock.NewResult(0, 1))
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
			WithArgs(newFirebaseToken, sqlmock.AnyArg(), userID, deviceType).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateFirebaseToken(userID, deviceType, newFirebaseToken)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No Rows Updated", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewUserFcmTokenRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(newFirebaseToken, sqlmock.AnyArg(), userID, deviceType).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateFirebaseToken(userID, deviceType, newFirebaseToken)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
