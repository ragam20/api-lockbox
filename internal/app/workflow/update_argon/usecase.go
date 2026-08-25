package updateargon

import (
	"context"

	"lockbox/internal/app/vault"
	dbtx "lockbox/internal/database/transaction"
	"lockbox/internal/http/response"

	authdomain "lockbox/internal/app/auth/domain"
	vaultapp "lockbox/internal/app/vault/application"
	vaultdomain "lockbox/internal/app/vault/domain"

	"gorm.io/gorm"
)

type UpdateArgonConfig interface {
	Execute(ctx context.Context, input vaultapp.UpdateArgonConfigInput) error
}

type updateArgonConfig struct {
	authRepo  authdomain.AuthRepository
	vaultRepo vaultdomain.VaultRepository
	txManager dbtx.TransactionManager
}

func NewUpdateArgon(authRepo authdomain.AuthRepository, vaultRepo vaultapp.VaultRepository, txManager dbtx.TransactionManager) UpdateArgonConfig {
	return &updateArgonConfig{
		authRepo:  authRepo,
		vaultRepo: vaultRepo,
		txManager: txManager,
	}
}

func (s *updateArgonConfig) Execute(ctx context.Context, input vaultapp.UpdateArgonConfigInput) error {
	if err := s.txManager.WithinTransaction(ctx, func(tx *gorm.DB) error {
		vaultRepo := s.vaultRepo.WithTx(tx)
		authRepo := s.authRepo.WithTx(tx)

		if err := vaultRepo.UpdateArgonConfig(ctx, input.ArgonConfig); err != nil {
			return err
		}

		if err := authRepo.UpdateAuthVerifier(ctx, input.AuthVerifier, input.UserID); err != nil {
			return err
		}

		if err := vaultRepo.UpdateProtectionKeys(ctx, input.ProtectionKeys); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return response.BadRequestError(vault.ErrInvalidUpdate.Error(), "invalid update argon config. Please try again later!")
	}

	return nil
}
