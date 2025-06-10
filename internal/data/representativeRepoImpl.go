package data

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

func (r *represetativeRepo) CreateRepresentative(representative domain.RepresentativeDb) (int, error) {
	result := r.db.Create(&representative)
	if result.Error != nil {
		return 0, result.Error
	}
	return representative.RepresentativeId, nil
}

func (r *represetativeRepo) GetRepresentative(representativeId int) (*domain.Representative, error) {
	if representativeId <= 0 {
		return nil, gorm.ErrInvalidData
	}
	var representative domain.Representative
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
	if representativeId <= 0 {
		return gorm.ErrInvalidData
	}

	updates := map[string]interface{}{
		"name":         representative.Name,
		"lastname":     representative.Lastname,
		"email":        sql.NullString{String: representative.Email, Valid: representative.Email != ""},
		"phone_number": sql.NullString{String: representative.PhoneNumber, Valid: representative.PhoneNumber != ""},
	}

	result := r.db.Model(&domain.RepresentativeDb{}).
		Where("representative_id = ?", representativeId).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
