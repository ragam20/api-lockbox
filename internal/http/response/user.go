package response

import "time"

type UserResponse struct {
	UserID        int64     `json:"userId"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	ImgURL        *string   `json:"imgURL"`
	MasterKeySalt string    `json:"masterKeySalt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
