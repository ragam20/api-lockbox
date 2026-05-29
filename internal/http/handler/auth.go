package handler

import (
	"net/http"
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
	appConfig     *config.AppConfig
}

func NewAuthHandler(
	authService auth.AuthService,
	cookieManager *apphttp.CookieManager,
	jwtCfg *config.JWTConfig,
	appConfig *config.AppConfig,
) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		cookieManager: cookieManager,
		jwtCfg:        jwtCfg,
		appConfig:     appConfig,
	}
}

func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req request.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
		return
	}

	resp, err := h.authService.RequestOTP(c.Request.Context(), req.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "otp sent", resp)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req request.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
		return
	}

	result, err := h.authService.VerifyOTP(c.Request.Context(), auth.VerifyOTPInput{
		Email:     req.Email,
		OTP:       req.OTP,
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
		Device:    loginDevice(c.Request.UserAgent()),
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	h.setAuthCookies(c, result.Token)

	response.Success(c, http.StatusOK, "otp verified", result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	refreshToken, err := c.Cookie(apphttp.RefreshTokenCookieName())
	if err != nil {
		var reqBody struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			response.Success(c, http.StatusOK, "success logout", true)
			return
		}

		refreshToken = reqBody.RefreshToken
	}

	c.SetCookieData(h.cookieManager.RevokeAccessToken())
	c.SetCookieData(h.cookieManager.RevokeRefreshToken())

	h.authService.Logout(c.Request.Context(), userID, refreshToken)

	response.Success(c, http.StatusOK, "success logout", true)
}

func (h *AuthHandler) Unlock(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), auth.LoginInput{
		Email:    req.Email,
		Verifier: req.Verifier,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "login success", result)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(apphttp.RefreshTokenCookieName())
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		_ = c.Error(response.Unauthorized())
		return
	}

	token, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}

	h.setAuthCookies(c, token)

	response.Success(c, http.StatusOK, "refresh token success", token)
}

func (h *AuthHandler) GetRegisterSession(c *gin.Context) {
	var req request.RegisterSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	resp, err := h.authService.GetRegisterSession(c.Request.Context(), req.Session)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "success get register session", resp)
}

func (h *AuthHandler) SendVerifyLink(c *gin.Context) {
	var req request.SendVerifyLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
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

	response.Success(c, http.StatusOK, "success verification email", response.VerifyEmailResponse{
		Session: session,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(response.BadRequestError(apphttp.ErrTransport.Error(), "invalid request body"))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), auth.RegisterInput{
		Session:          c.Query("session"),
		Email:            req.Email,
		Name:             req.Name,
		AuthVerifier:     req.AuthVerifier,
		AuthVerifierSalt: req.AuthVerifierSalt,
		MasterKeySalt:    req.MasterKeySalt,
		VaultKeySalt:     req.VaultKeySalt,
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

func (h *AuthHandler) setAuthCookies(c *gin.Context, token response.TokenResponse) {
	expiredAt := time.Now().Add(h.jwtCfg.Expired)
	http.SetCookie(c.Writer, h.cookieManager.SetAccessToken(token.AccessToken, expiredAt))
	http.SetCookie(c.Writer, h.cookieManager.SetRefreshToken(token.RefreshToken, expiredAt))
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
