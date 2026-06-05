package user

import "errors"

var (
	ErrUserNotFound = errors.New("USER_NOT_FOUND")

	ErrUserUpdate = errors.New("USER_UPDATE_INVALID")

	ErrFieldName = errors.New("USER_ERROR_FIELD_NAME")
)
