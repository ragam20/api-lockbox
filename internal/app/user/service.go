package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"lockbox/internal/http/response"
	"lockbox/internal/storage"
)

type UserService interface {
	RemoveAvatar(ctx context.Context, userID int64) error
	SaveAvatarURL(ctx context.Context, input SaveAvatarURLInput) error
	UploadAvatarRequest(ctx context.Context, userID int64) (response.UploadAvatarPermissionResponse, error)
	UpdateProfile(ctx context.Context, input UpdateUserInput) error
	GetUser(ctx context.Context, userID int64) (User, error)
}

type userService struct {
	userRepo UserRepository
	storage  storage.Storage
}

func NewUserService(userRepo UserRepository, storage storage.Storage) UserService {
	return &userService{userRepo: userRepo, storage: storage}
}

func (c *userService) RemoveAvatar(ctx context.Context, userID int64) error {
	user, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.BadRequestError(ErrFieldName.Error(), "user not found")
		}
		return response.InternalServerError(err.Error())
	}

	if user.PublicImgID == nil && user.ImgURL == nil {
		return nil
	}

	if err := c.storage.RemoveAvatar(ctx, *user.PublicImgID); err != nil {
		return response.BadRequestError(ErrUserUpdate.Error(), "can't remove avatar")
	}

	user.ImgURL = nil
	user.PublicImgID = nil

	if err := c.userRepo.RemoveAvatar(ctx, user); err != nil {
		return response.InternalServerError(err.Error())
	}

	return nil
}

func (c *userService) SaveAvatarURL(ctx context.Context, input SaveAvatarURLInput) error {
	usr, err := c.GetUser(ctx, input.UserID)
	if err != nil {
		return err
	}

	publicImgID := strings.TrimSpace(input.PublicImgID)
	imgURL := strings.TrimSpace(input.ImgURL)
	usr.PublicImgID = &publicImgID
	usr.ImgURL = &imgURL

	if err := c.userRepo.UpdateAvatar(ctx, usr); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.BadRequestError(ErrUserNotFound.Error(), "user not found")
		}

		return response.InternalServerError(err.Error())
	}

	return nil
}

func (c *userService) UploadAvatarRequest(ctx context.Context, userID int64) (response.UploadAvatarPermissionResponse, error) {
	storageResult := c.storage.UploadAvatarRequest()

	user, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.UploadAvatarPermissionResponse{}, response.BadRequestError(ErrUserNotFound.Error(), "user not found")
		}

		return response.UploadAvatarPermissionResponse{}, response.InternalServerError(err.Error())
	}

	publicID := fmt.Sprintf(storage.UserAvatarFolder, user.ID)

	params := map[string][]string{
		"public_id": {publicID},
		"timestamp": {fmt.Sprintf("%d", storageResult.Timestamp)},
	}

	signature, err := c.storage.Sign(params)
	if err != nil {
		return response.UploadAvatarPermissionResponse{}, response.InternalServerError(err.Error())
	}

	return response.UploadAvatarPermissionResponse{
		PublicID:  publicID,
		Signature: signature,
		APIKey:    storageResult.APIKey,
		UploadURL: storageResult.UploadURL,
		Timestamp: storageResult.Timestamp,
		Type:      storageResult.Type,
	}, nil
}

func (c *userService) UpdateProfile(ctx context.Context, input UpdateUserInput) error {
	if input.Name == "" {
		return response.BadRequestError(ErrFieldName.Error(), "field name is required")
	}

	newName := strings.TrimSpace(input.Name)

	usr, err := c.GetUser(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.BadRequestError(ErrUserNotFound.Error(), "user not found")
		}

		return response.InternalServerError(err.Error())
	}

	usr.Name = newName

	if err := c.userRepo.UpdateProfile(ctx, usr); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return response.BadRequestError(ErrUserNotFound.Error(), "user not found")
		}

		return response.InternalServerError(err.Error())
	}

	return nil
}

func (s *userService) GetUser(ctx context.Context, userID int64) (User, error) {
	usr, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return User{}, response.NotFoundError(err.Error(), "user not found")
		}
		return User{}, response.InternalServerError(err.Error())
	}

	return usr, nil
}
