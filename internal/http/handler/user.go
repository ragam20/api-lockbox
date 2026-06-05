package handler

import (
	"net/http"

	"lockbox/internal/app/user"
	apphttp "lockbox/internal/http"
	"lockbox/internal/http/request"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService user.UserService
}

func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RemoveAvatar(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(response.Unauthorized())
		return
	}

	err = h.userService.RemoveAvatar(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success remove avatar", true)
}

func (h *UserHandler) SaveAvatarURL(c *gin.Context) {
	id, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(response.Unauthorized())
		return
	}

	var req request.SaveAvatarRequest
	if err := c.ShouldBind(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
		return
	}

	if err := h.userService.SaveAvatarURL(c.Request.Context(), user.SaveAvatarURLInput{
		UserID:      id,
		PublicImgID: req.PublicImgID,
		ImgURL:      req.ImgURL,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success save url", true)
}

func (h *UserHandler) UploadAvatarPermission(c *gin.Context) {
	id, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(response.Unauthorized())
		return
	}

	resp, err := h.userService.UploadAvatarRequest(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success upload avatar request", resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	id, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(response.Unauthorized())
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid body request"))
		return
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), user.UpdateUserInput{
		UserID: id,
		Name:   req.Name,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success update profile", true)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(response.Unauthorized())
		return
	}

	usr, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success get user", response.UserResponse{
		ID:               usr.ID,
		Name:             usr.Name,
		Email:            usr.Email,
		ImgURL:           usr.ImgURL,
		MasterKeySalt:    usr.MasterKeySalt,
		AuthVerifierSalt: usr.AuthVerifierSalt,
		CreatedAt:        usr.CreatedAt,
		UpdatedAt:        usr.UpdatedAt,
	})
}
