package response

type RegisterSessionResponse struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type LoginResponse struct {
	User        UserResponse        `json:"user"`
	Vault       VaultResponse       `json:"vault"`
	ArgonConfig ArgonConfigResponse `json:"argonConfig"`
	Token       TokenResponse       `json:"token"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type VerifyEmailResponse struct {
	Session string `json:"session"`
}
