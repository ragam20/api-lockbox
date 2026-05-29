package response

type RequestOTPResponse struct {
	ExpiresIn int `json:"expiresIn"`
}

type VerifyOTPResponse struct {
	Email      string        `json:"email"`
	IsVerified bool          `json:"isVerified"`
	Token      TokenResponse `json:"token"`
}

type RegisterSessionResponse struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	ID    int64  `json:"id,string"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type LoginResponse struct{}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type VerifyEmailResponse struct {
	Session string `json:"session"`
}
