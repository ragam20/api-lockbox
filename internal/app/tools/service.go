package tools

import (
	"context"

	"lockbox/internal/app/auth"
	"lockbox/internal/app/vault"
	dbtx "lockbox/internal/database/transaction"
	"lockbox/internal/http/response"

	"gorm.io/gorm"
)

type ToolService interface {
	UpdateArgonConfig(ctx context.Context, input vault.UpdateArgonConfigInput) error
}

type toolService struct {
	authRepo  auth.AuthRepository
	vaultRepo vault.VaultRepository
	txManager dbtx.TransactionManager
}

func NewToolService(authRepo auth.AuthRepository,
	vaultRepo vault.VaultRepository,
	txManager dbtx.TransactionManager,
) ToolService {
	return &toolService{
		authRepo:  authRepo,
		vaultRepo: vaultRepo,
		txManager: txManager,
	}
}

func (s *toolService) UpdateArgonConfig(ctx context.Context, input vault.UpdateArgonConfigInput) error {
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
