package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"lockbox/internal"
	"lockbox/internal/config"
	appResponse "lockbox/internal/http/response"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func TestLoginRejectsInvalidVerifier(t *testing.T) {
	service := &authService{
		authRepo: &mockAuthRepo{
			user: User{
				ID:           1,
				Email:        "user@example.com",
				AuthVerifier: "correct-verifier",
			},
		},
		tokenManager: &mockTokenManager{},
	}

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Verifier: "wrong-verifier",
	})
	if err == nil {
		t.Fatal("expected invalid verifier to be rejected")
	}

	appErr := mustAppErr(t, err)
	if appErr.Code != ErrInvalidCredential.Error() {
		t.Fatalf("expected invalid credential code, got %q", appErr.Code)
	}
	if appErr.Message != "invalid credentials" {
		t.Fatalf("expected generic invalid credentials message, got %q", appErr.Message)
	}
}

func TestLoginRejectsUnknownUserWithInvalidCredentialCode(t *testing.T) {
	service := &authService{
		authRepo:     &mockAuthRepo{getByEmailErr: gorm.ErrRecordNotFound},
		tokenManager: &mockTokenManager{},
	}

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "missing@example.com",
		Verifier: "any-verifier",
	})
	if err == nil {
		t.Fatal("expected missing user to be rejected")
	}

	appErr := mustAppErr(t, err)
	if appErr.Code != ErrInvalidCredential.Error() {
		t.Fatalf("expected invalid credential code, got %q", appErr.Code)
	}
	if appErr.Message != "invalid credentials" {
		t.Fatalf("expected generic invalid credentials message, got %q", appErr.Message)
	}
}

