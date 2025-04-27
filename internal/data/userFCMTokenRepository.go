package data

import (
	"zeppelin/internal/domain"

	"gorm.io/gorm"
)

type userFcmTokenRepo struct {
	db *gorm.DB
}

func NewUserFcmTokenRepo(db *gorm.DB) domain.UserFcmTokenRepo {
	return &userFcmTokenRepo{db: db}
}

func (r *userFcmTokenRepo) CreateUserFcmToken(token domain.UserFcmTokenDb) error {
	result := r.db.Create(&token)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *userFcmTokenRepo) GetUserFcmTokensByUserID(userID string) ([]domain.UserFcmTokenDb, error) {
	var tokens []domain.UserFcmTokenDb
	result := r.db.Where("user_id = ?", userID).Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

func (r *userFcmTokenRepo) DeleteUserFcmTokenByToken(firebaseToken string) error {
	result := r.db.Where("firebase_token = ?", firebaseToken).Delete(&domain.UserFcmTokenDb{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *userFcmTokenRepo) UpdateDeviceInfo(firebaseToken string, deviceInfo string) error {
	result := r.db.Model(&domain.UserFcmTokenDb{}).Where("firebase_token = ?", firebaseToken).Update("device_info", deviceInfo)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *userFcmTokenRepo) UpdateFirebaseToken(userID, deviceType, newToken string) error {
	result := r.db.Model(&domain.UserFcmTokenDb{}).
		Where("user_id = ? AND device_type = ?", userID, deviceType).
		Update("firebase_token", newToken)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
