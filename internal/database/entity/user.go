package entity

import "time"

type UserEntity struct {
	ID int64 `gorm:"primaryKey"`

	Name          string `gorm:"size:255;not null"`
	Email         string `gorm:"size:255;not null;uniqueIndex"`
	EmailVerified bool
	ImgURL        *string

	MasterKeySalt string `gorm:"type:text;not null"`

	AuthVerifier     string `gorm:"type:text;not null"`
	AuthVerifierSalt string `gorm:"type:text;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserEntity) TableName() string {
	return "users"
}
