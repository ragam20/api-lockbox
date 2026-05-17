package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"lockbox/internal"
	appVault "lockbox/internal/app/vault"
	"lockbox/internal/config"

	"gorm.io/gorm"
)

func TestLoginRejectsInvalidVerifier(t *testing.T) {
	service := &authService{
		authRepo: &loginAuthRepo{
			user: User{
				ID:           1,
				Email:        "user@example.com",
				AuthVerifier: "correct-verifier",
			},
		},
	}

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Verifier: "wrong-verifier",
	})
	if err == nil {
		t.Fatal("expected invalid verifier to be rejected")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected generic invalid credentials error, got %q", err.Error())
	}
}

func TestLoginRejectsUnknownUserWithGenericError(t *testing.T) {
	service := &authService{
		authRepo: &loginAuthRepo{
			err: gorm.ErrRecordNotFound,
		},
	}

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "missing@example.com",
		Verifier: "any-verifier",
	})
	if err == nil {
		t.Fatal("expected missing user to be rejected")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected generic invalid credentials error, got %q", err.Error())
	}
}

func TestRefreshTokenRotatesSessionAndReturnsNewTokens(t *testing.T) {
	sessionRepo := &loginUserSessionRepo{
		session: UserSession{
			ID:               10,
			UserID:           1,
			RefreshTokenHash: internal.HashString("old-refresh-token"),
			ExpiresAt:        time.Now().Add(time.Hour),
		},
	}
	tokenManager := &loginTokenManager{
		accessToken: "new-access-token",
		secureToken: "new-refresh-token",
	}
	service := &authService{
		userSessionRepo: sessionRepo,
		tokenManager:    tokenManager,
		jwtCfg:          &config.JWTConfig{Expired: time.Hour},
	}

	result, err := service.RefreshToken(context.Background(), "old-refresh-token")
	if err != nil {
		t.Fatalf("expected refresh token to succeed: %v", err)
	}

	if result.AccessToken != "new-access-token" {
		t.Fatalf("expected new access token, got %q", result.AccessToken)
	}
	if result.RefreshToken != "new-refresh-token" {
		t.Fatalf("expected new refresh token, got %q", result.RefreshToken)
	}
	if sessionRepo.updateSessionID != 10 {
		t.Fatalf("expected session 10 to be updated, got %d", sessionRepo.updateSessionID)
	}
	if sessionRepo.updateOldHash != internal.HashString("old-refresh-token") {
		t.Fatalf("expected old hash to be checked, got %q", sessionRepo.updateOldHash)
	}
	if sessionRepo.updateNewHash != internal.HashString("new-refresh-token") {
		t.Fatalf("expected new hash to be stored, got %q", sessionRepo.updateNewHash)
	}
	if sessionRepo.session.RefreshTokenHash != internal.HashString("old-refresh-token") {
		t.Fatalf("expected stored lookup session to stay unchanged, got %q", sessionRepo.session.RefreshTokenHash)
	}
}

func TestLoginReturnsRawRefreshTokenAndStoresHash(t *testing.T) {
	sessionRepo := &loginUserSessionRepo{}
	tokenManager := &loginTokenManager{
		accessToken: "access-token",
		session: UserSession{
			ID:               10,
			RefreshToken:     "raw-refresh-token",
			RefreshTokenHash: internal.HashString("raw-refresh-token"),
			ExpiresAt:        time.Now().Add(time.Hour),
		},
	}
	service := &authService{
		authRepo: &loginAuthRepo{
			user: User{
				ID:            1,
				Email:         "user@example.com",
				Name:          "User",
				AuthVerifier:  "correct-verifier",
				MasterKeySalt: "master-key-salt",
			},
		},
		vaultRepo: &loginVaultRepo{
			vault: appVault.Vault{
				ID:           2,
				Timeout:      10,
				TimeoutType:  appVault.VaultTimeoutTypeLock,
				VaultKeySalt: "vault-key-salt",
			},
		},
		argonConfigRepo: &loginArgonConfigRepo{
			config: appVault.ArgonConfig{
				ID:          3,
				Iteration:   3,
				Memory:      64 * 1024,
				Parallelism: 2,
			},
		},
		userSessionRepo: sessionRepo,
		tokenManager:    tokenManager,
	}

	result, err := service.Login(context.Background(), LoginInput{
		Email:     "user@example.com",
		Verifier:  "correct-verifier",
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
		Device:    "Linux",
	})
	if err != nil {
		t.Fatalf("expected login to succeed: %v", err)
	}

	if result.Token.RefreshToken != "raw-refresh-token" {
		t.Fatalf("expected raw refresh token in response, got %q", result.Token.RefreshToken)
	}
	if sessionRepo.session.RefreshTokenHash != internal.HashString("raw-refresh-token") {
		t.Fatalf("expected hashed refresh token to be persisted, got %q", sessionRepo.session.RefreshTokenHash)
	}
	if sessionRepo.session.RefreshToken != "raw-refresh-token" {
		t.Fatalf("expected raw refresh token to remain on domain session, got %q", sessionRepo.session.RefreshToken)
	}
	if tokenManager.userAgent != "Mozilla/5.0" {
		t.Fatalf("expected user agent to be passed into session generation, got %q", tokenManager.userAgent)
	}
}

