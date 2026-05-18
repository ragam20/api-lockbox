package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lockbox/internal/app/vault"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type vaultItemRepository struct {
	BaseRepository
}

func NewVaultItemRepository(db *gorm.DB) vault.VaultItemRepository {
	return &vaultItemRepository{BaseRepository: BaseRepository{db}}
}

func (r *vaultItemRepository) GetVaultItem(ctx context.Context, input vault.GetVaultInput) (vault.GetVaultItemsResult, error) {
	var e []entity.VaultItemEntity

	archivedClause := "vault_items.archived_at IS NULL"
	if input.Archived {
		archivedClause = "vault_items.archived_at IS NOT NULL"
	}

	db := r.DB(ctx).
		Joins("JOIN vaults ON vaults.id = vault_items.vault_id").
		Where("vaults.user_id = ? AND vault_items.item_type = ? AND vault_items.deleted_at IS NULL AND "+archivedClause, input.UserID, input.VaultType)

	if search := strings.TrimSpace(input.Query.Search); search != "" {
		db = db.Where("vault_items.title ILIKE ?", "%"+search+"%")
	}

	countDB := db.Session(&gorm.Session{})
	var totalItems int64
	if err := countDB.Model(&entity.VaultItemEntity{}).Count(&totalItems).Error; err != nil {
		return vault.GetVaultItemsResult{}, err
	}

	field, order := input.Query.ParseOrder()
	offset := (input.Query.Page - 1) * input.Query.Limit
	db = db.Offset(offset).
		Limit(input.Query.Limit).
		Order(fmt.Sprintf("%s %s", field, order))

	if err := db.Find(&e).Error; err != nil {
		return vault.GetVaultItemsResult{}, err
	}

	domain := make([]vault.VaultItem, len(e))
	for i, item := range e {
		domain[i] = item.ToDomain()
	}

	return vault.GetVaultItemsResult{
		Items:      domain,
		TotalItems: totalItems,
	}, nil
}

func (r *vaultItemRepository) CreateVaultItem(ctx context.Context, item vault.VaultItem) error {
	itemEntity := entity.VaultItemEntity{
		ID:              item.ID,
		VaultID:         item.VaultID,
		Title:           item.Title,
		ItemType:        entity.VaultItemType(item.ItemType),
		IvCiphertext:    item.IvCiphertext,
		Ciphertext:      item.Ciphertext,
		IvProtectionKey: item.IvProtectionKey,
		ProtectionKey:   item.ProtectionKey,
	}

	return r.DB(ctx).Create(&itemEntity).Error
}

func (r *vaultItemRepository) GetVaultItemByID(ctx context.Context, input vault.GetVaultItemByIDInput) (vault.VaultItem, error) {
	var e entity.VaultItemEntity

	archivedClause := "vault_items.archived_at IS NULL"
	if input.Archived {
		archivedClause = "vault_items.archived_at IS NOT NULL"
	}

	if err := r.DB(ctx).
		Joins("JOIN vaults ON vaults.id = vault_items.vault_id").
		Where("vault_items.id = ? AND vaults.user_id = ? AND vault_items.deleted_at IS NULL AND "+archivedClause, input.ItemID, input.UserID).
		First(&e).Error; err != nil {
		return vault.VaultItem{}, err
	}

	return e.ToDomain(), nil
}

func (r *vaultItemRepository) ArchiveVaultItem(ctx context.Context, item vault.VaultItem) error {
	result := r.DB(ctx).
		Model(&entity.VaultItemEntity{}).
		Where("id = ? AND vault_id = ? AND deleted_at IS NULL AND archived_at IS NULL", item.ID, item.VaultID).
		Update("archived_at", item.ArchivedAt)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *vaultItemRepository) UnarchiveVaultItem(ctx context.Context, item vault.VaultItem) error {
	result := r.DB(ctx).
		Model(&entity.VaultItemEntity{}).
		Where("id = ? AND vault_id = ? AND deleted_at IS NULL AND archived_at IS NOT NULL", item.ID, item.VaultID).
		Update("archived_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *vaultItemRepository) UpdateVaultItem(ctx context.Context, item vault.VaultItem) error {
	result := r.DB(ctx).
		Model(&entity.VaultItemEntity{}).
		Where("id = ? AND vault_id = ? AND deleted_at IS NULL", item.ID, item.VaultID).
		Updates(map[string]any{
			"title":             item.Title,
			"item_type":         entity.VaultItemType(item.ItemType),
			"iv_ciphertext":     item.IvCiphertext,
			"ciphertext":        item.Ciphertext,
			"iv_protection_key": item.IvProtectionKey,
			"protection_key":    item.ProtectionKey,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *vaultItemRepository) DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error {
	now := time.Now().UTC()
	result := r.DB(ctx).
		Model(&entity.VaultItemEntity{}).
		Where(`
			id = ?
			AND deleted_at IS NULL
			AND vault_id IN (
				SELECT id FROM vaults WHERE user_id = ?
			)
		`, itemID, userID).
		Update("deleted_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
