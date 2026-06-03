package user

import "context"

type UserRepository interface {
	UpdateProfile(ctx context.Context, user User) error
	GetUser(ctx context.Context, userID int64) (User, error)
}
