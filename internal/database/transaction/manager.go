package transaction

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type gormTransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &gormTransactionManager{db: db}
}

func (m *gormTransactionManager) WithinTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
}
