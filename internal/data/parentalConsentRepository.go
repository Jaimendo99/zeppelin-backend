package data

import (
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type parentalConsentRepo struct {
	db *gorm.DB
}

func NewParentalConsentRepo(db *gorm.DB) domain.ParentalConsentRepo {
	return &parentalConsentRepo{db}
}

func (r *parentalConsentRepo) CreateConsent(consent domain.ParentalConsent) error {
	return r.db.Table("parental_consents").Create(&consent).Error
}

func (r *parentalConsentRepo) UpdateConsentStatus(token, status, ip, userAgent string) error {
	return r.db.Table("parental_consents").
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"status":       status,
			"ip_address":   ip,
			"user_agent":   userAgent,
			"responded_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *parentalConsentRepo) GetConsentByToken(token string) (*domain.ParentalConsent, error) {
	var consent domain.ParentalConsent
	if err := r.db.Table("parental_consents").
		Where("token = ?", token).
		First(&consent).Error; err != nil {
		return nil, err
	}
	return &consent, nil
}

func (r *parentalConsentRepo) GetConsentByUserID(userID string) (*domain.ParentalConsent, error) {
	var consent domain.ParentalConsent
	err := r.db.Table("parental_consents").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&consent).Error
	if err != nil {
		return nil, err
	}
	return &consent, nil
}
