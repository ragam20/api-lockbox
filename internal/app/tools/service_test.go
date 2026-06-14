package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"lockbox/internal/app/auth"
	"lockbox/internal/app/vault"
	appResponse "lockbox/internal/http/response"

	"gorm.io/gorm"
)

func TestUpdateArgonConfigSucceeds(t *testing.T) {
	vaultRepo := &mockVaultRepo{}
	authRepo := &mockAuthRepo{}
	service := &toolService{
		authRepo:  authRepo,
		vaultRepo: vaultRepo,
		txManager: &mockTxManager{},
	}

	err := service.UpdateArgonConfig(context.Background(), vault.UpdateArgonConfigInput{
		UserID:       1,
		AuthVerifier: "new-verifier",
		ArgonConfig:  vault.ArgonConfig{ID: 10, Memory: 65536},
		ProtectionKeys: []vault.ItemProtectionKey{
			{ID: 100, IvProtectionKey: "iv", ProtectionKey: "key"},
		},
	})
	if err != nil {
		t.Fatalf("expected update to succeed: %v", err)
	}

	if vaultRepo.updateArgonConfigArg.ID != 10 {
		t.Fatalf("expected argon config ID 10, got %d", vaultRepo.updateArgonConfigArg.ID)
	}
	if authRepo.updateAuthVerifierArg != "new-verifier" {
		t.Fatalf("expected auth verifier to be passed, got %q", authRepo.updateAuthVerifierArg)
	}
	if authRepo.updateAuthVerifierUserID != 1 {
		t.Fatalf("expected user ID 1, got %d", authRepo.updateAuthVerifierUserID)
	}
	if len(vaultRepo.updateProtectionKeysArg) != 1 || vaultRepo.updateProtectionKeysArg[0].ID != 100 {
		t.Fatalf("expected protection keys to be passed through")
	}
}

func TestUpdateArgonConfigFailsOnArgonUpdate(t *testing.T) {
	vaultRepo := &mockVaultRepo{updateArgonConfigErr: errors.New("db error")}
	service := &toolService{
		authRepo:  &mockAuthRepo{},
		vaultRepo: vaultRepo,
		txManager: &mockTxManager{},
	}

	err := service.UpdateArgonConfig(context.Background(), vault.UpdateArgonConfigInput{
		UserID:       1,
		AuthVerifier: "verifier",
		ArgonConfig:  vault.ArgonConfig{ID: 10},
	})
	if err == nil {
		t.Fatal("expected error when argon config update fails")
	}

	appErr := mustToolAppErr(t, err)
	if appErr.Code != vault.ErrInvalidUpdate.Error() {
		t.Fatalf("expected code %q, got %q", vault.ErrInvalidUpdate.Error(), appErr.Code)
	}
}

func TestUpdateArgonConfigFailsOnAuthVerifierUpdate(t *testing.T) {
	authRepo := &mockAuthRepo{updateAuthVerifierErr: errors.New("db error")}
	service := &toolService{
		authRepo:  authRepo,
		vaultRepo: &mockVaultRepo{},
		txManager: &mockTxManager{},
	}

	err := service.UpdateArgonConfig(context.Background(), vault.UpdateArgonConfigInput{
		UserID:       1,
		AuthVerifier: "verifier",
		ArgonConfig:  vault.ArgonConfig{ID: 10},
	})
	if err == nil {
		t.Fatal("expected error when auth verifier update fails")
	}

	appErr := mustToolAppErr(t, err)
	if appErr.Code != vault.ErrInvalidUpdate.Error() {
		t.Fatalf("expected code %q, got %q", vault.ErrInvalidUpdate.Error(), appErr.Code)
	}
}

func TestUpdateArgonConfigFailsOnProtectionKeysUpdate(t *testing.T) {
	vaultRepo := &mockVaultRepo{updateProtectionKeysErr: errors.New("db error")}
	service := &toolService{
		authRepo:  &mockAuthRepo{},
		vaultRepo: vaultRepo,
		txManager: &mockTxManager{},
	}

	err := service.UpdateArgonConfig(context.Background(), vault.UpdateArgonConfigInput{
		UserID:       1,
		AuthVerifier: "verifier",
		ArgonConfig:  vault.ArgonConfig{ID: 10},
		ProtectionKeys: []vault.ItemProtectionKey{
			{ID: 100},
		},
	})
	if err == nil {
		t.Fatal("expected error when protection keys update fails")
	}

	appErr := mustToolAppErr(t, err)
	if appErr.Code != vault.ErrInvalidUpdate.Error() {
		t.Fatalf("expected code %q, got %q", vault.ErrInvalidUpdate.Error(), appErr.Code)
	}
}

