package repository

import (
	"context"
	"time"

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
		LastUsedAt:       userSession.LastUsedAt,
	}

	if err := r.DB(ctx).Create(&e).Error; err != nil {
		return err
	}

	return nil
}

func (r *userSessionRepository) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (auth.UserSession, error) {
	var e entity.UserSessionEntity
	if err := r.DB(ctx).Where("refresh_token_hash = ?", refreshTokenHash).First(&e).Error; err != nil {
		return auth.UserSession{}, err
	}

	return auth.UserSession{
		ID:               e.ID,
		UserID:           e.UserID,
		Device:           e.Device,
		RefreshTokenHash: e.RefreshTokenHash,
		UserAgent:        e.UserAgent,
		IPAddress:        e.IPAddress,
		LastUsedAt:       e.LastUsedAt,
		ExpiresAt:        e.ExpiresAt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}, nil
}

func (r *userSessionRepository) UpdateRefreshToken(ctx context.Context, sessionID int64, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	result := r.DB(ctx).
		Model(&entity.UserSessionEntity{}).
		Where("id = ? AND refresh_token_hash = ?", sessionID, oldRefreshTokenHash).
		Updates(map[string]any{
			"refresh_token_hash": newRefreshTokenHash,
			"expires_at":         expiresAt,
			"last_used_at":       lastUsedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
