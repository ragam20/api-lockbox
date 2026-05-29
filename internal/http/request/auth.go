package request

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Verifier string `json:"verifier" binding:"required"`
}

type SendVerifyLinkRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RequestOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6,numeric"`
}

type RegisterSessionRequest struct {
	Session string `json:"session" binding:"required"`
}

type RegisterRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Name             string `json:"name" binding:"required"`
	AuthVerifier     string `json:"authVerifier" binding:"required"`
	AuthVerifierSalt string `json:"authVerifierSalt" binding:"required"`
	MasterKeySalt    string `json:"masterKeySalt" binding:"required"`
	VaultKeySalt     string `json:"vaultKeySalt" binding:"required"`
}
