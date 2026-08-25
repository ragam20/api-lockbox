package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	appvault "lockbox/internal/app/vault"
	vaultdomain "lockbox/internal/app/vault/domain"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type vaultRepository struct {
	BaseRepository
}

func NewVaultRepository(db *gorm.DB) vaultdomain.VaultRepository {
	return &vaultRepository{BaseRepository: BaseRepository{db}}
}

func (r *vaultRepository) WithTx(tx *gorm.DB) vaultdomain.VaultRepository {
	return &vaultRepository{BaseRepository: BaseRepository{tx}}
}

func (r *vaultRepository) UpdateArgonConfig(ctx context.Context, argonConfig vaultdomain.ArgonConfig) error {
	result := r.DB(ctx).
		Model(&entity.ArgonConfigEntity{}).
		Where("id = ?", argonConfig.ID).
		Updates(map[string]any{
			"memory":      argonConfig.Memory,
			"parallelism": argonConfig.Parallelism,
			"time":        argonConfig.Iterations,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return appvault.ErrNotFound
	}

	return nil
}

func (r *vaultRepository) UpdateProtectionKeys(ctx context.Context, keys []vaultdomain.ItemProtectionKey) error {
	if len(keys) == 0 {
		return nil
	}

	const batchSize = 1000

	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(keys); i += batchSize {
			end := min(i+batchSize, len(keys))

			if err := updateProtectionKeysBatch(tx, keys[i:end]); err != nil {
				return err
			}
		}

		return nil
	})
}

func updateProtectionKeysBatch(
	tx *gorm.DB,
	keys []vaultdomain.ItemProtectionKey,
) error {
	var (
		ids              = make([]int64, 0, len(keys))
		protectionArgs   = make([]any, 0, len(keys)*2)
		ivArgs           = make([]any, 0, len(keys)*2)
		protectionCase   strings.Builder
		ivProtectionCase strings.Builder
	)

	protectionCase.WriteString("CASE id ")
	ivProtectionCase.WriteString("CASE id ")

	for _, key := range keys {
		ids = append(ids, key.ID)

		protectionCase.WriteString("WHEN ? THEN ? ")
		protectionArgs = append(
			protectionArgs,
			key.ID,
			key.ProtectionKey,
		)

		ivProtectionCase.WriteString("WHEN ? THEN ? ")
		ivArgs = append(
			ivArgs,
			key.ID,
			key.IvProtectionKey,
		)
	}

	protectionCase.WriteString("END")
	ivProtectionCase.WriteString("END")

	args := make([]any, 0)

	args = append(args, protectionArgs...)
	args = append(args, ivArgs...)

	query := fmt.Sprintf(`
		UPDATE vault_items
		SET
			protection_key = %s,
			iv_protection_key = %s
		WHERE id IN ?
	`, protectionCase.String(), ivProtectionCase.String())

	args = append(args, ids)

	return tx.Exec(query, args...).Error
}

func (r *vaultRepository) GetProtectionKeys(ctx context.Context, userID int64) ([]vaultdomain.ItemProtectionKey, error) {
	var e []entity.VaultItemEntity
	err := r.DB(ctx).
		Model(&entity.VaultItemEntity{}).
		Joins("JOIN vaults ON vaults.id = vault_items.vault_id").
		Where("vaults.user_id = ?", userID).
		Where("vault_items.deleted_at IS NULL").
		Find(&e).
		Error
	if err != nil {
		return []vaultdomain.ItemProtectionKey{}, nil
	}

	result := make([]vaultdomain.ItemProtectionKey, len(e))

	for i, v := range e {
		result[i] = vaultdomain.ItemProtectionKey{
			ID:              v.ID,
			ProtectionKey:   v.ProtectionKey,
			IvProtectionKey: v.IvProtectionKey,
		}
	}
	return result, nil
}

func (r *vaultRepository) SetFavorite(ctx context.Context, itemID int64) error {
	result := r.DB(ctx).Model(&entity.VaultItemEntity{}).
		Where("id = ?", itemID).
		Update("favorite_at", time.Now())

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appvault.ErrNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return appvault.ErrNotFound
	}

	return nil
}

func (r *vaultRepository) DeleteFavorite(ctx context.Context, itemID int64) error {
	result := r.DB(ctx).Model(&entity.VaultItemEntity{}).
		Where("id = ?", itemID).
		Update("favorite_at", nil)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appvault.ErrNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return appvault.ErrNotFound
	}

	return nil
}

func (r *vaultRepository) GetVault(ctx context.Context, userID int64) (vaultdomain.Vault, error) {
	var e entity.VaultEntity
	if err := r.DB(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return vaultdomain.Vault{}, appvault.ErrNotFound
		}
		return vaultdomain.Vault{}, err
	}

	return vaultdomain.Vault{
		ID:           e.ID,
		Timeout:      e.Timeout,
		TimeoutType:  vaultdomain.VaultTimeoutType(e.TimeoutType),
		VaultKeySalt: e.VaultKeySalt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}, nil
}

func (r *vaultRepository) GetArgonConfig(ctx context.Context, userID int64) (vaultdomain.ArgonConfig, error) {
	var e entity.ArgonConfigEntity
	if err := r.DB(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return vaultdomain.ArgonConfig{}, appvault.ErrNotFound
		}
		return vaultdomain.ArgonConfig{}, err
	}

	return vaultdomain.ArgonConfig{
		ID:          e.ID,
		Iterations:  e.Time,
		Memory:      e.Memory,
		Parallelism: e.Parallelism,
		UpdatedAt:   e.UpdatedAt,
		CreatedAt:   e.CreatedAt,
	}, nil
}

