package test_test

import (
	"database/sql"
	"errors"
	"testing"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRepresentativeRepo_CreateRepresentative(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewRepresentativeRepo(gormDb)

	representative := domain.RepresentativeDb{
		Name:        "John",
		Lastname:    "Doe",
		Email:       "john.doe@example.com",
		PhoneNumber: "123456789",
		UserID:      "", // si es vacío está bien
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		// La consulta debe incluir los 5 campos
		mock.ExpectQuery(`INSERT INTO "representatives"`).
			WithArgs(
				representative.Name,
				representative.Lastname,
				representative.Email,
				representative.PhoneNumber,
				representative.UserID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"representative_id"}).AddRow(1))
		mock.ExpectCommit()

		id, err := repo.CreateRepresentative(representative)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "representatives"`).
			WithArgs(
				representative.Name,
				representative.Lastname,
				representative.Email,
				representative.PhoneNumber,
				representative.UserID,
			).
			WillReturnError(expectedErr)
		mock.ExpectRollback()

		id, err := repo.CreateRepresentative(representative)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepresentativeRepo_GetRepresentative(t *testing.T) {
	// Define the expected SQL pattern once
	// Use quoteSql to handle special characters correctly
	expectedSql := quoteSql(`SELECT * FROM "representatives" WHERE representative_id = $1 ORDER BY "representatives"."representative_id" LIMIT $2`)
	representativeId := 1

	t.Run("Success", func(t *testing.T) {
		// Create mock inside t.Run for isolation
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)

		rows := sqlmock.NewRows([]string{"representative_id", "name", "lastname", "email", "phone_number"}).
			AddRow(representativeId, "John", "Doe", "john.doe@example.com", "123456789")

		// Use the corrected expected SQL pattern
		mock.ExpectQuery(expectedSql).
			WithArgs(representativeId, 1). // GORM adds LIMIT 1 for First()
			WillReturnRows(rows)

		representative, err := repo.GetRepresentative(representativeId)

		// Assertions remain the same
		assert.NoError(t, err)
		require.NotNil(t, representative)
		assert.Equal(t, representativeId, representative.RepresentativeId) // Also check ID
		assert.Equal(t, "John", representative.Name)
		assert.Equal(t, "Doe", representative.Lastname)
		// Handle potential nulls if GetRepresentative returns RepresentativeDb or similar
		// Assuming GetRepresentative returns domain.Representative as refactored previously
		assert.Equal(t, "john.doe@example.com", representative.Email)
		assert.Equal(t, "123456789", representative.PhoneNumber)
		assert.NoError(t, mock.ExpectationsWereMet()) // Check mock at the end
	})

	t.Run("Not Found", func(t *testing.T) {
		// Create mock inside t.Run for isolation
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)

		// Use the corrected expected SQL pattern
		mock.ExpectQuery(expectedSql).
			WithArgs(representativeId, 1).
			WillReturnError(gorm.ErrRecordNotFound) // Simulate DB error

		representative, err := repo.GetRepresentative(representativeId)

		// Assertions remain the same
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Nil(t, representative)
		assert.NoError(t, mock.ExpectationsWereMet()) // Check mock at the end
	})

	t.Run("Invalid ID", func(t *testing.T) {
		// Create mock inside t.Run for isolation (even though it shouldn't be hit)
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)

		// This test case should not interact with the mock DB
		representative, err := repo.GetRepresentative(-1) // Or 0 depending on your check

		// Assertions remain the same
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidData, err) // Or your specific invalid ID error
		assert.Nil(t, representative)
		// Verify that NO expectations were made on the mock for this specific run
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepresentativeRepo_GetAllRepresentatives(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewRepresentativeRepo(gormDb)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"representative_id", "name", "lastname", "email", "phone_number"}).
			AddRow(1, "John", "Doe", "john.doe@example.com", "123456789").
			AddRow(2, "Jane", "Smith", "jane.smith@example.com", "987654321")

		// Use the actual table name
		mock.ExpectQuery(`SELECT \* FROM "representatives"`).
			WillReturnRows(rows)

		representatives, err := repo.GetAllRepresentatives()

		assert.NoError(t, err)
		require.NotNil(t, representatives)
		assert.Len(t, representatives, 2)
		assert.Equal(t, "John", representatives[0].Name)
		assert.Equal(t, "Jane", representatives[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty Result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"representative_id", "name", "lastname", "email", "phone_number"})

		mock.ExpectQuery(`SELECT \* FROM "representatives"`).
			WillReturnRows(rows)

		representatives, err := repo.GetAllRepresentatives()

		assert.NoError(t, err)
		assert.Len(t, representatives, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery(`SELECT \* FROM "representatives"`).
			WillReturnError(expectedErr)

		representatives, err := repo.GetAllRepresentatives()

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, representatives)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
func TestRepresentativeRepo_UpdateRepresentative(t *testing.T) {
	representativeId := 1
	representative := domain.RepresentativeInput{
		Name:        "John Updated",
		Lastname:    "Doe Updated",
		Email:       "john.updated@example.com",
		PhoneNumber: "999999999",
	}

	// The exact SQL GORM generates (based on the error message)
	// Note: The order of SET columns might vary based on map iteration order in Go,
	// but GORM is often consistent. If tests become flaky, using .* might be needed again,
	// but ensuring WithArgs order is correct is crucial.
	// Let's stick with the order observed in the error for now.
	expectedSql := `UPDATE "representatives" SET "email"=$1,"lastname"=$2,"name"=$3,"phone_number"=$4 WHERE representative_id = $5`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)

		mock.ExpectBegin()
		// Use the quoted exact SQL string
		mock.ExpectExec(quoteSql(expectedSql)).
			WithArgs( // Correct argument order based on the actual query ($1, $2, $3, $4, $5)
				sql.NullString{String: representative.Email, Valid: true},       // $1 = email
				representative.Lastname,                                         // $2 = lastname
				representative.Name,                                             // $3 = name
				sql.NullString{String: representative.PhoneNumber, Valid: true}, // $4 = phone_number
				representativeId,                                                // $5 = representative_id
			).
			WillReturnResult(sqlmock.NewResult(1, 1)) // RowsAffected = 1
		mock.ExpectCommit()

		err := repo.UpdateRepresentative(representativeId, representative)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not Found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)

		mock.ExpectBegin()
		// Expect the same exact UPDATE statement
		mock.ExpectExec(quoteSql(expectedSql)).
			WithArgs( // Correct argument order
				sql.NullString{String: representative.Email, Valid: true},       // $1
				representative.Lastname,                                         // $2
				representative.Name,                                             // $3
				sql.NullString{String: representative.PhoneNumber, Valid: true}, // $4
				representativeId,                                                // $5
			).
			WillReturnResult(sqlmock.NewResult(0, 0)) // Simulate 0 rows affected
		mock.ExpectCommit() // GORM usually commits even if RowsAffected is 0

		err := repo.UpdateRepresentative(representativeId, representative)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err) // Expect ErrRecordNotFound due to RowsAffected check
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid ID", func(t *testing.T) {
		// This test remains the same as it doesn't hit the DB mock
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)
		err := repo.UpdateRepresentative(0, representative)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidData, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update Failure", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewRepresentativeRepo(gormDb)

		expectedErr := errors.New("database error") // The specific DB error

		mock.ExpectBegin()
		// Expect the exact UPDATE statement
		mock.ExpectExec(quoteSql(expectedSql)).
			WithArgs( // Correct argument order
				sql.NullString{String: representative.Email, Valid: true},       // $1
				representative.Lastname,                                         // $2
				representative.Name,                                             // $3
				sql.NullString{String: representative.PhoneNumber, Valid: true}, // $4
				representativeId,                                                // $5
			).
			WillReturnError(expectedErr) // Simulate a database error
		mock.ExpectRollback() // Expect rollback on error

		err := repo.UpdateRepresentative(representativeId, representative)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err) // The error should be the one returned by the DB
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
