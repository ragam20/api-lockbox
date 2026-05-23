package auth

import "errors"

var (
	ErrInvalidCredential  = errors.New("AUTH_INVALID_CREDENTIAL")
	ErrEmailAlreadyExists = errors.New("AUTH_EMAIL_ALREADY_EXISTS")
	ErrUserNotFound       = errors.New("AUTH_USER_NOT_FOUND")
	ErrFailedRegister     = errors.New("AUTH_FAILED_REGISTER")
	ErrInvalidVerify      = errors.New("AUTH_INVALID_VERIFICATION")
	ErrUnauthorized       = errors.New("AUTH_UNAUTHORIZED")
	ErrTokenExpired       = errors.New("AUTH_TOKEN_EXPIRED")
	ErrInvalidToken       = errors.New("AUTH_INVALID_TOKEN")
	ErrSessionRequired    = errors.New("AUTH_SESSION_REQUIRED")
	ErrSessionExpired     = errors.New("AUTH_SSESSION_EXPIRED")

	ErrFieldEmailRequired    = errors.New("AUTH_FIELD_EMAIL_REQUIRED")
	ErrFieldPasswordRequired = errors.New("AUTH_FIELD_PASSWORD_REQUIRED")
	ErrFieldNameRequired     = errors.New("AUTH_FIELD_NAME_REQUIRED")
	ErrValidation            = errors.New("AUTH_INVALID_VALIDATION")
)
