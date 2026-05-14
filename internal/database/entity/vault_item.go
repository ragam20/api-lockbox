package entity

import "time"

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

	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	Vault VaultEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:VaultID;references:ID"`
}

func (VaultItemEntity) TableName() string {
	return "vault_items"
}
