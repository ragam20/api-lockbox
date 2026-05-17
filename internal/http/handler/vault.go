package handler

import (
	"net/http"
	"strconv"

	"lockbox/internal/app/vault"
	"lockbox/internal/http/request"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type VaultHandler struct {
	vaultService vault.VaultService
}

func NewVaultHandler(vaultService vault.VaultService) *VaultHandler {
	return &VaultHandler{vaultService: vaultService}
}

func (h *VaultHandler) GetVaultItem(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	pagination, err := GetPagination(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	vaultType := c.Query("type")
	if !vault.CheckVaultType(vaultType) {
		_ = c.Error(response.BadRequestError("vault type must be password or note"))
		return
	}

	archive, err := archiveQuery(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	sortBy := c.Query("sortBy")

	resp, err := h.vaultService.GetVaultItem(c.Request.Context(), vault.GetVaultInput{
		Query: vault.VaultItemQuery{
			Page:   pagination.Page,
			Limit:  pagination.Limit,
			SortBy: sortBy,
		},
		VaultType: vaultType,
		UserID:    userID,
		Archived:  archive,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	limit := pagination.Limit
	totalPages := int64(0)
	if limit > 0 {
		totalPages = (resp.TotalItems + int64(limit) - 1) / int64(limit)
	}

	response.SuccessWithMeta(
		c,
		http.StatusOK,
		"success get vault items",
		vaultItemsResponse(resp.Items),
		response.Meta{
			"page":       pagination.Page,
			"limit":      pagination.Limit,
			"totalItems": resp.TotalItems,
			"totalPages": totalPages,
		},
	)
}

func (h *VaultHandler) GetVaultItemByID(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	archive, err := archiveQuery(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.vaultService.GetVaultItemByID(c.Request.Context(), vault.GetVaultItemByIDInput{
		UserID:   userID,
		ItemID:   itemID,
		Archived: archive,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success get vault item", vaultItemResponse(item))
}

func (h *VaultHandler) CreateVaultItem(c *gin.Context) {
	var req request.VaultItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError("invalid request body"))
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.vaultService.CreateVaultItem(c.Request.Context(), vault.CreateVaultItemInput{
		UserID:          userID,
		Title:           req.Title,
		ItemType:        vault.VaultItemType(req.ItemType),
		IvCiphertext:    req.IvCiphertext,
		Ciphertext:      req.Ciphertext,
		IvProtectionKey: req.IvProtectionKey,
		ProtectionKey:   req.ProtectionKey,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, "vault item created", vaultItemResponse(item))
}

func (h *VaultHandler) UpdateVaultItem(c *gin.Context) {
	var req request.VaultItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError("invalid request body"))
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.vaultService.UpdateVaultItem(c.Request.Context(), vault.UpdateVaultItemInput{
		UserID:          userID,
		ItemID:          itemID,
		Title:           req.Title,
		ItemType:        vault.VaultItemType(req.ItemType),
		IvCiphertext:    req.IvCiphertext,
		Ciphertext:      req.Ciphertext,
		IvProtectionKey: req.IvProtectionKey,
		ProtectionKey:   req.ProtectionKey,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "vault item updated", vaultItemResponse(item))
}

func (h *VaultHandler) ArchiveVaultItem(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.vaultService.ArchiveVaultItem(c.Request.Context(), userID, itemID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "vault item archived", vaultItemResponse(item))
}

func (h *VaultHandler) UnarchiveVaultItem(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.vaultService.UnarchiveVaultItem(c.Request.Context(), userID, itemID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "vault item unarchived", vaultItemResponse(item))
}

func (h *VaultHandler) DeleteVaultItem(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.vaultService.DeleteVaultItem(c.Request.Context(), userID, itemID); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "vault item deleted", nil)
}

func vaultItemIDFromParam(c *gin.Context) (int64, error) {
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || itemID <= 0 {
		return 0, response.BadRequestError("invalid vault item id")
	}

	return itemID, nil
}

func vaultItemResponse(item vault.VaultItem) response.VaultItemResponse {
	return response.VaultItemResponse{
		ID:              item.ID,
		VaultID:         item.VaultID,
		Title:           item.Title,
		ItemType:        string(item.ItemType),
		IvCiphertext:    item.IvCiphertext,
		Ciphertext:      item.Ciphertext,
		IvProtectionKey: item.IvProtectionKey,
		ProtectionKey:   item.ProtectionKey,
		ArchivedAt:      item.ArchivedAt,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func vaultItemsResponse(items []vault.VaultItem) []response.VaultItemResponse {
	res := make([]response.VaultItemResponse, len(items))
	for i, item := range items {
		res[i] = vaultItemResponse(item)
	}
	return res
}
