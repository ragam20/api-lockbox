package domain

import (
	"strings"
	"time"
)

const (
	defaultArgonMemory      = 64 * 1024
	defaultArgonIterations  = 3
	defaultArgonParallelism = 3
	defaultVaultTimeout     = 10
	defaultVaultTimeoutType = VaultTimeoutTypeLock
)

type VaultTimeoutType string

const (
	VaultTimeoutTypeLock   VaultTimeoutType = "lock"
	VaultTimeoutTypeLogout VaultTimeoutType = "logout"
)

type Vault struct {
	ID           int64
	UserID       int64
	Timeout      int32
	TimeoutType  VaultTimeoutType
	VaultKeySalt string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type VaultItemType string

const (
	VaultItemTypePassword VaultItemType = "password"
	VaultItemTypeNote     VaultItemType = "note"
	VaultItemTypeAPIKey   VaultItemType = "apiKey"
	VaultItemTypeAll      VaultItemType = "all"
)

type VaultItem struct {
	ID              int64
	VaultID         int64
	Title           string
	ItemType        VaultItemType
	IvCiphertext    string
	Ciphertext      string
	IvProtectionKey string
	ProtectionKey   string
	FavoriteAt      *time.Time
	ArchivedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ArgonConfig struct {
	ID          int64
	Iterations  uint32
	Memory      uint32
	Parallelism uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewDefaultVault(id int64, vaultKeySalt string) Vault {
	return Vault{
		ID:           id,
		Timeout:      defaultVaultTimeout,
		TimeoutType:  defaultVaultTimeoutType,
		VaultKeySalt: vaultKeySalt,
	}
}

func NewDefaultArgonConfig(id int64) ArgonConfig {
	return ArgonConfig{
		ID:          id,
		Memory:      defaultArgonMemory,
		Iterations:  defaultArgonIterations,
		Parallelism: defaultArgonParallelism,
	}
}

type ItemProtectionKey struct {
	ID              int64
	IvProtectionKey string
	ProtectionKey   string
}

func (t VaultItemType) IsValid() bool {
	switch t {
	case VaultItemTypePassword, VaultItemTypeNote, VaultItemTypeAPIKey:
		return true
	}
	return false
}

type GetVaultInput struct {
	Query     VaultItemQuery
	VaultType string
	UserID    int64
	Archived  bool
}

type GetVaultItemByIDInput struct {
	UserID   int64
	ItemID   int64
	Archived bool
}

type GetVaultItemsResult struct {
	Items      []VaultItem
	TotalItems int64
}

type VaultItemQuery struct {
	Page   int
	Limit  int
	SortBy string
	Search string
}

func (f *VaultItemQuery) ParseOrder() (string, string) {
	allowedFields := map[string]string{"title": "title", "created_at": "created_at"}
	allowedOrders := map[string]bool{"desc": true, "asc": true}
	s := strings.Split(strings.TrimSpace(f.SortBy), ",")
	if len(s) != 2 {
		return "created_at", "desc"
	}
	field := strings.TrimSpace(s[0])
	order := strings.TrimSpace(s[1])
	column, ok := allowedFields[field]
	if !ok {
		column = "created_at"
	}
	if !allowedOrders[order] {
		order = "desc"
	}
	return column, order
}
