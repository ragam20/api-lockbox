package vault

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"lockbox/internal/http/response"
	"lockbox/internal/idgen"

	"gorm.io/gorm"
)

type (
	VaultService interface {
		GetVaultItem(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error)
		GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error)
		CreateVaultItem(ctx context.Context, input CreateVaultItemInput) (VaultItem, error)
		UpdateVaultItem(ctx context.Context, input UpdateVaultItemInput) (VaultItem, error)
		ArchiveVaultItem(ctx context.Context, userID int64, itemID int64) (VaultItem, error)
		UnarchiveVaultItem(ctx context.Context, userID int64, itemID int64) (VaultItem, error)
		DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error
	}

	vaultService struct {
		vaultRepo     VaultRepository
		vaultItemRepo VaultItemRepository
		idGen         idgen.IDGenerator
	}
)

func NewVaultService(vaultRepo VaultRepository, vaultItemRepo VaultItemRepository, idGen idgen.IDGenerator) VaultService {
	return &vaultService{
		vaultRepo:     vaultRepo,
		vaultItemRepo: vaultItemRepo,
		idGen:         idGen,
	}
}

func (s *vaultService) GetVaultItem(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error) {
	if input.UserID == 0 {
		return GetVaultItemsResult{}, response.Unauthorized()
	}

	if !CheckVaultType(input.VaultType) {
		return GetVaultItemsResult{}, response.BadRequestError("vault type must be password or note")
	}

	result, err := s.vaultItemRepo.GetVaultItem(ctx, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GetVaultItemsResult{}, nil
		}
		return GetVaultItemsResult{}, response.InternalServerError("failed to get vault items")
	}

	return result, nil
}

func (s *vaultService) GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error) {
	if input.UserID == 0 {
		return VaultItem{}, response.Unauthorized()
	}
	if input.ItemID == 0 {
		return VaultItem{}, response.BadRequestError("vault item id is required")
	}

	item, err := s.vaultItemRepo.GetVaultItemByID(ctx, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	return item, nil
}

func (s *vaultService) CreateVaultItem(ctx context.Context, input CreateVaultItemInput) (VaultItem, error) {
	input = normalizeCreateInput(input)
	if err := validateCreateVaultItemInput(input); err != nil {
		return VaultItem{}, err
	}

	vault, err := s.vaultRepo.GetVaultByUserID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault")
	}

	item := VaultItem{
		ID:              s.idGen.NewID(),
		VaultID:         vault.ID,
		Title:           input.Title,
		ItemType:        input.ItemType,
		IvCiphertext:    input.IvCiphertext,
		Ciphertext:      input.Ciphertext,
		IvProtectionKey: input.IvProtectionKey,
		ProtectionKey:   input.ProtectionKey,
	}

	if err := s.vaultItemRepo.CreateVaultItem(ctx, item); err != nil {
		return VaultItem{}, response.InternalServerError("failed to create vault item")
	}

	created, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: input.UserID, ItemID: item.ID})
	if err != nil {
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	return created, nil
}

func (s *vaultService) UpdateVaultItem(ctx context.Context, input UpdateVaultItemInput) (VaultItem, error) {
	input = normalizeUpdateInput(input)
	if err := validateUpdateVaultItemInput(input); err != nil {
		return VaultItem{}, err
	}

	existing, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: input.UserID, ItemID: input.ItemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	item := VaultItem{
		ID:              existing.ID,
		VaultID:         existing.VaultID,
		Title:           input.Title,
		ItemType:        input.ItemType,
		IvCiphertext:    input.IvCiphertext,
		Ciphertext:      input.Ciphertext,
		IvProtectionKey: input.IvProtectionKey,
		ProtectionKey:   input.ProtectionKey,
	}

	if err := s.vaultItemRepo.UpdateVaultItem(ctx, item); err != nil {
		return VaultItem{}, response.InternalServerError("failed to update vault item")
	}

	updated, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: input.UserID, ItemID: input.ItemID})
	if err != nil {
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	return updated, nil
}

func (s *vaultService) ArchiveVaultItem(ctx context.Context, userID int64, itemID int64) (VaultItem, error) {
	if userID == 0 {
		return VaultItem{}, response.Unauthorized()
	}
	if itemID == 0 {
		return VaultItem{}, response.BadRequestError("vault item id is required")
	}

	existing, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	now := time.Now().UTC()
	item := existing
	item.ArchivedAt = &now

	if err := s.vaultItemRepo.ArchiveVaultItem(ctx, item); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to archive vault item")
	}

	archived, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID, Archived: true})
	if err != nil {
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	return archived, nil
}

