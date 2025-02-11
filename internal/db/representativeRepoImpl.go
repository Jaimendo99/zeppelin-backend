package db

import (
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

func (r *represetativeRepo) GetRepresentative(representative_id int) (domain.RepresentativeDb, error) {
	var representative domain.RepresentativeDb
	result := r.db.First(&representative, representative_id)
	if result.Error != nil {
		return domain.RepresentativeDb{}, result.Error
	}
	return representative, nil
}
