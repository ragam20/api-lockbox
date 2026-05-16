package entity

import (
	"time"

	"lockbox/internal/app/auth"
)

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

func (e UserEntity) ToDomainUser() auth.User {
	return auth.User{
		ID:               e.ID,
		Name:             e.Name,
		Email:            e.Email,
		EmailVerified:    e.EmailVerified,
		ImgUrl:           e.ImgURL,
		MasterKeySalt:    e.MasterKeySalt,
		AuthVerifier:     e.AuthVerifier,
		AuthVerifierSalt: e.AuthVerifierSalt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}
