package domain

import "time"

type UserFcmTokenDb struct {
	TokenID       int       `json:"token_id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"not null"`
	FirebaseToken string    `json:"firebase_token" gorm:"not null"`
	DeviceType    string    `json:"device_type" gorm:"not null"` // WEB o MOBILE
	DeviceInfo    string    `json:"device_info"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

type UserFcmTokenDeleteInput struct {
	FirebaseToken string `json:"firebase_token" validate:"required"`
}

type UserFcmTokenUpdateDeviceInput struct {
	FirebaseToken string `json:"firebase_token" validate:"required"`
	DeviceInfo    string `json:"device_info" validate:"required"`
}

type UserFcmTokenInput struct {
	FirebaseToken string `json:"firebase_token" validate:"required"`
	DeviceType    string `json:"device_type" validate:"required,oneof=WEB MOBILE"`
	DeviceInfo    string `json:"device_info"`
}

type UserFcmTokenRepo interface {
	CreateUserFcmToken(token UserFcmTokenDb) error
	GetUserFcmTokensByUserID(userID string) ([]UserFcmTokenDb, error)
	DeleteUserFcmTokenByToken(firebaseToken string) error
	UpdateDeviceInfo(firebaseToken string, deviceInfo string) error
	UpdateFirebaseToken(userID string, deviceType string, firebaseToken string) error
}

func (UserFcmTokenDb) TableName() string {
	return "user_fcm_token"
}
