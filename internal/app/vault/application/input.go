package application

import (
	vaultdomain "lockbox/internal/app/vault/domain"
)

type (
	Vault                 = vaultdomain.Vault
	VaultItem             = vaultdomain.VaultItem
	VaultItemType         = vaultdomain.VaultItemType
	VaultTimeoutType      = vaultdomain.VaultTimeoutType
	ArgonConfig           = vaultdomain.ArgonConfig
	ItemProtectionKey     = vaultdomain.ItemProtectionKey
	GetVaultInput         = vaultdomain.GetVaultInput
	GetVaultItemIDInput   = vaultdomain.GetVaultItemByIDInput
	GetVaultItemsResult   = vaultdomain.GetVaultItemsResult
	GetVaultItemByIDInput = vaultdomain.GetVaultItemByIDInput
	VaultItemQuery        = vaultdomain.VaultItemQuery
	VaultRepository       = vaultdomain.VaultRepository
)

const (
	VaultTimeoutTypeLock   = vaultdomain.VaultTimeoutTypeLock
	VaultTimeoutTypeLogout = vaultdomain.VaultTimeoutTypeLogout
	VaultItemTypePassword  = vaultdomain.VaultItemTypePassword
	VaultItemTypeNote      = vaultdomain.VaultItemTypeNote
	VaultItemTypeAPIKey    = vaultdomain.VaultItemTypeAPIKey
	VaultItemTypeAll       = vaultdomain.VaultItemTypeAll
)

type CreateVaultItemInput struct {
	UserID          int64
	Title           string
	ItemType        VaultItemType
	IvCiphertext    string
	Ciphertext      string
	IvProtectionKey string
	ProtectionKey   string
}

type UpdateVaultItemInput struct {
	UserID          int64
	ItemID          int64
	Title           string
	ItemType        VaultItemType
	IvCiphertext    string
	Ciphertext      string
	IvProtectionKey string
	ProtectionKey   string
}

type UpdateArgonConfigInput struct {
	UserID         int64
	AuthVerifier   string
	ArgonConfig    ArgonConfig
	ProtectionKeys []vaultdomain.ItemProtectionKey
}

func CheckVaultType(vaultType string) bool {
	return vaultType == string(VaultItemTypeNote) || vaultType == string(VaultItemTypePassword) || vaultType == string(VaultItemTypeAPIKey) || vaultType == string(VaultItemTypeAll)
}
