package db_test

import (
	"testing"
	"zeppelin/internal/db"
	"zeppelin/internal/domain"

	"gorm.io/driver/sqlite"
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
