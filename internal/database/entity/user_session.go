package entity

import "time"

type UserSessionEntity struct {
	ID     int64 `gorm:"primaryKey"`
	UserID int64 `gorm:"not null;index"`

	Device           string `gorm:"type:varchar(255)"`
	RefreshTokenHash string `gorm:"not null;uniqueIndex"`
	UserAgent        string
	IPAddress        string

	LastUsedAt *time.Time
	ExpiresAt  time.Time `gorm:"not null;index"`

	User UserEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserSessionEntity) TableName() string {
	return "user_sessions"
}
