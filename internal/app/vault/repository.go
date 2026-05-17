package vault

import "context"

type ArgonConfigRepository interface {
	GetConfigByUserID(ctx context.Context, userID int64) (ArgonConfig, error)
}

type VaultRepository interface {
	GetVaultByUserID(ctx context.Context, userID int64) (Vault, error)
}

type VaultItemRepository interface {
	GetVaultItem(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error)
	CreateVaultItem(ctx context.Context, item VaultItem) error
	GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error)
	UpdateVaultItem(ctx context.Context, item VaultItem) error
	ArchiveVaultItem(ctx context.Context, item VaultItem) error
	UnarchiveVaultItem(ctx context.Context, item VaultItem) error
	DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error
}
