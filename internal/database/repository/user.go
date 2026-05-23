package repository

import (
	"context"
	"errors"

	"lockbox/internal/app/user"
	"lockbox/internal/database/entity"

	"gorm.io/gorm"
)

type userRepository struct {
	BaseRepository
}

func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &userRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *userRepository) GetUser(ctx context.Context, userID int64) (user.User, error) {
	var e entity.UserEntity
	if err := r.DB(ctx).Where("id = ?", userID).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user.User{}, user.ErrUserNotFound
		}
		return user.User{}, err
	}

	return user.User{
		ID:               e.ID,
		Name:             e.Name,
		ImgURL:           e.ImgURL,
		Email:            e.Email,
		EmailVerified:    e.EmailVerified,
		AuthVerifierSalt: e.AuthVerifierSalt,
		MasterKeySalt:    e.MasterKeySalt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}, nil
}
