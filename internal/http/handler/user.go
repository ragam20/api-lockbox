package handler

import (
	"net/http"

	"lockbox/internal/app/user"
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
