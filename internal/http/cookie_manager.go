package apphttp

import (
	"net/http"
	"time"

	"lockbox/internal/config"
)

const (
	refreshTokenCookie = "refreshToken"
	accessTokenCookie  = "authSession"
)

type CookieManager struct {
	appCfg *config.AppConfig
}

func NewCookieManager(appCfg *config.AppConfig) *CookieManager {
	return &CookieManager{appCfg: appCfg}
}

func (m *CookieManager) SetAccessToken(accessToken string, expiredAt time.Time) *http.Cookie {
	return m.newTokenCookie(accessTokenCookie, accessToken, expiredAt)
}

func (m *CookieManager) RevokeAccessToken() *http.Cookie {
	return m.newExpiredTokenCookie(accessTokenCookie)
}

func (m *CookieManager) SetRefreshToken(refreshToken string, expiredAt time.Time) *http.Cookie {
	return m.newTokenCookie(refreshTokenCookie, refreshToken, expiredAt)
}

func (m *CookieManager) RevokeRefreshToken() *http.Cookie {
	return m.newExpiredTokenCookie(refreshTokenCookie)
}

func AccessTokenCookieName() string {
	return accessTokenCookie
}

func RefreshTokenCookieName() string {
	return refreshTokenCookie
}

func (m *CookieManager) newTokenCookie(name, value string, expiredAt time.Time) *http.Cookie {
	maxAge := max(int(time.Until(expiredAt).Seconds()), 0)

	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Secure:   m.isSecure(),
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: m.sameSite(),
	}
}

func (m *CookieManager) newExpiredTokenCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Secure:   m.isSecure(),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: m.sameSite(),
	}
}

func (m *CookieManager) isSecure() bool {
	return m.appCfg.Env != "development"
}

func (m *CookieManager) sameSite() http.SameSite {
	if m.appCfg.Env == "production" || m.appCfg.Env == "prod" {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}
