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
