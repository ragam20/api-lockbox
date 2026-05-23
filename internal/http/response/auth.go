package response

type RegisterSessionResponse struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	ID    int64  `json:"id,string"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type LoginResponse struct {
	Token TokenResponse `json:"token"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type VerifyEmailResponse struct {
	Session string `json:"session"`
}
