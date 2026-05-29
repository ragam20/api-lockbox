package repository

import (
	"context"
	"errors"
	"time"

	"lockbox/internal/app/auth"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type authRepository struct {
	BaseRepository
}

func NewAuthRepository(db *gorm.DB) auth.AuthRepository {
	return &authRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *authRepository) WithTx(tx *gorm.DB) auth.AuthRepository {
	return &authRepository{BaseRepository: NewBaseRepository(tx)}
}

func (r *authRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.DB(ctx).
		Model(&entity.UserEntity{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *authRepository) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	var e entity.UserEntity
	if err := r.DB(ctx).Where("email = ?", email).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, err
	}

	return e.ToDomainUser(), nil
}

func (r *authRepository) CreateUser(ctx context.Context, user auth.User) error {
	userEntity := entity.UserEntity{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		EmailVerified:    user.EmailVerified,
		MasterKeySalt:    user.MasterKeySalt,
		AuthVerifier:     user.AuthVerifier,
		AuthVerifierSalt: user.AuthVerifierSalt,
	}

	return r.DB(ctx).Create(&userEntity).Error
}

func (r *authRepository) RevokeSession(ctx context.Context, userID int64, tokenHash string) error {
	return r.DB(ctx).Where("refresh_token_hash = ? AND user_id = ?", tokenHash, userID).
		Updates(map[string]any{
			"revoke_at": time.Now(),
		}).Error
}

func (r *authRepository) CreateSession(ctx context.Context, userID int64, userSession auth.UserSession) error {
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

func (r *authRepository) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (auth.UserSession, error) {
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
		RevokeAt:         e.RevokeAt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}, nil
}

func (r *authRepository) UpdateRefreshToken(ctx context.Context, sessionID int64, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
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
