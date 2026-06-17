package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"lockbox/internal"
	"lockbox/internal/config"
	"lockbox/internal/http/response"
	"lockbox/internal/idgen"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUser struct {
	UserID int64
	Claims *jwt.MapClaims
}

type TokenManager interface {
	GenerateUserSession(ipAddress, userAgent, device string) (UserSession, error)
	GenerateAccessToken(userID int64) (string, error)
	ValidateAccessToken(accessToken string) (*JWTUser, error)
	GenerateSecureToken(byteSize int) (string, error)
	RegisterVerificationKey(token string) string
	RegisterSessionKey(session string) string
}

type tokenManager struct {
	idgen idgen.IDGenerator
	cfg   *config.JWTConfig
}

const (
	registerVerificationKeyPrefix = "register_verification"
	registerSessionKeyPrefix      = "register_session"
)

func NewTokenManager(idgen idgen.IDGenerator, cfg *config.JWTConfig) TokenManager {
	return &tokenManager{
		idgen: idgen,
		cfg:   cfg,
	}
}

func (m *tokenManager) GenerateUserSession(ipAddress, userAgent, device string) (UserSession, error) {
	secureToken, err := m.GenerateSecureToken(32)
	if err != nil {
		return UserSession{}, err
	}

	return UserSession{
		ID:               m.idgen.NewID(),
		RefreshToken:     secureToken,
		RefreshTokenHash: internal.HashString(secureToken),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Device:           device,
		LastUsedAt:       time.Now(),
		ExpiresAt:        time.Now().Add(m.cfg.RefreshExp),
	}, nil
}

func (m *tokenManager) GenerateSecureToken(byteSize int) (string, error) {
	if byteSize <= 0 {
		return "", fmt.Errorf("token byte size must be positive")
	}

	token := make([]byte, byteSize)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}

func (m *tokenManager) RegisterVerificationKey(token string) string {
	return fmt.Sprintf("%s:%s", registerVerificationKeyPrefix, token)
}

func (m *tokenManager) RegisterSessionKey(session string) string {
	return fmt.Sprintf("%s:%s", registerSessionKeyPrefix, session)
}

func (m *tokenManager) GenerateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(m.cfg.Expired).Unix(),
	}

	return m.generateJWT(claims, m.cfg.Secret)
}

func (m *tokenManager) ValidateAccessToken(accessToken string) (*JWTUser, error) {
	claims, err := m.validateJWT(accessToken, m.cfg.Secret)
	if err != nil {
		return nil, response.Unauthorized()
	}

	if !m.checkClaims(claims, "userId", "exp") {
		return nil, response.BadRequestError(ErrTokenExpired.Error(), "invalid jwt")
	}

	uidFloat, ok := claims["userId"].(float64)
	if !ok {
		return nil, response.BadRequestError(ErrInvalidToken.Error(), "invalid userId claim")
	}

	return &JWTUser{
		UserID: int64(uidFloat),
		Claims: &claims,
	}, nil
}

func (m *tokenManager) generateJWT(jwtClaims jwt.MapClaims, secret string) (string, error) {
	t := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwtClaims,
	)
	return t.SignedString([]byte(secret))
}

func (m *tokenManager) validateJWT(currToken string, secret string) (jwt.MapClaims, error) {
	j, err := jwt.Parse(currToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, response.BadRequestError(ErrInvalidToken.Error(), "invalid jwt signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !j.Valid {
		return nil, response.BadRequestError(ErrInvalidToken.Error(), "invalid jwt")
	}

	claims, ok := j.Claims.(jwt.MapClaims)
	if !ok {
		return nil, response.BadRequestError(ErrInvalidToken.Error(), "invalid jwt claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, response.BadRequestError(ErrInvalidToken.Error(), "invalid jwt exp claim")
	}

	if int64(exp) < time.Now().Unix() {
		return nil, response.BadRequestError(ErrTokenExpired.Error(), "token expired")
	}

	return claims, nil
}

func (m *tokenManager) checkClaims(claims jwt.MapClaims, keys ...string) bool {
	for _, key := range keys {
		if _, ok := claims[key]; !ok {
			return false
		}
	}
	return true
}
