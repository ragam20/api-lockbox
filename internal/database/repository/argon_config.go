package repository

import (
	"context"

	"lockbox/internal/app/vault"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type argonConfigRepository struct {
	BaseRepository
}

func NewArgonConfigRepository(db *gorm.DB) vault.ArgonConfigRepository {
	return &argonConfigRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *argonConfigRepository) GetConfigByUserID(ctx context.Context, userID int64) (vault.ArgonConfig, error) {
	var e entity.ArgonConfigEntity
	if err := r.DB(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		return vault.ArgonConfig{}, err
	}

	return e.ToDomain(), nil
}
