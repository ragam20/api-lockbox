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

func (r *userRepository) RemoveAvatar(ctx context.Context, usr user.User) error {
	result := r.DB(ctx).Model(&entity.UserEntity{}).Where("id = ?", usr.ID).
		Updates(map[string]any{
			"img_url":       usr.ImgURL,
			"public_img_id": usr.PublicImgID,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, usr user.User) error {
	result := r.DB(ctx).Model(&entity.UserEntity{}).Where("id = ?", usr.ID).Updates(map[string]any{
		"name": usr.Name,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateAvatar(ctx context.Context, usr user.User) error {
	result := r.DB(ctx).Model(&entity.UserEntity{}).Where("id = ?", usr.ID).Updates(map[string]any{
		"public_img_id": usr.PublicImgID,
		"img_url":       usr.ImgURL,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
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
		PublicImgID:      e.PublicImgID,
		ImgURL:           e.ImgURL,
		Email:            e.Email,
		EmailVerified:    e.EmailVerified,
		AuthVerifierSalt: e.AuthVerifierSalt,
		MasterKeySalt:    e.MasterKeySalt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}, nil
}
