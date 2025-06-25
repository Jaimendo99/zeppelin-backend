package test_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
)

func TestParentalConsentRepo(t *testing.T) {
	t.Run("CreateConsent - Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewParentalConsentRepo(gormDb)

		consent := domain.ParentalConsent{
			UserID:      "user123",
			Token:       "abc-token",
			Status:      "PENDING",
			IPAddress:   "192.168.0.1",
			UserAgent:   "Mozilla/5.0",
			RespondedAt: nil,
			CreatedAt:   time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "parental_consents"`).
			WithArgs(sqlmock.AnyArg(), "user123", 0, "abc-token", "PENDING", "192.168.0.1", "Mozilla/5.0", nil, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateConsent(consent)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateConsentStatus - Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewParentalConsentRepo(gormDb)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "parental_consents" SET "ip_address"=\$1,"responded_at"=NOW\(\),"status"=\$2,"user_agent"=\$3 WHERE token = \$4`).
			WithArgs("127.0.0.1", "ACCEPTED", "TestAgent", "abc-token").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateConsentStatus("abc-token", "ACCEPTED", "127.0.0.1", "TestAgent")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetConsentByToken - Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewParentalConsentRepo(gormDb)

		token := "abc-token"
		columns := []string{"consent_id", "user_id", "representative_id", "token", "status", "ip_address", "user_agent", "responded_at", "created_at"}
		createdAt := time.Now()

		mock.ExpectQuery(`SELECT \* FROM "parental_consents" WHERE token = \$1 ORDER BY "parental_consents"\."consent_id" LIMIT \$2`).
			WithArgs(token, 1).
			WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "user123", 0, token, "PENDING", "192.168.0.1", "Mozilla/5.0", nil, createdAt))

		result, err := repo.GetConsentByToken(token)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user123", result.UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetConsentByUserID - Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewParentalConsentRepo(gormDb)

		userID := "user123"
		columns := []string{"consent_id", "user_id", "representative_id", "token", "status", "ip_address", "user_agent", "responded_at", "created_at"}
		createdAt := time.Now()

		mock.ExpectQuery(`SELECT \* FROM "parental_consents" WHERE user_id = \$1 ORDER BY created_at DESC,"parental_consents"\."consent_id" LIMIT \$2`).
			WithArgs(userID, 1).
			WillReturnRows(sqlmock.NewRows(columns).AddRow(1, userID, 0, "abc-token", "PENDING", "192.168.0.1", "Mozilla/5.0", nil, createdAt))

		result, err := repo.GetConsentByUserID(userID)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
