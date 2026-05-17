package vault

import (
	"strings"
	"time"
)

type VaultTimeoutType string

const (
	VaultTimeoutTypeLock   VaultTimeoutType = "lock"
	VaultTimeoutTypeLogout VaultTimeoutType = "logout"
)

type Vault struct {
	ID           int64
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
	ArchivedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ArgonConfig struct {
	ID          int64
	Iteration   uint32
	Memory      uint32
	Parallelism uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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
}

func (f *VaultItemQuery) ParseOrder() (string, string) {
	allowedFields := map[string]string{
		"title":      "title",
		"created_at": "created_at",
	}

	allowedOrders := map[string]bool{
		"desc": true,
		"asc":  true,
	}

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

func CheckVaultType(vaultType string) bool {
	return vaultType == string(VaultItemTypeNote) || vaultType == string(VaultItemTypePassword)
}
