package response

import "time"

type VaultResponse struct {
	ID           int64     `json:"id,string"`
	Timeout      int32     `json:"timeout"`
	TimeoutType  string    `json:"timeoutType"`
	VaultKeySalt string    `json:"vaultKeySalt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ArgonConfigResponse struct {
	ID          int64     `json:"id,string"`
	Iteration   uint32    `json:"iteration"`
	Memory      uint32    `json:"memory"`
	Parallelism uint32    `json:"parallelism"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type VaultItemResponse struct {
	ID              int64      `json:"id,string"`
	VaultID         int64      `json:"vaultId"`
	Title           string     `json:"title"`
	ItemType        string     `json:"itemType"`
	IvCiphertext    string     `json:"ivCiphertext"`
	Ciphertext      string     `json:"ciphertext"`
	IvProtectionKey string     `json:"ivProtectionKey"`
	ProtectionKey   string     `json:"protectionKey"`
	ArchivedAt      *time.Time `json:"archivedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}
