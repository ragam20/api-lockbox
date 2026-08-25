package updateargon

import (
	"context"

	"lockbox/internal/app/auth"
	"lockbox/internal/app/vault"
	dbtx "lockbox/internal/database/transaction"
	"lockbox/internal/http/response"

	"gorm.io/gorm"
)

type UpdateArgonConfig interface {
	Execute(ctx context.Context, input vault.UpdateArgonConfigInput) error
}

type updateArgonConfig struct {
	authRepo  auth.AuthRepository
	vaultRepo vault.VaultRepository
	txManager dbtx.TransactionManager
}

func NewUpdateArgon(authRepo auth.AuthRepository, vaultRepo vault.VaultRepository, txManager dbtx.TransactionManager) UpdateArgonConfig {
	return &updateArgonConfig{
		authRepo:  authRepo,
		vaultRepo: vaultRepo,
		txManager: txManager,
	}
}

func (s *updateArgonConfig) Execute(ctx context.Context, input vault.UpdateArgonConfigInput) error {
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
