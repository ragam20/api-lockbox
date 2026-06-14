package request

type VaultRequest struct {
	VaultKeySalt string `json:"vaultKeySalt" binding:"required"`
}

type ArgonConfigRequest struct {
	ID          int64 `json:"id,string" binding:"required"`
	Memory      int32 `json:"memory" binding:"required"`
	Iterations  int32 `json:"iterations" binding:"required"`
	Parallelism int32 `json:"parallelism" binding:"required"`
}

type VaultItemRequest struct {
	Title           string `json:"title" binding:"required"`
	ItemType        string `json:"itemType" binding:"required"`
	IvCiphertext    string `json:"ivCiphertext" binding:"required"`
	Ciphertext      string `json:"ciphertext" binding:"required"`
	IvProtectionKey string `json:"ivProtectionKey" binding:"required"`
	ProtectionKey   string `json:"protectionKey" binding:"required"`
}

type UpdateProtectionKeyRequest struct {
	ItemID          int64  `json:"itemId,string" binding:"required"`
	IvProtectionKey string `json:"ivProtectionKey" binding:"required"`
	ProtectionKey   string `json:"protectionKey" binding:"required"`
}

type UpdateArgonConfigRequest struct {
	AuthVerifier   string                       `json:"authVerifier" binding:"required"`
	ArgonConfig    ArgonConfigRequest           `json:"argonConfig" binding:"required"`
	ProtectionKeys []UpdateProtectionKeyRequest `json:"protectionKeys" binding:"required"`
}