func TestRefreshTokenRotatesSessionAndReturnsNewTokens(t *testing.T) {
	store := newMockCache()
	tokenManager := &mockTokenManager{
		accessToken: "new-access-token",
		secureToken: "new-refresh-token",
	}
	service := &authService{
		authRepo: &mockAuthRepo{
			session: UserSession{
				ID:               10,
				UserID:           1,
				RefreshTokenHash: internal.HashString("old-refresh-token"),
				ExpiresAt:        time.Now().Add(time.Hour),
			},
		},
		tokenManager: tokenManager,
		jwtCfg:       &config.JWTConfig{Expired: time.Hour},
		cache:        store,
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

	repo := service.authRepo.(*mockAuthRepo)
	if repo.updateSessionID != 10 {
		t.Fatalf("expected session 10 to be updated, got %d", repo.updateSessionID)
	}
	if repo.updateOldHash != internal.HashString("old-refresh-token") {
		t.Fatalf("expected old hash to be checked, got %q", repo.updateOldHash)
	}
	if repo.updateNewHash != internal.HashString("new-refresh-token") {
		t.Fatalf("expected new hash to be stored, got %q", repo.updateNewHash)
	}
}

type mockAuthRepo struct {
	user             User
	getByEmailErr    error
	emailExists      bool
	emailExistsErr   error
	createUserErr    error
	createSessionErr error
	session          UserSession
	sessionErr       error
	updateErr        error

	createdUser      User
	createdSession   UserSession
	createdSessionID int64

	updateSessionID int64
	updateOldHash   string
	updateNewHash   string
	updateExpiresAt time.Time
	updateLastUsed  time.Time

	revokedUserID    int64
	revokedTokenHash string

	withTxCalls int
}

func (r *mockAuthRepo) UpdateAuthVerifier(ctx context.Context, authVerifier string, userID int64) error {
	return nil
}

func (r *mockAuthRepo) WithTx(tx *gorm.DB) AuthRepository {
	r.withTxCalls++
	return r
}

func (r *mockAuthRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	if r.getByEmailErr != nil {
		return User{}, r.getByEmailErr
	}
	return r.user, nil
}

func (r *mockAuthRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.emailExists, r.emailExistsErr
}

func (r *mockAuthRepo) CreateUser(ctx context.Context, user User) error {
	r.createdUser = user
	return r.createUserErr
}

func (r *mockAuthRepo) RevokeSession(ctx context.Context, userID int64, tokenHash string) error {
	r.revokedUserID = userID
	r.revokedTokenHash = tokenHash
	return nil
}

func (r *mockAuthRepo) CreateSession(ctx context.Context, userID int64, userSession UserSession) error {
	r.createdSessionID = userID
	r.createdSession = userSession
	r.createdSession.UserID = userID
	return r.createSessionErr
}

func (r *mockAuthRepo) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (UserSession, error) {
	if r.sessionErr != nil {
		return UserSession{}, r.sessionErr
	}
	if r.session.RefreshTokenHash != refreshTokenHash {
		return UserSession{}, gorm.ErrRecordNotFound
	}
	return r.session, nil
}

func (r *mockAuthRepo) UpdateRefreshToken(ctx context.Context, sessionID int64, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	r.updateSessionID = sessionID
	r.updateOldHash = oldRefreshTokenHash
	r.updateNewHash = newRefreshTokenHash
	r.updateExpiresAt = expiresAt
	r.updateLastUsed = lastUsedAt
	return r.updateErr
}

type mockTokenManager struct {
	accessToken string
	secureToken string

	generateAccessTokenErr error
	generateUserSessionErr error
	generateSecureTokenErr error

	generateUserSessionArgs struct {
		ipAddress string
		userAgent string
		device    string
	}
	generateSecureTokenByteSize int
}

func (m *mockTokenManager) GenerateUserSession(ipAddress, userAgent, device string) (UserSession, error) {
	m.generateUserSessionArgs.ipAddress = ipAddress
	m.generateUserSessionArgs.userAgent = userAgent
	m.generateUserSessionArgs.device = device
	if m.generateUserSessionErr != nil {
		return UserSession{}, m.generateUserSessionErr
	}
	if m.secureToken == "" {
		m.secureToken = "refresh-token"
	}
	return UserSession{
		ID:               10,
		RefreshToken:     m.secureToken,
		RefreshTokenHash: internal.HashString(m.secureToken),
		ExpiresAt:        time.Now().Add(time.Hour),
		LastUsedAt:       time.Now().UTC(),
	}, nil
}

func (m *mockTokenManager) GenerateAccessToken(userID int64) (string, error) {
	if m.generateAccessTokenErr != nil {
		return "", m.generateAccessTokenErr
	}
	if m.accessToken == "" {
		return "access-token", nil
	}
	return m.accessToken, nil
}

func (m *mockTokenManager) ValidateAccessToken(accessToken string) (*JWTUser, error) {
	return nil, nil
}

func (m *mockTokenManager) GenerateSecureToken(byteSize int) (string, error) {
	m.generateSecureTokenByteSize = byteSize
	if m.generateSecureTokenErr != nil {
		return "", m.generateSecureTokenErr
	}
	if m.secureToken == "" {
		return "secure-token", nil
	}
	return m.secureToken, nil
}

func (m *mockTokenManager) RegisterVerificationKey(token string) string {
	return "register_verification:" + token
}

func (m *mockTokenManager) RegisterSessionKey(session string) string {
	return "register_session:" + session
}

type mockCache struct {
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (c *mockCache) Set(ctx context.Context, key string, value string, exp time.Duration) error {
	c.data[key] = value
	return nil
}

func (c *mockCache) Get(ctx context.Context, key string) (string, error) {
	value, ok := c.data[key]
	if !ok {
		return "", redis.Nil
	}
	return value, nil
}

func (c *mockCache) GetDel(ctx context.Context, key string) (string, error) {
	value, ok := c.data[key]
	if !ok {
		return "", redis.Nil
	}
	delete(c.data, key)
	return value, nil
}

func (c *mockCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.data, key)
	}
	return nil
}

func mustAppErr(t *testing.T, err error) *appResponse.AppErrResponse {
	t.Helper()

	var appErr *appResponse.AppErrResponse
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppErrResponse, got %T (%v)", err, err)
	}
	return appErr
}
