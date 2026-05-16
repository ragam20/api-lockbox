package handler

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"lockbox/internal/app/auth"
	"lockbox/internal/config"
	apphttp "lockbox/internal/http"
	"lockbox/internal/http/request"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
	"github.com/mssola/useragent"
)

type AuthHandler struct {
	authService   auth.AuthService
	cookieManager *apphttp.CookieManager
	jwtCfg        *config.JWTConfig
}

func NewAuthHandler(
	authService auth.AuthService,
	cookieManager *apphttp.CookieManager,
	jwtCfg *config.JWTConfig,
) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		cookieManager: cookieManager,
		jwtCfg:        jwtCfg,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError("invalid request body"))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), auth.LoginInput{
		Email:     req.Email,
		Verifier:  req.Verifier,
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
		Device:    loginDevice(c.Request.UserAgent()),
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	expiredAt := time.Now().Add(h.jwtCfg.Expired)
	http.SetCookie(c.Writer, h.cookieManager.SetAccessToken(result.Token.AccessToken, expiredAt))
	http.SetCookie(c.Writer, h.cookieManager.SetRefreshToken(result.Token.RefreshToken, expiredAt))

	response.Success(c, http.StatusOK, "login success", result)
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

func loginDevice(userAgentHeader string) string {
	userAgentHeader = strings.TrimSpace(userAgentHeader)
	if userAgentHeader == "" {
		return ""
	}

	ua := useragent.New(userAgentHeader)
	parts := make([]string, 0, 3)

	appendPart := func(part string) {
		part = strings.TrimSpace(part)
		if part == "" {
			return
		}
		for _, existing := range parts {
			if strings.EqualFold(existing, part) {
				return
			}
		}
		parts = append(parts, part)
	}

	appendPart(ua.Model())
	appendPart(ua.Platform())

	osInfo := ua.OSInfo()
	osName := strings.TrimSpace(osInfo.Name)
	if osName != "" && strings.TrimSpace(osInfo.Version) != "" {
		osName = osName + " " + strings.TrimSpace(osInfo.Version)
	}
	appendPart(osName)

	if len(parts) == 0 {
		browser, _ := ua.Browser()
		appendPart(browser)
	}

	if len(parts) == 0 {
		return userAgentHeader
	}

	return strings.Join(parts, " - ")
}
