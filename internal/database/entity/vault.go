package entity

import (
	"time"

	vaultdomain "lockbox/internal/app/vault/domain"
)

type VaultTimeoutType string

const (
	VaultTimeoutTypeLock   VaultTimeoutType = "lock"
	VaultTimeoutTypeLogout VaultTimeoutType = "logout"
)

func (t VaultTimeoutType) IsValid() bool {
	switch t {
	case VaultTimeoutTypeLock, VaultTimeoutTypeLogout:
		return true
	}
	return false
}

type VaultEntity struct {
	ID     int64 `gorm:"primaryKey"`
	UserID int64 `gorm:"not null;index"`

	Timeout     int32            `gorm:"default:10"`
	TimeoutType VaultTimeoutType `gorm:"type:varchar(20);check:timeout_type IN ('lock', 'logout')"`

	VaultKeySalt string `gorm:"type:text;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User UserEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
}

func (VaultEntity) TableName() string {
	return "vaults"
}

func (e VaultEntity) ToDomain() vaultdomain.Vault {
	return vaultdomain.Vault{
		ID:           e.ID,
		Timeout:      e.Timeout,
		TimeoutType:  vaultdomain.VaultTimeoutType(e.TimeoutType),
		VaultKeySalt: e.VaultKeySalt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}
