package entity

import "time"

type UserEntity struct {
	ID       int64  `gorm:"primaryKey"`
	Name     string `gorm:"size:255;not null"`
	Email    string `gorm:"size:255;not null;uniqueIndex"`
	ImageURL string `gorm:"column:image_url;type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserEntity) TableName() string {
	return "users"
}
