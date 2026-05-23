package vault

import "errors"

var (
	ErrVaultType  = errors.New("VAULT_INVALID_TYPE")
	ErrNotFound   = errors.New("VAULT_NOT_FOUND")
	ErrRequiredID = errors.New("VAULT_ID_REQUIRED")

	ErrFieldTitle = errors.New("VAULT_FIELD_TITLE_REQUIRED")

	ErrInvalidInput = errors.New("VAULT_INVALID_VALIDATION")
)
