package user

import "errors"

var (
	ErrUserNotFound = errors.New("USER_NOT_FOUND")

	ErrFieldName = errors.New("USER_ERROR_FIELD_NAME")
)
