package application

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	appvault "lockbox/internal/app/vault"
	vaultdomain "lockbox/internal/app/vault/domain"

	dbtx "lockbox/internal/database/transaction"
	"lockbox/internal/http/response"
	"lockbox/internal/idgen"

	"gorm.io/gorm"
)

const (
	defaultArgonMemory      = 64 * 1024
	defaultArgonIterations  = 3
	defaultArgonParallelism = 3
	defaultVaultTimeout     = 10
	defaultVaultTimeoutType = VaultTimeoutTypeLock
)

type (
	VaultService interface {
		GetProtectionKeys(ctx context.Context, userID int64) ([]ItemProtectionKey, error)
		GetArgonConfig(ctx context.Context, userID int64) (ArgonConfig, error)
		SetFavorite(ctx context.Context, itemID int64) error
		DeleteFavorite(ctx context.Context, itemID int64) error
		GetVault(ctx context.Context, userID int64) (Vault, error)
		CreateVault(ctx context.Context, userID int64, vaultKeySalt string) error
		CreateArgonConfig(ctx context.Context, userID int64) error
		GetVaultItems(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error)
		GetFavorites(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error)
		GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error)
		CreateVaultItem(ctx context.Context, input CreateVaultItemInput) (VaultItem, error)
		UpdateVaultItem(ctx context.Context, input UpdateVaultItemInput) (VaultItem, error)
		ArchiveVaultItem(ctx context.Context, userID int64, itemID int64) (VaultItem, error)
		UnarchiveVaultItem(ctx context.Context, userID int64, itemID int64) (VaultItem, error)
		DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error
	}

	vaultService struct {
		vaultRepo vaultdomain.VaultRepository
		idGen     idgen.IDGenerator
		txManager dbtx.TransactionManager
	}
)

func NewVaultService(vaultRepo vaultdomain.VaultRepository, idGen idgen.IDGenerator, dbtx dbtx.TransactionManager) VaultService {
	return &vaultService{
		vaultRepo: vaultRepo,
		idGen:     idGen,
		txManager: dbtx,
	}
}

func (s *vaultService) GetProtectionKeys(ctx context.Context, userID int64) ([]ItemProtectionKey, error) {
	keys, err := s.vaultRepo.GetProtectionKeys(ctx, userID)
	if err != nil {
		return []ItemProtectionKey{}, response.InternalServerError(err.Error())
	}

	result := make([]ItemProtectionKey, 0, len(keys))
	for _, v := range keys {
		result = append(result, ItemProtectionKey{
			ID:              v.ID,
			ProtectionKey:   v.ProtectionKey,
			IvProtectionKey: v.IvProtectionKey,
		})
	}

	return result, nil
}

func (s *vaultService) SetFavorite(ctx context.Context, itemID int64) error {
	if itemID <= 0 {
		return response.BadRequestError(appvault.ErrRequiredID.Error(), "item id is required")
	}

	if err := s.vaultRepo.SetFavorite(ctx, itemID); err != nil {
		if errors.Is(err, appvault.ErrNotFound) {
			return response.BadRequestError(appvault.ErrNotFound.Error(), "vault not found")
		}
		return response.InternalServerError(err.Error())
	}

	return nil
}

func (s *vaultService) DeleteFavorite(ctx context.Context, itemID int64) error {
	if itemID <= 0 {
		return response.BadRequestError(appvault.ErrRequiredID.Error(), "item id is required")
	}

	if err := s.vaultRepo.DeleteFavorite(ctx, itemID); err != nil {
		if errors.Is(err, appvault.ErrNotFound) {
			return response.BadRequestError(appvault.ErrNotFound.Error(), "vault not found")
		}
		return response.InternalServerError(err.Error())
	}

	return nil
}

func (s *vaultService) GetVault(ctx context.Context, userID int64) (Vault, error) {
	vault, err := s.vaultRepo.GetVault(ctx, userID)
	if err != nil {
		if errors.Is(err, appvault.ErrNotFound) {
			return Vault{}, response.BadRequestError(appvault.ErrNotFound.Error(), "vault not found")
		}
		return Vault{}, response.InternalServerError(err.Error())
	}

	return vault, nil
}

func (s *vaultService) GetArgonConfig(ctx context.Context, userID int64) (ArgonConfig, error) {
	conf, err := s.vaultRepo.GetArgonConfig(ctx, userID)
	if err != nil {
		if errors.Is(err, appvault.ErrNotFound) {
			return ArgonConfig{}, response.BadRequestError(appvault.ErrNotFound.Error(), "argon config not found")
		}
		return ArgonConfig{}, response.InternalServerError(err.Error())
	}

	return conf, nil
}