type mockVaultRepo struct {
	updateArgonConfigArg    vault.ArgonConfig
	updateArgonConfigErr    error
	updateProtectionKeysArg []vault.ItemProtectionKey
	updateProtectionKeysErr error
}

func (r *mockVaultRepo) WithTx(tx *gorm.DB) vault.VaultRepository { return r }
func (r *mockVaultRepo) UpdateArgonConfig(ctx context.Context, argonConfig vault.ArgonConfig) error {
	r.updateArgonConfigArg = argonConfig
	return r.updateArgonConfigErr
}
func (r *mockVaultRepo) UpdateProtectionKeys(ctx context.Context, keys []vault.ItemProtectionKey) error {
	r.updateProtectionKeysArg = keys
	return r.updateProtectionKeysErr
}
func (r *mockVaultRepo) GetProtectionKeys(ctx context.Context, userID int64) ([]vault.ItemProtectionKey, error) {
	return nil, nil
}
func (r *mockVaultRepo) SetFavorite(ctx context.Context, itemID int64) error  { return nil }
func (r *mockVaultRepo) DeleteFavorite(ctx context.Context, itemID int64) error { return nil }
func (r *mockVaultRepo) GetVault(ctx context.Context, userID int64) (vault.Vault, error) {
	return vault.Vault{}, nil
}
func (r *mockVaultRepo) GetArgonConfig(ctx context.Context, userID int64) (vault.ArgonConfig, error) {
	return vault.ArgonConfig{}, nil
}
func (r *mockVaultRepo) GetVaultByUserID(ctx context.Context, userID int64) (vault.Vault, error) {
	return vault.Vault{}, nil
}
func (r *mockVaultRepo) GetConfigByUserID(ctx context.Context, userID int64) (vault.ArgonConfig, error) {
	return vault.ArgonConfig{}, nil
}
func (r *mockVaultRepo) CreateArgonConfig(ctx context.Context, argonConfig vault.ArgonConfig, userID int64) error {
	return nil
}
func (r *mockVaultRepo) CreateVault(ctx context.Context, vault vault.Vault, userID int64) error {
	return nil
}
func (r *mockVaultRepo) GetVaultItems(ctx context.Context, input vault.GetVaultInput) (vault.GetVaultItemsResult, error) {
	return vault.GetVaultItemsResult{}, nil
}
func (r *mockVaultRepo) GetFavorites(ctx context.Context, input vault.GetVaultInput) (vault.GetVaultItemsResult, error) {
	return vault.GetVaultItemsResult{}, nil
}
func (r *mockVaultRepo) CreateVaultItem(ctx context.Context, item vault.VaultItem) error { return nil }
func (r *mockVaultRepo) GetVaultItemByID(ctx context.Context, input vault.GetVaultItemByIDInput) (vault.VaultItem, error) {
	return vault.VaultItem{}, nil
}
func (r *mockVaultRepo) UpdateVaultItem(ctx context.Context, item vault.VaultItem) error   { return nil }
func (r *mockVaultRepo) ArchiveVaultItem(ctx context.Context, item vault.VaultItem) error  { return nil }
func (r *mockVaultRepo) UnarchiveVaultItem(ctx context.Context, item vault.VaultItem) error { return nil }
func (r *mockVaultRepo) DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error {
	return nil
}

type mockAuthRepo struct {
	updateAuthVerifierArg    string
	updateAuthVerifierUserID int64
	updateAuthVerifierErr    error
}

func (r *mockAuthRepo) WithTx(tx *gorm.DB) auth.AuthRepository { return r }
func (r *mockAuthRepo) UpdateAuthVerifier(ctx context.Context, authVerifier string, userID int64) error {
	r.updateAuthVerifierArg = authVerifier
	r.updateAuthVerifierUserID = userID
	return r.updateAuthVerifierErr
}
func (r *mockAuthRepo) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	return auth.User{}, nil
}
func (r *mockAuthRepo) EmailExists(ctx context.Context, email string) (bool, error) { return false, nil }
func (r *mockAuthRepo) CreateUser(ctx context.Context, user auth.User) error       { return nil }
func (r *mockAuthRepo) RevokeSession(ctx context.Context, userID int64, tokenHash string) error {
	return nil
}
func (r *mockAuthRepo) CreateSession(ctx context.Context, userID int64, userSession auth.UserSession) error {
	return nil
}
func (r *mockAuthRepo) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (auth.UserSession, error) {
	return auth.UserSession{}, nil
}
func (r *mockAuthRepo) UpdateRefreshToken(ctx context.Context, sessionID int64, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	return nil
}

type mockTxManager struct{}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return fn(&gorm.DB{})
}

func mustToolAppErr(t *testing.T, err error) *appResponse.AppErrResponse {
	t.Helper()

	var appErr *appResponse.AppErrResponse
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppErrResponse, got %T (%v)", err, err)
	}
	return appErr
}
