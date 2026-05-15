package request

type SendVerifyLinkRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RegisterRequest struct {
	Email            string       `json:"email" binding:"required,email"`
	Name             string       `json:"name" binding:"required"`
	AuthVerifier     string       `json:"authVerifier" binding:"required"`
	AuthVerifierSalt string       `json:"authVerifierSalt" binding:"required"`
	MasterKeySalt    string       `json:"masterKeySalt" binding:"required"`
	Vault            VaultRequest `json:"vault" binding:"required"`
}
