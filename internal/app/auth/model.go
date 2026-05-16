package auth

import "time"

type User struct {
	ID               int64
	Email            string
	Name             string
	ImgUrl           *string
	MasterKeySalt    string
	AuthVerifier     string
	AuthVerifierSalt string
	EmailVerified    bool
	UpdatedAt        time.Time
	CreatedAt        time.Time
}

type ArgonConfig struct {
	ID          int64
	UserID      int64
	Memory      uint32
	Iterations  uint32
	Parallelism uint32
}

type Vault struct {
	ID           int64
	UserID       int64
	Timeout      int32
	TimeoutType  VaultTimeoutType
	VaultKeySalt string
}

type VaultTimeoutType string

const (
	VaultTimeoutTypeLock   VaultTimeoutType = "lock"
	VaultTimeoutTypeLogout VaultTimeoutType = "logout"
)

type SendVerificationLinkInput struct {
	Email string
}

type LoginInput struct {
	UserID    int64
	Email     string
	Verifier  string
	UserAgent string
	IPAddress string
	Device    string
}

type RegisterInput struct {
	Session          string
	Email            string
	Name             string
	AuthVerifier     string
	AuthVerifierSalt string
	MasterKeySalt    string
	Vault            VaultInput
}

type VaultInput struct {
	VaultKeySalt string
}

type UserSession struct {
	ID               int64
	Device           string
	RefreshToken     string
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	LastUsedAt       time.Time
	ExpiresAt        time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