func (s *vaultService) UnarchiveVaultItem(ctx context.Context, userID int64, itemID int64) (VaultItem, error) {
	if userID == 0 {
		return VaultItem{}, response.Unauthorized()
	}
	if itemID == 0 {
		return VaultItem{}, response.BadRequestError("vault item id is required")
	}

	existing, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID, Archived: true})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	item := existing
	item.ArchivedAt = nil

	if err := s.vaultItemRepo.UnarchiveVaultItem(ctx, item); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError("vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to unarchive vault item")
	}

	unarchived, err := s.vaultItemRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID})
	if err != nil {
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	return unarchived, nil
}

func (s *vaultService) DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error {
	if userID == 0 {
		return response.Unauthorized()
	}
	if itemID == 0 {
		return response.BadRequestError("vault item id is required")
	}

	if err := s.vaultItemRepo.DeleteVaultItem(ctx, userID, itemID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NotFoundError("vault item not found")
		}
		return response.InternalServerError("failed to delete vault item")
	}

	return nil
}

func normalizeCreateInput(input CreateVaultItemInput) CreateVaultItemInput {
	input.Title = strings.TrimSpace(input.Title)
	input.ItemType = VaultItemType(strings.TrimSpace(string(input.ItemType)))
	input.IvCiphertext = strings.TrimSpace(input.IvCiphertext)
	input.Ciphertext = strings.TrimSpace(input.Ciphertext)
	input.IvProtectionKey = strings.TrimSpace(input.IvProtectionKey)
	input.ProtectionKey = strings.TrimSpace(input.ProtectionKey)
	return input
}

func normalizeUpdateInput(input UpdateVaultItemInput) UpdateVaultItemInput {
	input.Title = strings.TrimSpace(input.Title)
	input.ItemType = VaultItemType(strings.TrimSpace(string(input.ItemType)))
	input.IvCiphertext = strings.TrimSpace(input.IvCiphertext)
	input.Ciphertext = strings.TrimSpace(input.Ciphertext)
	input.IvProtectionKey = strings.TrimSpace(input.IvProtectionKey)
	input.ProtectionKey = strings.TrimSpace(input.ProtectionKey)
	return input
}

func validateCreateVaultItemInput(input CreateVaultItemInput) error {
	if input.UserID == 0 {
		return response.Unauthorized()
	}

	return validateVaultItemFields(
		input.Title,
		input.ItemType,
		input.IvCiphertext,
		input.Ciphertext,
		input.IvProtectionKey,
		input.ProtectionKey,
	)
}

func validateUpdateVaultItemInput(input UpdateVaultItemInput) error {
	if input.UserID == 0 {
		return response.Unauthorized()
	}
	if input.ItemID == 0 {
		return response.BadRequestError("vault item id is required")
	}

	return validateVaultItemFields(
		input.Title,
		input.ItemType,
		input.IvCiphertext,
		input.Ciphertext,
		input.IvProtectionKey,
		input.ProtectionKey,
	)
}

func validateVaultItemFields(
	title string,
	itemType VaultItemType,
	ivCiphertext string,
	ciphertext string,
	ivProtectionKey string,
	protectionKey string,
) error {
	if title == "" {
		return response.BadRequestError("title is required")
	}
	if !itemType.IsValid() {
		return response.BadRequestError("itemType must be password or note")
	}

	base64Fields := map[string]string{
		"ivCiphertext":    ivCiphertext,
		"ciphertext":      ciphertext,
		"ivProtectionKey": ivProtectionKey,
		"protectionKey":   protectionKey,
	}
	for field, value := range base64Fields {
		if !isBase64(value) {
			return response.BadRequestError(fmt.Sprintf("%s must be a valid base64 string", field))
		}
	}

	return nil
}

func (t VaultItemType) IsValid() bool {
	switch t {
	case VaultItemTypePassword, VaultItemTypeNote:
		return true
	}
	return false
}

func isBase64(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	for _, encoding := range encodings {
		if _, err := encoding.DecodeString(value); err == nil {
			return true
		}
	}

	return false
}
