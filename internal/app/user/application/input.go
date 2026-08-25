package application

type UpdateUserInput struct {
	UserID int64
	Name   string
}

type SaveAvatarURLInput struct {
	UserID      int64
	PublicImgID string
	ImgURL      string
}
