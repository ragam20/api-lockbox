package entity

import (
	"time"

	"lockbox/internal/app/vault"
)

type ArgonConfigEntity struct {
	ID     int64 `gorm:"primaryKey"`
	UserID int64 `gorm:"not null;uniqueIndex"`

	Time        uint32 `gorm:"not null;default:3;check:time >= 1 AND time <= 8"`
	Memory      uint32 `gorm:"not null;default:65536;check:memory >= 65536 AND memory <= 131072"`
	Parallelism uint32 `gorm:"not null;default:2;check:parallelism >= 1 AND parallelism <= 7"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User UserEntity `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;references:ID"`
}

func (ArgonConfigEntity) TableName() string {
	return "argon_config"
}

func (e ArgonConfigEntity) ToDomain() vault.ArgonConfig {
	return vault.ArgonConfig{
		ID:          e.ID,
		Iterations:  e.Time,
		Parallelism: e.Parallelism,
		Memory:      e.Memory,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
