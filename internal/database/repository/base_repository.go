package repository

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) BaseRepository {
	return BaseRepository{db: db}
}

func (r BaseRepository) DB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r BaseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
