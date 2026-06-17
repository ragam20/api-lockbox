package user

import "context"

type UserRepository interface {
	RemoveAvatar(ctx context.Context, user User) error
	UpdateProfile(ctx context.Context, user User) error
	UpdateAvatar(ctx context.Context, user User) error
	GetUser(ctx context.Context, userID int64) (User, error)
}
