package auth

import "context"

type AuthRepository interface {
	WithTransaction(ctx context.Context, fn func(repo AuthRepository) error) error
	GetByEmail(ctx context.Context, email string) (User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user User) error
	CreateArgonConfig(ctx context.Context, argonConfig ArgonConfig) error
	CreateVault(ctx context.Context, vault Vault) error
}

type UserSessionRepository interface {
	CreateSession(ctx context.Context, userID int64, userSession UserSession) error
}
