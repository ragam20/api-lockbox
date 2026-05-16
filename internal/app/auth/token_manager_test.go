package auth

import (
	"testing"
	"time"

	"lockbox/internal"
	"lockbox/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateUserSessionStoresRefreshTokenHash(t *testing.T) {
	manager := NewTokenManager(staticIDGen{id: 1}, &config.JWTConfig{
		Secret:  "secret",
		Expired: time.Hour,
	})

	session, err := manager.GenerateUserSession("127.0.0.1", "Mozilla/5.0", "Linux")
	if err != nil {
		t.Fatalf("expected session generation to succeed: %v", err)
	}

	if session.RefreshToken == "" {
		t.Fatal("expected raw refresh token")
	}
	if session.RefreshTokenHash == session.RefreshToken {
		t.Fatal("expected stored refresh token hash to differ from raw refresh token")
	}
	if session.RefreshTokenHash != internal.HashString(session.RefreshToken) {
		t.Fatalf("expected refresh token hash %q, got %q", internal.HashString(session.RefreshToken), session.RefreshTokenHash)
	}
}

func TestValidateJWTRejectsInvalidClaimsWithoutPanic(t *testing.T) {
	manager := &tokenManager{cfg: &config.JWTConfig{Secret: "secret", Expired: time.Hour}}

	tests := map[string]jwt.MapClaims{
		"missing exp":    {"userId": float64(1)},
		"wrong exp type": {"userId": float64(1), "exp": "not-a-number"},
		"expired":        {"userId": float64(1), "exp": float64(time.Now().Add(-time.Hour).Unix())},
	}

	for name, claims := range tests {
		t.Run(name, func(t *testing.T) {
			token := signedTestToken(t, claims, "secret")

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("validateJWT panicked: %v", r)
				}
			}()

			if _, err := manager.validateJWT(token, "secret"); err == nil {
				t.Fatal("expected invalid token to be rejected")
			}
		})
	}
}

func TestValidateJWTAcceptsValidToken(t *testing.T) {
	manager := &tokenManager{cfg: &config.JWTConfig{Secret: "secret", Expired: time.Hour}}
	token := signedTestToken(t, jwt.MapClaims{
		"userId": float64(1),
		"exp":    float64(time.Now().Add(time.Hour).Unix()),
	}, "secret")

	claims, err := manager.validateJWT(token, "secret")
	if err != nil {
		t.Fatalf("expected valid token: %v", err)
	}
	if claims["userId"] != float64(1) {
		t.Fatalf("expected userId claim, got %v", claims["userId"])
	}
}

func signedTestToken(t *testing.T, claims jwt.MapClaims, secret string) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return token
}

type staticIDGen struct {
	id int64
}

func (g staticIDGen) NewID() int64 {
	return g.id
}
