package domain

import "time"

type ParentalConsent struct {
	ConsentID        int        `json:"consent_id"`
	UserID           string     `json:"user_id"`
	RepresentativeID int        `json:"representative_id"`
	Token            string     `json:"token"`
	Status           string     `json:"status"`
	IPAddress        string     `json:"ip_address"`
	UserAgent        string     `json:"user_agent"`
	RespondedAt      *time.Time `json:"responded_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ParentalConsentDb struct {
	ConsentID   int    `gorm:"primaryKey"`
	UserID      string `gorm:"column:user_id"` // Foreign key hacia user
	Status      string
	RespondedAt *time.Time
}

type ParentalConsentRepo interface {
	CreateConsent(consent ParentalConsent) error
	UpdateConsentStatus(token, status, ip, userAgent string) error
	GetConsentByToken(token string) (*ParentalConsent, error)
	GetConsentByUserID(userID string) (*ParentalConsent, error)
}

func (ParentalConsentDb) TableName() string {
	return "parental_consents"
}
