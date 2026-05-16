package repository

import (
	"context"

	"lockbox/internal/app/auth"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type userSessionRepository struct {
	BaseRepository
}

func NewUserSessionRepository(db *gorm.DB) auth.UserSessionRepository {
	return &userSessionRepository{BaseRepository{db: db}}
}

func (r *userSessionRepository) CreateSession(ctx context.Context, userID int64, userSession auth.UserSession) error {
	e := entity.UserSessionEntity{
		ID:               userSession.ID,
		UserID:           userID,
		Device:           userSession.Device,
		RefreshTokenHash: userSession.RefreshTokenHash,
		UserAgent:        userSession.UserAgent,
		IPAddress:        userSession.IPAddress,
		ExpiresAt:        userSession.ExpiresAt,
	}

	if err := r.DB(ctx).Create(e).Error; err != nil {
		return err
	}

	return nil
}
