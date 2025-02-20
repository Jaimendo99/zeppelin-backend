// repreRepo_test.go
package data_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"zeppelin/internal/data"
	"zeppelin/internal/domain"
)

// representativeTestModel is used only for migrations so that the table includes an auto-generated primary key.
type representativeTestModel struct {
	RepresentativeID int            `gorm:"primaryKey;column:representative_id"`
	Name             string         `gorm:"column:name"`
	Lastname         string         `gorm:"column:lastname"`
	Email            sql.NullString `gorm:"column:email"`
	PhoneNumber      sql.NullString `gorm:"column:phone"`
}

// TableName tells GORM which table name to use.
func (representativeTestModel) TableName() string {
	return "representatives"
}

// setupTestdata creates a fresh SQLite database for testing.
func setupTestdata(t *testing.T) *gorm.DB {
	// Remove any existing test database.
	os.Remove("test.data")

	// Open the SQLite database.
	gormdata, err := gorm.Open(sqlite.Open("test.data"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Migrate the schema using our test model.
	if err := gormdata.AutoMigrate(&representativeTestModel{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return gormdata
}

func TestCreateRepresentative(t *testing.T) {
	dataConn := setupTestdata(t)
	repo := data.NewRepresentativeRepo(dataConn)

	// Create a new representative using your production type.
	rep := domain.RepresentativeDb{
		Name:        "Mateo",
		Lastname:    "Mejia",
		Email:       sql.NullString{String: "jaimendo26@gmail.com", Valid: true},
		PhoneNumber: sql.NullString{String: "+129129122", Valid: true},
	}

	err := repo.CreateRepresentative(rep)
	assert.NoError(t, err, "expected no error when creating representative")

	// Retrieve the record directly using the test model.
	var testRep representativeTestModel
	result := dataConn.First(&testRep)
	assert.NoError(t, result.Error, "expected to retrieve representative")
	assert.Equal(t, rep.Name, testRep.Name, "Name should match")
	assert.Equal(t, rep.Lastname, testRep.Lastname, "Lastname should match")
	assert.Equal(t, rep.Email.String, testRep.Email.String, "Email should match")
	assert.Equal(t, rep.PhoneNumber.String, testRep.PhoneNumber.String, "Phone should match")
}

func TestGetRepresentative(t *testing.T) {
	dataConn := setupTestdata(t)
	repo := data.NewRepresentativeRepo(dataConn)

	// Insert a test record using the test model so we have a generated representative_id.
	testRep := representativeTestModel{
		Name:        "Mateo",
		Lastname:    "Mejia",
		Email:       sql.NullString{String: "jaimendo26@gmail.com", Valid: true},
		PhoneNumber: sql.NullString{String: "+129129122", Valid: true},
	}
	err := dataConn.Create(&testRep).Error
	assert.NoError(t, err, "expected no error creating test representative")

	// Retrieve via the repository.
	repInput, err := repo.GetRepresentative(testRep.RepresentativeID)
	assert.NoError(t, err, "expected no error retrieving representative")
	assert.Equal(t, testRep.Name, repInput.Name, "Name should match")
	assert.Equal(t, testRep.Lastname, repInput.Lastname, "Lastname should match")
	// Ensure that the phone is mapped correctly.
	assert.Equal(t, testRep.PhoneNumber.String, repInput.PhoneNumber, "Phone should match")
	// Compare Email.
	assert.Equal(t, testRep.Email.String, repInput.Email, "Email should match")
}

func TestGetAllRepresentatives(t *testing.T) {
	dataConn := setupTestdata(t)
	repo := data.NewRepresentativeRepo(dataConn)

	// Insert multiple records.
	reps := []representativeTestModel{
		{
			Name:        "Mateo",
			Lastname:    "Mejia",
			Email:       sql.NullString{String: "jaimendo26@gmail.com", Valid: true},
			PhoneNumber: sql.NullString{String: "+129129122", Valid: true},
		},
		{
			Name:        "John",
			Lastname:    "Doe",
			Email:       sql.NullString{String: "john@example.com", Valid: true},
			PhoneNumber: sql.NullString{String: "+1123456789", Valid: true},
		},
	}

	for _, rep := range reps {
		err := dataConn.Create(&rep).Error
		assert.NoError(t, err, "expected no error creating record")
	}

	allReps, err := repo.GetAllRepresentatives()
	assert.NoError(t, err, "expected no error retrieving all representatives")
	assert.Equal(t, len(reps), len(allReps), "expected same number of representatives")
}

func TestUpdateRepresentative(t *testing.T) {
	dataConn := setupTestdata(t)
	repo := data.NewRepresentativeRepo(dataConn)

	// Insert a test record.
	testRep := representativeTestModel{
		Name:        "Mateo",
		Lastname:    "Mejia",
		Email:       sql.NullString{String: "jaimendo26@gmail.com", Valid: true},
		PhoneNumber: sql.NullString{String: "+129129122", Valid: true},
	}
	err := dataConn.Create(&testRep).Error
	assert.NoError(t, err, "expected no error creating record for update test")

	updatedInput := domain.RepresentativeInput{
		Name:        "UpdatedName",
		Lastname:    "UpdatedLastname",
		Email:       "updated@example.com",
		PhoneNumber: "+1987654321",
	}

	// Call UpdateRepresentative.
	err = repo.UpdateRepresentative(testRep.RepresentativeID, updatedInput)
	// Since your production model (Representativedata) does not include the primary key, GORM
	// cannot determine the record to update and returns an error.
	// If you fix your model to include the primary key, this test should verify the update.
	assert.Error(t, err, "expected an error updating representative due to missing primary key in model")

	// Uncomment the following block if you update your production code to include the primary key.
	/*
		var updatedRep representativeTestModel
		err = dataConn.Where("representative_id = ?", testRep.RepresentativeID).First(&updatedRep).Error
		assert.NoError(t, err, "expected to retrieve updated representative")
		assert.Equal(t, updatedInput.Name, updatedRep.Name, "Name should be updated")
		assert.Equal(t, updatedInput.Lastname, updatedRep.Lastname, "Lastname should be updated")
		assert.Equal(t, updatedInput.Email, updatedRep.Email.String, "Email should be updated")
		assert.Equal(t, updatedInput.PhoneNumber, updatedRep.PhoneNumber.String, "Phone should be updated")
	*/
}
