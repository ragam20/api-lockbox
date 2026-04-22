package entity

import "time"

type SessionEntity struct {
	ID        int64      `gorm:"primaryKey"`
	UserID    int64      `gorm:"not null;index"`
	User      UserEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	ExpiredAt time.Time  `gorm:"not null;index"`
}

func (SessionEntity) TableName() string {
	return "sessions"
}
