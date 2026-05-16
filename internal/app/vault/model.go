package vault

import "time"

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

type ArgonConfig struct {
	ID          int64
	Iteration   uint32
	Memory      uint32
	Parallelism uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
