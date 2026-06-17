package response

import "time"

type UserResponse struct {
	ID               int64     `json:"id,string"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	ImgURL           *string   `json:"imgURL"`
	AuthVerifierSalt string    `json:"authVerifierSalt"`
	MasterKeySalt    string    `json:"masterKeySalt"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
