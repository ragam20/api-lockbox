package request

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

type SaveAvatarRequest struct {
	PublicImgID string `json:"publicImgId"`
	ImgURL      string `json:"imgURL"`
}
