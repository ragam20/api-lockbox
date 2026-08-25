package handler

import (
	"net/http"
	"strconv"

	"lockbox/internal/app/vault"
	argonconfig "lockbox/internal/app/workflow/update_argon"
	apphttp "lockbox/internal/http"
	"lockbox/internal/http/request"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type VaultHandler struct {
	vaultService      vault.VaultService
	updateArgonConfig argonconfig.UpdateArgonConfig
}

func NewVaultHandler(vaultService vault.VaultService, updateArgonConfig argonconfig.UpdateArgonConfig) *VaultHandler {
	return &VaultHandler{vaultService: vaultService, updateArgonConfig: updateArgonConfig}
}

func (h *VaultHandler) UpdateArgonConfig(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var req request.UpdateArgonConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
		return
	}

	protectionKeys := make([]vault.ItemProtectionKey, 0, len(req.ProtectionKeys))
	for _, v := range req.ProtectionKeys {
		protectionKeys = append(protectionKeys, vault.ItemProtectionKey{
			ID:              v.ItemID,
			IvProtectionKey: v.IvProtectionKey,
			ProtectionKey:   v.ProtectionKey,
		})
	}

	err = h.updateArgonConfig.Execute(
		c.Request.Context(),
		vault.UpdateArgonConfigInput{
			UserID:       userID,
			AuthVerifier: req.AuthVerifier,
			ArgonConfig: vault.ArgonConfig{
				ID:          req.ArgonConfig.ID,
				Iterations:  uint32(req.ArgonConfig.Iterations),
				Memory:      uint32(req.ArgonConfig.Memory),
				Parallelism: uint32(req.ArgonConfig.Parallelism),
			},
			ProtectionKeys: protectionKeys,
		},
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success update argon config", true)
}

func (h *VaultHandler) GetProtectionKeys(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	keys, err := h.vaultService.GetProtectionKeys(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := make([]response.ItemProtectionKeyResponse, 0, len(keys))
	for _, v := range keys {
		resp = append(resp, response.ItemProtectionKeyResponse{
			ID:              v.ID,
			ProtectionKey:   v.ProtectionKey,
			IvProtectionKey: v.IvProtectionKey,
		})
	}

	response.Success(c, http.StatusOK, "success get protectionKeys", resp)
}

func (h *VaultHandler) SetFavorite(c *gin.Context) {
	_, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.vaultService.SetFavorite(c.Request.Context(), itemID); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success set favorite", true)
}

func (h *VaultHandler) DeleteFavorite(c *gin.Context) {
	_, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemID, err := vaultItemIDFromParam(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.vaultService.DeleteFavorite(c.Request.Context(), itemID); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success delete favorite", true)
}

func (h *VaultHandler) GetVault(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	vault, err := h.vaultService.GetVault(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success get vault", response.VaultResponse{
		ID:           vault.ID,
		Timeout:      vault.Timeout,
		TimeoutType:  string(vault.TimeoutType),
		VaultKeySalt: vault.VaultKeySalt,
		CreatedAt:    vault.CreatedAt,
		UpdatedAt:    vault.UpdatedAt,
	})
}

func (h *VaultHandler) GetArgonConfig(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	m, err := h.vaultService.GetArgonConfig(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success get argon config", response.ArgonConfigResponse{
		ID:          m.ID,
		Iteration:   m.Iterations,
		Memory:      m.Memory,
		Parallelism: m.Parallelism,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	})
}

func (h *VaultHandler) GetVaultItems(c *gin.Context) {
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
		_ = c.Error(response.BadRequestError(vault.ErrVaultType.Error(), "vault type must be password or note"))
		return
	}

	archive, err := archiveQuery(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	resp, err := h.vaultService.GetVaultItems(c.Request.Context(), vault.GetVaultInput{
		Query: vault.VaultItemQuery{
			Page:   pagination.Page,
			Limit:  pagination.Limit,
			SortBy: c.Query("sortBy"),
			Search: c.Query("search"),
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

func (h *VaultHandler) GetFavorites(c *gin.Context) {
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
	if vaultType == "" {
		vaultType = string(vault.VaultItemTypeAll)
	}
	if !vault.CheckVaultType(vaultType) {
		_ = c.Error(response.BadRequestError(vault.ErrVaultType.Error(), "vault type must be password or note"))
		return
	}

	archive, err := archiveQuery(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp, err := h.vaultService.GetFavorites(c.Request.Context(), vault.GetVaultInput{
		Query: vault.VaultItemQuery{
			Page:   pagination.Page,
			Limit:  pagination.Limit,
			SortBy: c.Query("sortBy"),
			Search: c.Query("search"),
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
		"success get favorite vault items",
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
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
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
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
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
		return 0, response.BadRequestError(apphttp.ErrTransport.Error(), "invalid vault item id")
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
		FavoriteAt:      item.FavoriteAt,
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
