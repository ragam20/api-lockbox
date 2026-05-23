package user

import (
	"context"
	"errors"

	"lockbox/internal/http/response"
)

type UserService interface {
	GetUser(ctx context.Context, userID int64) (User, error)
}

type userService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) UserService {
	return &userService{}
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
