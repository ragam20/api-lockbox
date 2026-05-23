package vault

import "context"

type VaultRepository interface {
	GetArgonConfig(ctx context.Context, userID int64) (ArgonConfig, error)
	GetVaultByUserID(ctx context.Context, userID int64) (Vault, error)
	GetConfigByUserID(ctx context.Context, userID int64) (ArgonConfig, error)
	CreateArgonConfig(ctx context.Context, argonConfig ArgonConfig, userID int64) error
	CreateVault(ctx context.Context, vault Vault, userID int64) error
	GetVaultItem(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error)
	CreateVaultItem(ctx context.Context, item VaultItem) error
	GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error)
	UpdateVaultItem(ctx context.Context, item VaultItem) error
	ArchiveVaultItem(ctx context.Context, item VaultItem) error
	UnarchiveVaultItem(ctx context.Context, item VaultItem) error
	DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error
}
