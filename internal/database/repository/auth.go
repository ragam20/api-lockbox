package repository

import (
	"context"

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

func (r *authRepository) WithTransaction(ctx context.Context, fn func(repo auth.AuthRepository) error) error {
	return r.BaseRepository.WithTransaction(ctx, func(tx *gorm.DB) error {
		return fn(&authRepository{BaseRepository: NewBaseRepository(tx)})
	})
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

func (r *authRepository) CreateArgonConfig(ctx context.Context, argonConfig auth.ArgonConfig) error {
	argonEntity := entity.ArgonConfigEntity{
		ID:          argonConfig.ID,
		UserID:      argonConfig.UserID,
		Time:        argonConfig.Iterations,
		Memory:      argonConfig.Memory,
		Parallelism: argonConfig.Parallelism,
	}

	return r.DB(ctx).Create(&argonEntity).Error
}

func (r *authRepository) CreateVault(ctx context.Context, vault auth.Vault) error {
	vaultEntity := entity.VaultEntity{
		ID:           vault.ID,
		UserID:       vault.UserID,
		Timeout:      vault.Timeout,
		TimeoutType:  entity.VaultTimeoutType(vault.TimeoutType),
		VaultKeySalt: vault.VaultKeySalt,
	}

	return r.DB(ctx).Create(&vaultEntity).Error
}
