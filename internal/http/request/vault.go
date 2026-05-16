package request

type VaultRequest struct {
	VaultKeySalt string `json:"vaultKeySalt" binding:"required"`
}

type ArgonConfigRequest struct {
	Memory      int32 `json:"memory" binding:"required"`
	Iterations  int32 `json:"iterations" binding:"required"`
	Parallelism int32 `json:"parallelism" binding:"required"`
}
