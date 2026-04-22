package entity

import "time"

type VaultEntity struct {
	ID     int64      `gorm:"primaryKey"`
	UserID int64      `gorm:"not null;index"`
	User   UserEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (VaultEntity) TableName() string {
	return "vaults"
}