func (r *vaultRepository) GetVaultByUserID(ctx context.Context, userID int64) (vaultdomain.Vault, error) {
	var e entity.VaultEntity
	if err := r.DB(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		return vaultdomain.Vault{}, err
	}

	return e.ToDomain(), nil
}

func (r *vaultRepository) GetConfigByUserID(ctx context.Context, userID int64) (vaultdomain.ArgonConfig, error) {
	var e entity.ArgonConfigEntity
	if err := r.DB(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		return vaultdomain.ArgonConfig{}, err
	}

	return e.ToDomain(), nil
}

func (r *vaultRepository) CreateArgonConfig(ctx context.Context, argonConfig vaultdomain.ArgonConfig, userID int64) error {
	argonEntity := entity.ArgonConfigEntity{
		ID:          argonConfig.ID,
		UserID:      userID,
		Time:        argonConfig.Iterations,
		Memory:      argonConfig.Memory,
		Parallelism: argonConfig.Parallelism,
	}

	return r.DB(ctx).Create(&argonEntity).Error
}

func (r *vaultRepository) CreateVault(ctx context.Context, vault vaultdomain.Vault, userID int64) error {
	vaultEntity := entity.VaultEntity{
		ID:           vault.ID,
		UserID:       userID,
		Timeout:      vault.Timeout,
		TimeoutType:  entity.VaultTimeoutType(vault.TimeoutType),
		VaultKeySalt: vault.VaultKeySalt,
	}

	return r.DB(ctx).Create(&vaultEntity).Error
}

func (r *vaultRepository) GetVaultItems(ctx context.Context, input vaultdomain.GetVaultInput) (vaultdomain.GetVaultItemsResult, error) {
	return r.listVaultItems(ctx, input, false)
}

func (r *vaultRepository) GetFavorites(ctx context.Context, input vaultdomain.GetVaultInput) (vaultdomain.GetVaultItemsResult, error) {
	return r.listVaultItems(ctx, input, true)
}

func (r *vaultRepository) listVaultItems(ctx context.Context, input vaultdomain.GetVaultInput, favoritesOnly bool) (vaultdomain.GetVaultItemsResult, error) {
	var e []entity.VaultItemEntity

	archivedClause := "vault_items.archived_at IS NULL"
	if input.Archived {
		archivedClause = "vault_items.archived_at IS NOT NULL"
	}

	db := r.DB(ctx).
		Joins("JOIN vaults ON vaults.id = vault_items.vault_id").
		Where("vaults.user_id = ? AND vault_items.deleted_at IS NULL AND "+archivedClause, input.UserID)

	if favoritesOnly {
		db = db.Where("vault_items.favorite_at IS NOT NULL")
	}

	if input.VaultType != "all" {
		db = db.Where("vault_items.item_type = ?", input.VaultType)
	}

	if search := strings.TrimSpace(input.Query.Search); search != "" {
		db = db.Where("vault_items.title ILIKE ?", "%"+search+"%")
	}

	countDB := db.Session(&gorm.Session{})
	var totalItems int64
	if err := countDB.Model(&entity.VaultItemEntity{}).Count(&totalItems).Error; err != nil {
		return vaultdomain.GetVaultItemsResult{}, err
	}

	field, order := input.Query.ParseOrder()
	offset := (input.Query.Page - 1) * input.Query.Limit
	db = db.Offset(offset).
		Limit(input.Query.Limit).
		Order(fmt.Sprintf("%s %s", field, order))

	if err := db.Find(&e).Error; err != nil {
		return vaultdomain.GetVaultItemsResult{}, err
	}

	domain := make([]vaultdomain.VaultItem, len(e))
	for i, item := range e {
		domain[i] = item.ToDomain()
	}

	return vaultdomain.GetVaultItemsResult{
		Items:      domain,
		TotalItems: totalItems,
	}, nil
}

func (r *vaultRepository) CreateVaultItem(ctx context.Context, item vaultdomain.VaultItem) error {
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

func (r *vaultRepository) GetVaultItemByID(ctx context.Context, input vaultdomain.GetVaultItemByIDInput) (vaultdomain.VaultItem, error) {
	var e entity.VaultItemEntity

	archivedClause := "vault_items.archived_at IS NULL"
	if input.Archived {
		archivedClause = "vault_items.archived_at IS NOT NULL"
	}

	if err := r.DB(ctx).
		Joins("JOIN vaults ON vaults.id = vault_items.vault_id").
		Where("vault_items.id = ? AND vaults.user_id = ? AND vault_items.deleted_at IS NULL AND "+archivedClause, input.ItemID, input.UserID).
		First(&e).Error; err != nil {
		return vaultdomain.VaultItem{}, err
	}

	return e.ToDomain(), nil
}

func (r *vaultRepository) ArchiveVaultItem(ctx context.Context, item vaultdomain.VaultItem) error {
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

func (r *vaultRepository) UnarchiveVaultItem(ctx context.Context, item vaultdomain.VaultItem) error {
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

func (r *vaultRepository) UpdateVaultItem(ctx context.Context, item vaultdomain.VaultItem) error {
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

func (r *vaultRepository) DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error {
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
