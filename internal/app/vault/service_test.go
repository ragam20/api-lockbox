package vault

import (
	"context"
	"errors"
	"testing"
	"time"

	appResponse "lockbox/internal/http/response"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestGetArgonConfigReturnsConfigAndMapsMissing(t *testing.T) {
	repo := &mockVaultRepo{
		getArgonConfigResult: ArgonConfig{
			ID:          11,
			Iterations:  4,
			Memory:      128 * 1024,
			Parallelism: 4,
		},
	}
	service := &vaultService{vaultRepo: repo}

	result, err := service.GetArgonConfig(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected argon config lookup to succeed: %v", err)
	}
	if result.ID != 11 || result.Iterations != 4 || result.Memory != 128*1024 || result.Parallelism != 4 {
		t.Fatalf("unexpected argon config result: %+v", result)
	}
	if repo.getArgonConfigUserID != 7 {
		t.Fatalf("expected userID 7 to be queried, got %d", repo.getArgonConfigUserID)
	}

	repo.getArgonConfigErr = ErrNotFound
	_, err = service.GetArgonConfig(context.Background(), 7)
	if err == nil {
		t.Fatal("expected missing argon config to return an error")
	}

	appErr := mustVaultAppErr(t, err)
	if appErr.Code != ErrNotFound.Error() {
		t.Fatalf("expected not-found code, got %q", appErr.Code)
	}
	if appErr.Message != "argon config not found" {
		t.Fatalf("expected argon config not found message, got %q", appErr.Message)
	}
}

func TestCreateVaultUsesDefaults(t *testing.T) {
	repo := &mockVaultRepo{}
	service := &vaultService{vaultRepo: repo, idGen: stubIDGen{id: 101}}

	if err := service.CreateVault(context.Background(), 7, "vault-key-salt"); err != nil {
		t.Fatalf("expected create vault to succeed: %v", err)
	}

	if repo.createVaultUserID != 7 {
		t.Fatalf("expected userID 7, got %d", repo.createVaultUserID)
	}
	if repo.createVaultArg.ID != 101 {
		t.Fatalf("expected generated vault ID 101, got %d", repo.createVaultArg.ID)
	}
	if repo.createVaultArg.Timeout != defaultVaultTimeout {
		t.Fatalf("expected default timeout %d, got %d", defaultVaultTimeout, repo.createVaultArg.Timeout)
	}
	if repo.createVaultArg.TimeoutType != defaultVaultTimeoutType {
		t.Fatalf("expected default timeout type %q, got %q", defaultVaultTimeoutType, repo.createVaultArg.TimeoutType)
	}
	if repo.createVaultArg.VaultKeySalt != "vault-key-salt" {
		t.Fatalf("expected vault key salt to be passed through, got %q", repo.createVaultArg.VaultKeySalt)
	}
}

func TestCreateArgonConfigUsesDefaults(t *testing.T) {
	repo := &mockVaultRepo{}
	service := &vaultService{vaultRepo: repo, idGen: stubIDGen{id: 202}}

	if err := service.CreateArgonConfig(context.Background(), 7); err != nil {
		t.Fatalf("expected create argon config to succeed: %v", err)
	}

	if repo.createArgonConfigUserID != 7 {
		t.Fatalf("expected userID 7, got %d", repo.createArgonConfigUserID)
	}
	if repo.createArgonConfigArg.ID != 202 {
		t.Fatalf("expected generated argon config ID 202, got %d", repo.createArgonConfigArg.ID)
	}
	if repo.createArgonConfigArg.Memory != defaultArgonMemory {
		t.Fatalf("expected default memory %d, got %d", defaultArgonMemory, repo.createArgonConfigArg.Memory)
	}
	if repo.createArgonConfigArg.Iterations != defaultArgonIterations {
		t.Fatalf("expected default iterations %d, got %d", defaultArgonIterations, repo.createArgonConfigArg.Iterations)
	}
	if repo.createArgonConfigArg.Parallelism != defaultArgonParallelism {
		t.Fatalf("expected default parallelism %d, got %d", defaultArgonParallelism, repo.createArgonConfigArg.Parallelism)
	}
}

func TestGetVaultItemRejectsInvalidType(t *testing.T) {
	service := &vaultService{vaultRepo: &mockVaultRepo{}}

	_, err := service.GetVaultItem(context.Background(), GetVaultInput{
		UserID:    7,
		VaultType: "invalid",
	})
	if err == nil {
		t.Fatal("expected invalid vault type to be rejected")
	}

	appErr := mustVaultAppErr(t, err)
	if appErr.Code != ErrVaultType.Error() {
		t.Fatalf("expected invalid type code, got %q", appErr.Code)
	}
}

func TestGetVaultItemTrimsSearchAndReturnsResult(t *testing.T) {
	repo := &mockVaultRepo{
		getVaultItemResult: GetVaultItemsResult{
			Items: []VaultItem{{ID: 1, Title: "First"}},
			TotalItems: 1,
		},
	}
	service := &vaultService{vaultRepo: repo}

	result, err := service.GetVaultItem(context.Background(), GetVaultInput{
		UserID:    7,
		VaultType: string(VaultItemTypePassword),
		Query: VaultItemQuery{
			Page:   2,
			Limit:  5,
			SortBy: "title,asc",
			Search: "  hello  ",
		},
	})
	if err != nil {
		t.Fatalf("expected vault items lookup to succeed: %v", err)
	}
	if repo.getVaultItemInput.Query.Search != "hello" {
		t.Fatalf("expected search to be trimmed, got %q", repo.getVaultItemInput.Query.Search)
	}
	if repo.getVaultItemInput.Query.Page != 2 || repo.getVaultItemInput.Query.Limit != 5 {
		t.Fatalf("unexpected pagination input: %+v", repo.getVaultItemInput.Query)
	}
	if result.TotalItems != 1 || len(result.Items) != 1 || result.Items[0].Title != "First" {
		t.Fatalf("unexpected vault items result: %+v", result)
	}
}

func TestCreateVaultItemReturnsCreatedItem(t *testing.T) {
	repo := &mockVaultRepo{
		getVaultByUserIDResult: Vault{ID: 99},
	}
	repo.getVaultItemByIDFunc = func(input GetVaultItemByIDInput) (VaultItem, error) {
		if input.ItemID != repo.createVaultItemArg.ID {
			return VaultItem{}, gorm.ErrRecordNotFound
		}
		return repo.createVaultItemArg, nil
	}
	service := &vaultService{vaultRepo: repo, idGen: stubIDGen{id: 303}}

	result, err := service.CreateVaultItem(context.Background(), CreateVaultItemInput{
		UserID:          7,
		Title:           "  Login  ",
		ItemType:        VaultItemTypePassword,
		IvCiphertext:    " YQ== ",
		Ciphertext:      " Yg== ",
		IvProtectionKey: " Yw== ",
		ProtectionKey:   " ZA== ",
	})
	if err != nil {
		t.Fatalf("expected create vault item to succeed: %v", err)
	}
	if repo.getVaultByUserIDUserID != 7 {
		t.Fatalf("expected vault lookup for user 7, got %d", repo.getVaultByUserIDUserID)
	}
	if repo.createVaultItemArg.ID != 303 {
		t.Fatalf("expected generated item ID 303, got %d", repo.createVaultItemArg.ID)
	}
	if repo.createVaultItemArg.VaultID != 99 {
		t.Fatalf("expected vault ID 99, got %d", repo.createVaultItemArg.VaultID)
	}
	if repo.createVaultItemArg.Title != "Login" {
		t.Fatalf("expected title to be trimmed, got %q", repo.createVaultItemArg.Title)
	}
	if result.ID != 303 || result.VaultID != 99 {
		t.Fatalf("unexpected created item result: %+v", result)
	}
}

func TestUpdateVaultItemReturnsUpdatedItem(t *testing.T) {
	existing := VaultItem{ID: 55, VaultID: 99, Title: "Old", ItemType: VaultItemTypeNote}
	repo := &mockVaultRepo{}
	calls := 0
	repo.getVaultItemByIDFunc = func(input GetVaultItemByIDInput) (VaultItem, error) {
		calls++
		if calls == 1 {
			return existing, nil
		}
		return repo.updatedVaultItem, nil
	}
	service := &vaultService{vaultRepo: repo}

	result, err := service.UpdateVaultItem(context.Background(), UpdateVaultItemInput{
		UserID:          7,
		ItemID:          55,
		Title:           "  Updated  ",
		ItemType:        VaultItemTypePassword,
		IvCiphertext:    " YQ== ",
		Ciphertext:      " Yg== ",
		IvProtectionKey: " Yw== ",
		ProtectionKey:   " ZA== ",
	})
	if err != nil {
		t.Fatalf("expected update vault item to succeed: %v", err)
	}
	if repo.updateVaultItemArg.ID != 55 || repo.updateVaultItemArg.VaultID != 99 {
		t.Fatalf("unexpected update payload: %+v", repo.updateVaultItemArg)
	}
	if repo.updateVaultItemArg.Title != "Updated" {
		t.Fatalf("expected title to be trimmed, got %q", repo.updateVaultItemArg.Title)
	}
	if result.Title != "Updated" || result.ItemType != VaultItemTypePassword {
		t.Fatalf("unexpected updated item result: %+v", result)
	}
}

func TestArchiveUnarchiveAndDeleteVaultItem(t *testing.T) {
	t.Run("archive", func(t *testing.T) {
		repo := &mockVaultRepo{}
		existing := VaultItem{ID: 70, VaultID: 99, Title: "Item"}
		calls := 0
		repo.getVaultItemByIDFunc = func(input GetVaultItemByIDInput) (VaultItem, error) {
			calls++
			if calls == 1 {
				return existing, nil
			}
			return repo.archivedVaultItem, nil
		}
		service := &vaultService{vaultRepo: repo}

		result, err := service.ArchiveVaultItem(context.Background(), 7, 70)
		if err != nil {
			t.Fatalf("expected archive to succeed: %v", err)
		}
		if repo.archiveVaultItemArg.ID != 70 || repo.archiveVaultItemArg.VaultID != 99 {
			t.Fatalf("unexpected archive payload: %+v", repo.archiveVaultItemArg)
		}
		if repo.archiveVaultItemArg.ArchivedAt == nil {
			t.Fatal("expected archived item to have ArchivedAt set")
		}
		if result.ArchivedAt == nil {
			t.Fatal("expected archived item to be returned")
		}
	})

	t.Run("unarchive", func(t *testing.T) {
		repo := &mockVaultRepo{}
		existing := VaultItem{ID: 71, VaultID: 99, Title: "Item", ArchivedAt: timePtr(time.Now().UTC())}
		calls := 0
		repo.getVaultItemByIDFunc = func(input GetVaultItemByIDInput) (VaultItem, error) {
			calls++
			if calls == 1 {
				return existing, nil
			}
			return repo.unarchivedVaultItem, nil
		}
		service := &vaultService{vaultRepo: repo}

		result, err := service.UnarchiveVaultItem(context.Background(), 7, 71)
		if err != nil {
			t.Fatalf("expected unarchive to succeed: %v", err)
		}
		if repo.unarchiveVaultItemArg.ID != 71 || repo.unarchiveVaultItemArg.VaultID != 99 {
			t.Fatalf("unexpected unarchive payload: %+v", repo.unarchiveVaultItemArg)
		}
		if repo.unarchiveVaultItemArg.ArchivedAt != nil {
			t.Fatal("expected unarchived item to clear ArchivedAt")
		}
		if result.ArchivedAt != nil {
			t.Fatal("expected returned item to be unarchived")
		}
	})

	t.Run("delete", func(t *testing.T) {
		repo := &mockVaultRepo{}
		service := &vaultService{vaultRepo: repo}

		if err := service.DeleteVaultItem(context.Background(), 7, 80); err != nil {
			t.Fatalf("expected delete to succeed: %v", err)
		}
		if repo.deleteVaultItemUserID != 7 || repo.deleteVaultItemItemID != 80 {
			t.Fatalf("unexpected delete payload: userID=%d itemID=%d", repo.deleteVaultItemUserID, repo.deleteVaultItemItemID)
		}
	})
}

type mockVaultRepo struct {
	getArgonConfigResult ArgonConfig
	getArgonConfigErr    error
	getArgonConfigUserID int64

	createVaultArg    Vault
	createVaultUserID int64
	createVaultErr    error

	createArgonConfigArg    ArgonConfig
	createArgonConfigUserID int64
	createArgonConfigErr    error

	getVaultByUserIDResult Vault
	getVaultByUserIDErr    error
	getVaultByUserIDUserID int64

	getVaultItemInput  GetVaultInput
	getVaultItemResult GetVaultItemsResult
	getVaultItemErr    error

	createVaultItemArg VaultItem
	createVaultItemErr error

	getVaultItemByIDCalls  int
	getVaultItemByIDFunc   func(input GetVaultItemByIDInput) (VaultItem, error)
	getVaultItemByIDResult VaultItem
	getVaultItemByIDErr    error

	updateVaultItemArg VaultItem
	updateVaultItemErr error

	archiveVaultItemArg VaultItem
	archiveVaultItemErr error

	unarchiveVaultItemArg VaultItem
	unarchiveVaultItemErr error

	deleteVaultItemUserID int64
	deleteVaultItemItemID int64
	deleteVaultItemErr    error

	archivedVaultItem   VaultItem
	unarchivedVaultItem  VaultItem
	updatedVaultItem     VaultItem
}

func (r *mockVaultRepo) GetArgonConfig(ctx context.Context, userID int64) (ArgonConfig, error) {
	r.getArgonConfigUserID = userID
	return r.getArgonConfigResult, r.getArgonConfigErr
}

func (r *mockVaultRepo) CreateVault(ctx context.Context, vault Vault, userID int64) error {
	r.createVaultArg = vault
	r.createVaultUserID = userID
	return r.createVaultErr
}

func (r *mockVaultRepo) CreateArgonConfig(ctx context.Context, argonConfig ArgonConfig, userID int64) error {
	r.createArgonConfigArg = argonConfig
	r.createArgonConfigUserID = userID
	return r.createArgonConfigErr
}

func (r *mockVaultRepo) GetConfigByUserID(ctx context.Context, userID int64) (ArgonConfig, error) {
	r.getArgonConfigUserID = userID
	return r.getArgonConfigResult, r.getArgonConfigErr
}

func (r *mockVaultRepo) GetVaultByUserID(ctx context.Context, userID int64) (Vault, error) {
	r.getVaultByUserIDUserID = userID
	return r.getVaultByUserIDResult, r.getVaultByUserIDErr
}

func (r *mockVaultRepo) GetVaultItem(ctx context.Context, input GetVaultInput) (GetVaultItemsResult, error) {
	r.getVaultItemInput = input
	return r.getVaultItemResult, r.getVaultItemErr
}

func (r *mockVaultRepo) CreateVaultItem(ctx context.Context, item VaultItem) error {
	r.createVaultItemArg = item
	return r.createVaultItemErr
}

func (r *mockVaultRepo) GetVaultItemByID(ctx context.Context, input GetVaultItemByIDInput) (VaultItem, error) {
	r.getVaultItemByIDCalls++
	if r.getVaultItemByIDFunc != nil {
		return r.getVaultItemByIDFunc(input)
	}
	return r.getVaultItemByIDResult, r.getVaultItemByIDErr
}

func (r *mockVaultRepo) UpdateVaultItem(ctx context.Context, item VaultItem) error {
	r.updateVaultItemArg = item
	r.updatedVaultItem = item
	return r.updateVaultItemErr
}

func (r *mockVaultRepo) ArchiveVaultItem(ctx context.Context, item VaultItem) error {
	r.archiveVaultItemArg = item
	r.archivedVaultItem = item
	return r.archiveVaultItemErr
}

func (r *mockVaultRepo) UnarchiveVaultItem(ctx context.Context, item VaultItem) error {
	r.unarchiveVaultItemArg = item
	r.unarchivedVaultItem = item
	return r.unarchiveVaultItemErr
}

func (r *mockVaultRepo) DeleteVaultItem(ctx context.Context, userID int64, itemID int64) error {
	r.deleteVaultItemUserID = userID
	r.deleteVaultItemItemID = itemID
	return r.deleteVaultItemErr
}

type stubIDGen struct {
	id int64
}

func (g stubIDGen) NewID() int64 { return g.id }

func (g stubIDGen) NewUUIDV7() (uuid.UUID, error) { return uuid.Nil, nil }

func mustVaultAppErr(t *testing.T, err error) *appResponse.AppErrResponse {
	t.Helper()

	var appErr *appResponse.AppErrResponse
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppErrResponse, got %T (%v)", err, err)
	}
	return appErr
}

func timePtr(t time.Time) *time.Time { return &t }