func (s *vaultService) CreateVault(ctx context.Context, userID int64, vaultKeySalt string) error {
	vault := vaultdomain.NewDefaultVault(s.idGen.NewID(), vaultKeySalt)

	if err := s.vaultRepo.CreateVault(ctx, vault, userID); err != nil {
		return err
	}

	return nil
}

func (s *vaultService) CreateArgonConfig(ctx context.Context, userID int64) error {
	argonConfig := vaultdomain.NewDefaultArgonConfig(s.idGen.NewID())

	if err := s.vaultRepo.CreateArgonConfig(ctx, argonConfig, userID); err != nil {
		return err
	}

	return nil
}

func (s *vaultService) GetVaultItems(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error) {
	input, err := prepareVaultItemsInput(input)
	if err != nil {
		return GetVaultItemsResult{}, err
	}

	result, err := s.vaultRepo.GetVaultItems(ctx, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GetVaultItemsResult{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault not found")
		}
		return GetVaultItemsResult{}, response.InternalServerError("failed to get vault items")
	}

	return result, nil
}

func (s *vaultService) GetFavorites(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error) {
	input, err := prepareVaultItemsInput(input)
	if err != nil {
		return GetVaultItemsResult{}, err
	}

	result, err := s.vaultRepo.GetFavorites(ctx, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GetVaultItemsResult{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault not found")
		}
		return GetVaultItemsResult{}, response.InternalServerError("failed to get favorite vault items")
	}

	return result, nil
}

func prepareVaultItemsInput(input GetVaultInput) (GetVaultInput, error) {
	if input.UserID == 0 {
		return GetVaultInput{}, response.Unauthorized()
	}

	if !CheckVaultType(input.VaultType) {
		return GetVaultInput{}, response.BadRequestError(appvault.ErrVaultType.Error(), "vault type must be password or note")
	}

	input.Query.Search = strings.TrimSpace(input.Query.Search)

	return input, nil
}

func (s *vaultService) GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error) {
	if input.UserID == 0 {
		return VaultItem{}, response.Unauthorized()
	}
	if input.ItemID == 0 {
		return VaultItem{}, response.BadRequestError(appvault.ErrRequiredID.Error(), "vault item id is required")
	}

	item, err := s.vaultRepo.GetVaultItemByID(ctx, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
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
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault not found")
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

	if err := s.vaultRepo.CreateVaultItem(ctx, item); err != nil {
		return VaultItem{}, response.InternalServerError("failed to create vault item")
	}

	created, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: input.UserID, ItemID: item.ID})
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

	existing, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: input.UserID, ItemID: input.ItemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
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

	if err := s.vaultRepo.UpdateVaultItem(ctx, item); err != nil {
		return VaultItem{}, response.InternalServerError("failed to update vault item")
	}

	updated, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: input.UserID, ItemID: input.ItemID})
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
		return VaultItem{}, response.BadRequestError(appvault.ErrRequiredID.Error(), "vault item id is required")
	}

	existing, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	now := time.Now().UTC()
	item := existing
	item.ArchivedAt = &now

	if err := s.vaultRepo.ArchiveVaultItem(ctx, item); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to archive vault item")
	}

	archived, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID, Archived: true})
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
		return VaultItem{}, response.BadRequestError(appvault.ErrRequiredID.Error(), "vault item id is required")
	}

	existing, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID, Archived: true})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to get vault item")
	}

	item := existing
	item.ArchivedAt = nil

	if err := s.vaultRepo.UnarchiveVaultItem(ctx, item); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VaultItem{}, response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
		}
		return VaultItem{}, response.InternalServerError("failed to unarchive vault item")
	}

	unarchived, err := s.vaultRepo.GetVaultItemByID(ctx, GetVaultItemByIDInput{UserID: userID, ItemID: itemID})
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
		return response.BadRequestError(appvault.ErrRequiredID.Error(), "vault item id is required")
	}

	if err := s.vaultRepo.DeleteVaultItem(ctx, userID, itemID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NotFoundError(appvault.ErrNotFound.Error(), "vault item not found")
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
		return response.BadRequestError(appvault.ErrRequiredID.Error(), "vault item id is required")
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
		return response.BadRequestError(appvault.ErrFieldTitle.Error(), "title is required")
	}
	if !itemType.IsValid() {
		return response.BadRequestError(appvault.ErrVaultType.Error(), "itemType must be password or note")
	}

	base64Fields := map[string]string{
		"ivCiphertext":    ivCiphertext,
		"ciphertext":      ciphertext,
		"ivProtectionKey": ivProtectionKey,
		"protectionKey":   protectionKey,
	}
	for field, value := range base64Fields {
		if !isBase64(value) {
			return response.BadRequestError(appvault.ErrInvalidInput.Error(), fmt.Sprintf("%s must be a valid base64 string", field))
		}
	}

	return nil
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
