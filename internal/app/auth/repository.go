package auth

import (
	"context"
	"time"
)

type AuthRepository interface {
	WithTransaction(ctx context.Context, fn func(repo AuthRepository) error) error
	GetByEmail(ctx context.Context, email string) (User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user User) error
	RevokeSession(ctx context.Context, userID int64, tokenHash string) error
	CreateSession(ctx context.Context, userID int64, userSession UserSession) error
	GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (UserSession, error)
	UpdateRefreshToken(ctx context.Context, sessionID int64, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error
}
