package user

import (
	"context"
	"errors"
	"testing"
	"time"

	appResponse "lockbox/internal/http/response"
)

func TestNewUserServiceWiresRepository(t *testing.T) {
	repo := &mockUserRepo{
		user: User{
			ID:               7,
			Name:             "Test User",
			Email:            "user@example.com",
			EmailVerified:    true,
			MasterKeySalt:    "master-salt",
			AuthVerifierSalt: "auth-salt",
			CreatedAt:        time.Unix(1, 0).UTC(),
			UpdatedAt:        time.Unix(2, 0).UTC(),
		},
	}

	service := NewUserService(repo)
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
	if result.AuthVerifierSalt != repo.user.AuthVerifierSalt {
		t.Fatalf("expected auth verifier salt %q, got %q", repo.user.AuthVerifierSalt, result.AuthVerifierSalt)
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

func mustUserAppErr(t *testing.T, err error) *appResponse.AppErrResponse {
	t.Helper()

	var appErr *appResponse.AppErrResponse
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppErrResponse, got %T (%v)", err, err)
	}
	return appErr
}
