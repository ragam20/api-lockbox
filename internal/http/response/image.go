package response

type UploadAvatarPermissionResponse struct {
	UploadURL string `json:"uploadURL"`
	Signature string `json:"signature"`
	APIKey    string `json:"apiKey"`
	PublicID  string `json:"publicId"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
}
