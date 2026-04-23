package auth

import (
	"time"

	"lockbox/internal/config"
	"lockbox/internal/http/response"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUser struct {
	UserID int64
	Claims *jwt.MapClaims
}

type JWTManager interface {
	GenerateAccessToken(userID int64) (string, error)
	ValidateAccessToken(accessToken string) (*JWTUser, error)
}

type jwtManager struct {
	cfg *config.JWTConfig
}

func NewJWTManager(cfg *config.JWTConfig) JWTManager {
	return &jwtManager{
		cfg: cfg,
	}
}

func (m *jwtManager) GenerateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(m.cfg.Expired).Unix(),
	}

	return m.generateJWT(claims, m.cfg.Secret)
}

func (m *jwtManager) ValidateAccessToken(accessToken string) (*JWTUser, error) {
	claims, err := m.validateJWT(accessToken, m.cfg.Secret)
	if err != nil {
		return nil, response.Unauthorized()
	}

	if !m.checkClaims(claims, "userId", "exp") {
		return nil, response.BadRequestError("invalid jwt")
	}

	uidFloat, ok := claims["userId"].(float64)
	if !ok {
		return nil, response.BadRequestError("invalid userId claim")
	}

	return &JWTUser{
		UserID: int64(uidFloat),
		Claims: &claims,
	}, nil
}

func (m *jwtManager) generateJWT(jwtClaims jwt.MapClaims, secret string) (string, error) {
	t := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwtClaims,
	)
	return t.SignedString([]byte(secret))
}

func (m *jwtManager) validateJWT(currToken string, secret string) (jwt.MapClaims, error) {
	j, err := jwt.Parse(currToken, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !j.Valid {
		return nil, response.BadRequestError("invalid jwt")
	}

	claims, _ := j.Claims.(jwt.MapClaims)

	if int64(claims["exp"].(float64)) < time.Now().Unix() {
		return nil, response.BadRequestError("token expired")
	}

	return claims, nil
}

func (m *jwtManager) checkClaims(claims jwt.MapClaims, keys ...string) bool {
	for _, key := range keys {
		if _, ok := claims[key]; !ok {
			return false
		}
	}
	return true
}
