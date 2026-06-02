package entity

import (
	"time"

	"lockbox/internal/app/vault"
)

type VaultItemType string

const (
	VaultItemTypePassword VaultItemType = "password"
	VaultItemTypeNote     VaultItemType = "note"
)

func (t VaultItemType) IsValid() bool {
	switch t {
	case VaultItemTypeNote, VaultItemTypePassword:
		return true
	}
	return false
}

type VaultItemEntity struct {
	ID      int64 `gorm:"primaryKey"`
	VaultID int64 `gorm:"not null;index"`

	Title    string
	ItemType VaultItemType `gorm:"type:varchar(20);not null;check:item_type IN ('password', 'note')"`

	IvCiphertext string `gorm:"type:text;not null"`
	Ciphertext   string `gorm:"type:text;not null"`

	IvProtectionKey string `gorm:"type:text;not null"`
	ProtectionKey   string `gorm:"type:text;not null"`

	FavoriteAt *time.Time
	ArchivedAt *time.Time

	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	Vault VaultEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:VaultID;references:ID"`
}

func (VaultItemEntity) TableName() string {
	return "vault_items"
}

func (e VaultItemEntity) ToDomain() vault.VaultItem {
	return vault.VaultItem{
		ID:              e.ID,
		VaultID:         e.VaultID,
		Title:           e.Title,
		ItemType:        vault.VaultItemType(e.ItemType),
		IvCiphertext:    e.IvCiphertext,
		Ciphertext:      e.Ciphertext,
		IvProtectionKey: e.IvProtectionKey,
		ProtectionKey:   e.ProtectionKey,
		FavoriteAt:      e.FavoriteAt,
		ArchivedAt:      e.ArchivedAt,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}
