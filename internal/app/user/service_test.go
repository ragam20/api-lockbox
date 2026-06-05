package user

import (
	"context"
	"errors"
	"testing"
	"time"

	appResponse "lockbox/internal/http/response"
	"lockbox/internal/storage"
)

func TestNewUserServiceWiresRepository(t *testing.T) {
	publicImgID := strPtr("public-img-id")
	repo := &mockUserRepo{
		user: User{
			ID:               7,
			Name:             "Test User",
			Email:            "user@example.com",
			PublicImgID:      publicImgID,
			EmailVerified:    true,
			MasterKeySalt:    "master-salt",
			AuthVerifierSalt: "auth-salt",
			CreatedAt:        time.Unix(1, 0).UTC(),
			UpdatedAt:        time.Unix(2, 0).UTC(),
		},
	}

	service := NewUserService(repo, mockStorage{})
	result, err := service.GetUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected get user to succeed: %v", err)
	}

	if repo.calledWithUserID != 7 {
		t.Fatalf("expected repo to be called with userID 7, got %d", repo.calledWithUserID)
	}
	if result.ID != repo.user.ID || result.Email != repo.user.Email || result.Name != repo.user.Name {
		t.Fatalf("unexpected user result: %+v", result)
	}
	if result.PublicImgID == nil || *result.PublicImgID != *publicImgID {
		t.Fatalf("expected public image id %q, got %+v", *publicImgID, result.PublicImgID)
	}
	if result.AuthVerifierSalt != repo.user.AuthVerifierSalt {
		t.Fatalf("expected auth verifier salt %q, got %q", repo.user.AuthVerifierSalt, result.AuthVerifierSalt)
	}
}

func TestSaveAvatarURL(t *testing.T) {
	avatarURL := "https://cdn.example.com/avatar.png"
	publicImgID := "lockbox/users/7/avatar"
	repo := &mockUserRepo{
		user: User{
			ID:    7,
			Name:  "Test User",
			Email: "user@example.com",
		},
	}

	service := NewUserService(repo, mockStorage{})
	if err := service.SaveAvatarURL(context.Background(), SaveAvatarURLInput{UserID: 7, PublicImgID: publicImgID, ImgURL: "  " + avatarURL + "  "}); err != nil {
		t.Fatalf("expected save avatar url to succeed: %v", err)
	}

	if repo.calledWithUserID != 7 {
		t.Fatalf("expected repo to be called with userID 7, got %d", repo.calledWithUserID)
	}
	if repo.updatedUser.ID != repo.user.ID {
		t.Fatalf("expected updated user ID %d, got %d", repo.user.ID, repo.updatedUser.ID)
	}
	if repo.updatedUser.PublicImgID == nil {
		t.Fatal("expected public image id to be set")
	}
	if got := *repo.updatedUser.PublicImgID; got != publicImgID {
		t.Fatalf("expected public image id %q, got %q", publicImgID, got)
	}
	if repo.updatedUser.ImgURL == nil {
		t.Fatal("expected avatar url to be set")
	}
	if got := *repo.updatedUser.ImgURL; got != avatarURL {
		t.Fatalf("expected trimmed avatar url %q, got %q", avatarURL, got)
	}
}

func TestGetUserMapsNotFound(t *testing.T) {
	service := &userService{userRepo: &mockUserRepo{err: ErrUserNotFound}}

	_, err := service.GetUser(context.Background(), 7)
	if err == nil {
		t.Fatal("expected missing user to return an error")
	}

	appErr := mustUserAppErr(t, err)
	if appErr.Code != ErrUserNotFound.Error() {
		t.Fatalf("expected not-found code, got %q", appErr.Code)
	}
	if appErr.Message != "user not found" {
		t.Fatalf("expected not found message, got %q", appErr.Message)
	}
}

func TestSaveAvatarURLMapsNotFound(t *testing.T) {
	service := &userService{userRepo: &mockUserRepo{err: ErrUserNotFound}}

	err := service.SaveAvatarURL(context.Background(), SaveAvatarURLInput{UserID: 7, PublicImgID: "public-img-id", ImgURL: "https://cdn.example.com/avatar.png"})
	if err == nil {
		t.Fatal("expected missing user to return an error")
	}

	appErr := mustUserAppErr(t, err)
	if appErr.Code != ErrUserNotFound.Error() {
		t.Fatalf("expected not-found code, got %q", appErr.Code)
	}
	if appErr.Message != "user not found" {
		t.Fatalf("expected not found message, got %q", appErr.Message)
	}
}

func TestGetUserMapsInternalError(t *testing.T) {
	service := &userService{userRepo: &mockUserRepo{err: errors.New("db down")}}

	_, err := service.GetUser(context.Background(), 7)
	if err == nil {
		t.Fatal("expected internal error to be returned")
	}

	appErr := mustUserAppErr(t, err)
	if appErr.Code != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("expected internal server error code, got %q", appErr.Code)
	}
	if appErr.Message != "db down" {
		t.Fatalf("expected original error message, got %q", appErr.Message)
	}
}

func TestSaveAvatarURLMapsInternalError(t *testing.T) {
	service := &userService{userRepo: &mockUserRepo{err: errors.New("db down")}}

	err := service.SaveAvatarURL(context.Background(), SaveAvatarURLInput{UserID: 7, PublicImgID: "public-img-id", ImgURL: "https://cdn.example.com/avatar.png"})
	if err == nil {
		t.Fatal("expected internal error to be returned")
	}

	appErr := mustUserAppErr(t, err)
	if appErr.Code != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("expected internal server error code, got %q", appErr.Code)
	}
	if appErr.Message != "db down" {
		t.Fatalf("expected original error message, got %q", appErr.Message)
	}
}

type mockUserRepo struct {
	user             User
	err              error
	updateErr        error
	calledWithUserID int64
	updatedUser      User
}

func (r *mockUserRepo) GetUser(ctx context.Context, userID int64) (User, error) {
	r.calledWithUserID = userID
	if r.err != nil {
		return User{}, r.err
	}
	return r.user, nil
}

func (r *mockUserRepo) UpdateProfile(ctx context.Context, user User) error {
	r.updatedUser = user
	if r.updateErr != nil {
		return r.updateErr
	}
	return nil
}

func (r *mockUserRepo) UpdateAvatar(ctx context.Context, user User) error {
	r.updatedUser = user
	if r.updateErr != nil {
		return r.updateErr
	}
	return nil
}

func (r *mockUserRepo) RemoveAvatar(ctx context.Context, user User) error {
	r.updatedUser = user
	if r.updateErr != nil {
		return r.updateErr
	}
	return nil
}

type mockStorage struct{}

func (mockStorage) RemoveAvatar(ctx context.Context, publicImgID string) error {
	return nil
}

func (mockStorage) Sign(params map[string][]string) (string, error) {
	return "", nil
}

func (mockStorage) UploadAvatarRequest() storage.StorageResult {
	return storage.StorageResult{}
}

func strPtr(v string) *string {
	return &v
}

func mustUserAppErr(t *testing.T, err error) *appResponse.AppErrResponse {
	t.Helper()

	var appErr *appResponse.AppErrResponse
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppErrResponse, got %T (%v)", err, err)
	}
	return appErr
}
