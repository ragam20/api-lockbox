// Package storage provides an interface for saving and loading data.
package storage

import "context"

const (
	UserAvatarFolder = "lockbox/users/%d/avatar"
)

type Storage interface {
	RemoveAvatar(ctx context.Context, publicImgID string) error
	Sign(params map[string][]string) (string, error)
	UploadAvatarRequest() StorageResult
}

type StorageResult struct {
	UploadURL string
	Timestamp int64
	CloudName string
	APIKey    string
	APISecret string
	Type      string
}
