package db

import (
	"fmt"
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type represetativeRepo struct {
	db *gorm.DB
}

func NewRepresentativeRepo(db *gorm.DB) domain.RepresentativeRepo {
	return &represetativeRepo{db: db}
}

func (r *represetativeRepo) CreateRepresentative(representative domain.RepresentativeDb) error {
	result := r.db.Create(&representative)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *represetativeRepo) GetRepresentative(representativeId int) (domain.RepresentativeDb, error) {
	var representative domain.RepresentativeDb
	result := r.db.Where("representative_id = ?", representativeId).First(&representative)
	if result.Error != nil {
		return domain.RepresentativeDb{}, result.Error
	}
	return representative, nil
}

func (r *represetativeRepo) GetAllRepresentatives() ([]domain.Representative, error) {
	var representatives []domain.Representative
	result := r.db.Find(&representatives)
	if result.Error != nil {
		return nil, result.Error
	}
	return representatives, nil
}

func (r *represetativeRepo) UpdateRepresentative(representativeId int, representative domain.RepresentativeInput) error {
	var representativeDb domain.Representative
	result := r.db.Where("representative_id = ?", representativeId).First(&representativeDb)
	if result.Error != nil {
		fmt.Printf("REPRE_REPO: %v  -1-   ", result.Error)
		return result.Error
	}
	representativeDb.Name = representative.Name
	representativeDb.Lastname = representative.Lastname
	representativeDb.Email = representative.Email
	representativeDb.Phone = representative.Phone
	result = r.db.Save(&representativeDb)
	if result.Error != nil {
		fmt.Printf("REPRE_REPO: %v  -2-   ", result.Error)
		return result.Error
	}
	return nil
}
