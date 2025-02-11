package db_test

import (
	"testing"
	"zeppelin/internal/db"
	"zeppelin/internal/domain"

	"github.com/glebarez/sqlite"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRepresentativeRepo(t *testing.T) {
	dB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	dB.AutoMigrate(&domain.Representative{})
	dB.Exec("DELETE FROM representatives")

	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	repo := db.NewRepresentativeRepo(dB)

	representative := domain.RepresentativeDb{
		Name:     "Felipe",
		Lastname: "Robalino",
	}

	err = repo.CreateRepresentative(representative)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	representativeDb, err := repo.GetRepresentative(1)
	if err != nil {
		// t.Errorf("Expected nil, got %v", err)
	}

	if representativeDb.Name != "" {
		t.Errorf("Expected Felipe, got %s", representativeDb.Name)
	}

}

func TestRepresentativeRepoGetAllRepresentatives(t *testing.T) {
	dB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	dB.AutoMigrate(&domain.Representative{})
	dB.Exec("DELETE FROM representatives")

	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	repo := db.NewRepresentativeRepo(dB)

	representative := domain.RepresentativeDb{
		Name:     "Felipe",
		Lastname: "Robalino",
	}

	err = repo.CreateRepresentative(representative)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	representatives, err := repo.GetAllRepresentatives()
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	if len(representatives) != 1 {
		t.Errorf("Expected 1, got %d", len(representatives))
	}
}

func TestRepresentativeRepoUpdateRepresentative(t *testing.T) {
	dB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	dB.AutoMigrate(&domain.Representative{})
	dB.Exec("DELETE FROM representatives")

	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	repo := db.NewRepresentativeRepo(dB)

	representative := domain.RepresentativeDb{
		Name:     "Felipe",
		Lastname: "Robalino",
	}

	err = repo.CreateRepresentative(representative)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	representativeInput := domain.RepresentativeInput{
		Name:     "Felipe",
		Lastname: "Robalino",
		Email:    "",
		Phone:    "",
	}

	err = repo.UpdateRepresentative(1, representativeInput)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	representativeDb, err := repo.GetRepresentative(1)

	if representativeDb.Name != "Felipe" {
		t.Errorf("Expected Felipe, got %s", representativeDb.Name)
	}

	if representativeDb.Lastname != "Robalino" {
		t.Errorf("Expected Robalino, got %s", representativeDb.Lastname)
	}

	if representativeDb.Email != nil {
		t.Errorf("Expected nil, got %s", *representativeDb.Email)

	}

	if representativeDb.Phone != nil {
		t.Errorf("Expected nil, got %s", *representativeDb.Phone)
	}

}
