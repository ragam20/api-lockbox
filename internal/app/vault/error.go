package vault

import (
	"errors"

	vaultdomain "lockbox/internal/app/vault/domain"
)

type Vault = vaultdomain.Vault
type VaultItem = vaultdomain.VaultItem
type VaultItemType = vaultdomain.VaultItemType
type VaultTimeoutType = vaultdomain.VaultTimeoutType
type ArgonConfig = vaultdomain.ArgonConfig
type ItemProtectionKey = vaultdomain.ItemProtectionKey
type VaultRepository = vaultdomain.VaultRepository
type GetVaultInput = vaultdomain.GetVaultInput
type GetVaultItemByIDInput = vaultdomain.GetVaultItemByIDInput
type GetVaultItemsResult = vaultdomain.GetVaultItemsResult
type VaultItemQuery = vaultdomain.VaultItemQuery

type UpdateArgonConfigInput struct {
	UserID         int64
	AuthVerifier   string
	ArgonConfig    vaultdomain.ArgonConfig
	ProtectionKeys []vaultdomain.ItemProtectionKey
}

var (
	ErrVaultType     = errors.New("VAULT_INVALID_TYPE")
	ErrNotFound      = errors.New("VAULT_NOT_FOUND")
	ErrRequiredID    = errors.New("VAULT_ID_REQUIRED")
	ErrInvalidUpdate = errors.New("INVALID_UPDATE")

	ErrFieldTitle = errors.New("VAULT_FIELD_TITLE_REQUIRED")

	ErrInvalidInput = errors.New("VAULT_INVALID_VALIDATION")
)

func NewDefaultVault(id int64, vaultKeySalt string) Vault {
	return vaultdomain.NewDefaultVault(id, vaultKeySalt)
}

func NewDefaultArgonConfig(id int64) ArgonConfig {
	return vaultdomain.NewDefaultArgonConfig(id)
}
