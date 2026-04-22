package entity

import "time"

type VaultItemEntity struct {
	ID      int64       `gorm:"primaryKey"`
	VaultID int64       `gorm:"not null;index"`
	Vault   VaultEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:VaultID;references:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (VaultItemEntity) TableName() string {
	return "vault_items"
}
