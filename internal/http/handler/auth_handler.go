package handler

import (
	"net/http"
	"net/url"

	"lockbox/internal/app/auth"
	"lockbox/internal/http/request"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService auth.AuthService
}

func NewHandler(authService auth.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SendVerifyLink(c *gin.Context) {
	var req request.SendVerifyLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError("invalid request body"))
		return
	}

	if err := h.authService.SendVerifyLink(c.Request.Context(), auth.SendVerificationLinkInput{
		Email: req.Email,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "verification link sent", nil)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	session, err := h.authService.VerifyEmail(c.Request.Context(), c.Query("token"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Redirect(http.StatusFound, "/register/complete?session="+url.QueryEscape(session))
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError("invalid request body"))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), auth.RegisterInput{
		Session:          c.Query("session"),
		Email:            req.Email,
		Name:             req.Name,
		AuthVerifier:     req.AuthVerifier,
		AuthVerifierSalt: req.AuthVerifierSalt,
		MasterKeySalt:    req.MasterKeySalt,
		Vault: auth.VaultInput{
			VaultKeySalt: req.Vault.VaultKeySalt,
		},
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, "registration completed", response.RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}