type loginAuthRepo struct {
	user User
	err  error
}

func (r *loginAuthRepo) WithTransaction(ctx context.Context, fn func(repo AuthRepository) error) error {
	return fn(r)
}

func (r *loginAuthRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	if r.err != nil {
		return User{}, r.err
	}
	return r.user, nil
}

func (r *loginAuthRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (r *loginAuthRepo) CreateUser(ctx context.Context, user User) error {
	return nil
}

func (r *loginAuthRepo) CreateArgonConfig(ctx context.Context, argonConfig ArgonConfig) error {
	return nil
}

func (r *loginAuthRepo) CreateVault(ctx context.Context, vault Vault) error {
	return nil
}

type loginVaultRepo struct {
	vault appVault.Vault
	err   error
}

func (r *loginVaultRepo) GetVaultByUserID(ctx context.Context, userID int64) (appVault.Vault, error) {
	return r.vault, r.err
}

type loginArgonConfigRepo struct {
	config appVault.ArgonConfig
	err    error
}

func (r *loginArgonConfigRepo) GetConfigByUserID(ctx context.Context, userID int64) (appVault.ArgonConfig, error) {
	return r.config, r.err
}

type loginUserSessionRepo struct {
	session         UserSession
	err             error
	updateSessionID int64
	updateOldHash   string
	updateNewHash   string
}

func (r *loginUserSessionRepo) CreateSession(ctx context.Context, userID int64, userSession UserSession) error {
	r.session = userSession
	r.session.UserID = userID
	return r.err
}

func (r *loginUserSessionRepo) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (UserSession, error) {
	if r.err != nil {
		return UserSession{}, r.err
	}
	if r.session.RefreshTokenHash != refreshTokenHash {
		return UserSession{}, gorm.ErrRecordNotFound
	}
	return r.session, nil
}

func (r *loginUserSessionRepo) UpdateRefreshToken(ctx context.Context, sessionID int64, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	r.updateSessionID = sessionID
	r.updateOldHash = oldRefreshTokenHash
	r.updateNewHash = newRefreshTokenHash
	if r.err != nil {
		return r.err
	}
	return nil
}

type loginTokenManager struct {
	accessToken string
	secureToken string
	session     UserSession
	userAgent   string
}

func (m *loginTokenManager) GenerateUserSession(ipAddress, userAgent, device string) (UserSession, error) {
	m.userAgent = userAgent
	m.session.IPAddress = ipAddress
	m.session.UserAgent = userAgent
	m.session.Device = device
	return m.session, nil
}

func (m *loginTokenManager) GenerateAccessToken(userID int64) (string, error) {
	if m.accessToken == "" {
		return "", errors.New("missing access token")
	}
	return m.accessToken, nil
}

func (m *loginTokenManager) ValidateAccessToken(accessToken string) (*JWTUser, error) {
	return nil, nil
}

func (m *loginTokenManager) GenerateSecureToken(byteSize int) (string, error) {
	return m.secureToken, nil
}

func (m *loginTokenManager) RegisterVerificationKey(token string) string {
	return ""
}

func (m *loginTokenManager) RegisterSessionKey(session string) string {
	return ""
}
