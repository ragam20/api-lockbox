package repository

import (
	"context"

	"lockbox/internal/app/vault"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type vaultRepository struct {
	BaseRepository
}

func NewVaultRepository(db *gorm.DB) vault.VaultRepository {
	return &vaultRepository{BaseRepository: BaseRepository{db}}
}

func (r *vaultRepository) GetVaultByUserID(ctx context.Context, userID int64) (vault.Vault, error) {
	var e entity.VaultEntity
	if err := r.DB(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		return vault.Vault{}, err
	}

	return e.ToDomain(), nil
}
