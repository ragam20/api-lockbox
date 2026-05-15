package auth

import "context"

type AuthRepository interface {
	WithTransaction(ctx context.Context, fn func(repo AuthRepository) error) error
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user User) error
	CreateArgonConfig(ctx context.Context, argonConfig ArgonConfig) error
	CreateVault(ctx context.Context, vault Vault) error
}
