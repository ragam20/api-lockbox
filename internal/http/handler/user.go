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
