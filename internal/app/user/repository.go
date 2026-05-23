package user

import "context"

type UserRepository interface {
	GetUser(ctx context.Context, userID int64) (User, error)
}
