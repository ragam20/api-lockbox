package cloudinary

import (
	"context"
	"fmt"
	"time"

	"lockbox/internal/config"
	"lockbox/internal/logger"
	"lockbox/internal/storage"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const (
	UploadURL = "https://api.cloudinary.com/v1_1/"
)

type CloudinaryService struct {
	cfg    *config.CloudinaryConfig
	cld    *cloudinary.Cloudinary
	logger logger.Logger
}

func NewCloudinaryService(cfg *config.CloudinaryConfig, logger logger.Logger) storage.Storage {
	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		logger.Error("Cloudinary service error", err, nil)
		return nil
	}

	return &CloudinaryService{cfg: cfg, cld: cld, logger: logger}
}

func (c *CloudinaryService) RemoveAvatar(ctx context.Context, publicImgID string) error {
	_, err := c.cld.Upload.Destroy(
		ctx,
		uploader.DestroyParams{
			PublicID: publicImgID,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *CloudinaryService) UploadAvatarRequest() storage.StorageResult {
	url := fmt.Sprintf("%s%s/image/upload", UploadURL, c.cfg.CloudName)
	return storage.StorageResult{
		UploadURL: url,
		Timestamp: time.Now().Unix(),
		CloudName: c.cfg.CloudName,
		APIKey:    c.cfg.APIKey,
		APISecret: c.cfg.APISecret,
		Type:      "image",
	}
}

func (c *CloudinaryService) Sign(params map[string][]string) (string, error) {
	signature, err := api.SignParameters(params, c.cfg.APISecret)
	if err != nil {
		return "", err
	}

	return signature, nil
}
