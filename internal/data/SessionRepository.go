package data

import (
	"errors"
	"gorm.io/gorm"
	"time"
	"zeppelin/internal/domain" // Asegúrate de que la ruta de importación sea correcta
)

type sessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) domain.SessionRepo {
	return &sessionRepo{
		db: db,
	}
}

func (r *sessionRepo) StartSession(userID string) (int, error) {
	session := domain.Session{
		UserID: userID,
		Start:  time.Now(),
		End:    nil,
	}

	if err := r.db.Create(&session).Error; err != nil {
		return 0, err
	}
	return session.SessionID, nil
}

func (r *sessionRepo) EndSession(sessionID int) error {
	result := r.db.Model(&domain.Session{}).
		Where("session_id = ?", sessionID).
		Update("end", time.Now())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("session not found or already ended")
	}
	return nil
}

// GetActiveSessionByUserID obtiene la sesión activa (sin fecha de fin) de un usuario
func (r *sessionRepo) GetActiveSessionByUserID(userID string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.Where("user_id = ? AND \"end\" IS NULL", userID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Devolver nil, nil si no se encuentra una sesión activa
		}
		return nil, err
	}
	return &session, nil // Devolver un puntero a la sesión activa encontrada
}
