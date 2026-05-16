package response

import "time"

type RegisterResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type LoginResponse struct {
	UserID int64   `json:"userId"`
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	ImgURL *string `json:"imgURL"`

	MasterKeySalt string `json:"masterKeySalt"`

	Vault       VaultResponse       `json:"vault"`
	ArgonConfig ArgonConfigResponse `json:"argonConfig"`
	Token       TokenResponse       `json:"token"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
