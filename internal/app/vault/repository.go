package vault

import "context"

type ArgonConfigRepository interface {
	GetConfigByUserID(ctx context.Context, userID int64) (ArgonConfig, error)
}

type VaultRepository interface {
	GetVaultByUserID(ctx context.Context, userID int64) (Vault, error)
}
