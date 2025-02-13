package db

import (
	"database/sql"
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

func (r *represetativeRepo) GetRepresentative(representativeId int) (*domain.RepresentativeInput, error) {
	if representativeId < 0 {
		return nil, gorm.ErrInvalidData
	}
	var representative domain.RepresentativeInput
	result := r.db.Where("representative_id = ?", representativeId).First(&representative)
	if result.Error != nil {
		return nil, result.Error
	}
	return &representative, nil
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
	if representativeId < 0 {
		return gorm.ErrInvalidData
	}
	var representativeDb domain.RepresentativeDb
	result := r.db.Where("representative_id = ?", representativeId).First(&representativeDb)
	if result.Error != nil {
		return result.Error
	}
	representativeDb.Name = representative.Name
	representativeDb.Lastname = representative.Lastname
	representativeDb.Email = sql.NullString{String: representative.Email}
	representativeDb.PhoneNumber = sql.NullString{String: representative.PhoneNumber}

	result = r.db.Save(&representativeDb)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
