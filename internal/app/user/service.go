package user

import (
	"context"
	"errors"
	"strings"

	"lockbox/internal/http/response"
)

type UserService interface {
	UpdateProfile(ctx context.Context, input UpdateUserInput) error
	GetUser(ctx context.Context, userID int64) (User, error)
}

type userService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) UserService {
	return &userService{userRepo: userRepo}
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
