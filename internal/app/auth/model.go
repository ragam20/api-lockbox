package auth

type User struct {
	ID               int64
	Email            string
	Name             string
	MasterKeySalt    string
	AuthVerifier     string
	AuthVerifierSalt string
	EmailVerified    bool
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
