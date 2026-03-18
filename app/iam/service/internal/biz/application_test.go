package biz

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/go-kratos/kratos/v2/log"
)

type fakeApplicationRepo struct {
	mu   sync.RWMutex
	apps map[string]*entity.Application
}

func newFakeApplicationRepo() *fakeApplicationRepo {
	return &fakeApplicationRepo{apps: make(map[string]*entity.Application)}
}

func (r *fakeApplicationRepo) Create(_ context.Context, app *entity.Application) (*entity.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if app.ID == "" {
		app.ID = fmt.Sprintf("app-%d", len(r.apps)+1)
	}
	stored := *app
	r.apps[stored.ID] = &stored
	out := stored
	return &out, nil
}

func (r *fakeApplicationRepo) GetByID(_ context.Context, _, id string) (*entity.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.apps[id]
	if !ok {
		return nil, fmt.Errorf("application not found: %s", id)
	}
	out := *a
	return &out, nil
}

func (r *fakeApplicationRepo) GetByClientID(_ context.Context, clientID string) (*entity.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.apps {
		if a.ClientID == clientID {
			out := *a
			return &out, nil
		}
	}
	return nil, fmt.Errorf("application not found by client_id: %s", clientID)
}

func (r *fakeApplicationRepo) ListByTenantID(_ context.Context, tenantID string, page, pageSize int32) ([]*entity.Application, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entity.Application
	for _, a := range r.apps {
		if a.TenantID == tenantID {
			out := *a
			result = append(result, &out)
		}
	}
	total := int64(len(result))
	start := int((page - 1) * pageSize)
	if start >= len(result) {
		return nil, total, nil
	}
	end := start + int(pageSize)
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (r *fakeApplicationRepo) Update(_ context.Context, _ string, app *entity.Application) (*entity.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.apps[app.ID]; !ok {
		return nil, fmt.Errorf("application not found: %s", app.ID)
	}
	stored := *app
	r.apps[stored.ID] = &stored
	out := stored
	return &out, nil
}

func (r *fakeApplicationRepo) Delete(_ context.Context, _, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.apps[id]; !ok {
		return fmt.Errorf("application not found: %s", id)
	}
	delete(r.apps, id)
	return nil
}

func (r *fakeApplicationRepo) UpdateClientSecretHash(_ context.Context, _, id string, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.apps[id]
	if !ok {
		return fmt.Errorf("application not found: %s", id)
	}
	a.ClientSecretHash = hash
	return nil
}

func newTestApplicationUsecase() (*ApplicationUsecase, *fakeApplicationRepo) {
	repo := newFakeApplicationRepo()
	uc := NewApplicationUsecase(repo, log.DefaultLogger)
	return uc, repo
}

func TestApplicationUsecase_Create(t *testing.T) {
	uc, repo := newTestApplicationUsecase()
	ctx := context.Background()

	app := &entity.Application{Name: "test-app", TenantID: "tenant-1"}
	created, plainSecret, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if len(created.ClientID) != 32 {
		t.Errorf("ClientID length = %d, want 32 hex chars", len(created.ClientID))
	}
	if len(plainSecret) != 64 {
		t.Errorf("plainSecret length = %d, want 64 hex chars", len(plainSecret))
	}

	repo.mu.RLock()
	stored := repo.apps[created.ID]
	repo.mu.RUnlock()
	if stored == nil {
		t.Fatal("application not stored in repo")
	}
	if !helpers.BcryptCheck(plainSecret, stored.ClientSecretHash) {
		t.Error("stored hash does not match plain secret")
	}
}

func TestApplicationUsecase_Get(t *testing.T) {
	uc, _ := newTestApplicationUsecase()
	ctx := context.Background()

	app := &entity.Application{Name: "get-test", TenantID: "tenant-1"}
	created, _, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := uc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("Get().ID = %q, want %q", got.ID, created.ID)
	}
	if got.Name != "get-test" {
		t.Errorf("Get().Name = %q, want %q", got.Name, "get-test")
	}
}

func TestApplicationUsecase_RegenerateClientSecret(t *testing.T) {
	uc, repo := newTestApplicationUsecase()
	ctx := context.Background()

	app := &entity.Application{Name: "regen-test", TenantID: "tenant-1"}
	created, oldSecret, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	repo.mu.RLock()
	oldHash := repo.apps[created.ID].ClientSecretHash
	repo.mu.RUnlock()

	newSecret, err := uc.RegenerateClientSecret(ctx, created.ID)
	if err != nil {
		t.Fatalf("RegenerateClientSecret() error: %v", err)
	}
	if newSecret == oldSecret {
		t.Error("new secret should differ from old secret")
	}
	if len(newSecret) != 64 {
		t.Errorf("newSecret length = %d, want 64 hex chars", len(newSecret))
	}

	repo.mu.RLock()
	newHash := repo.apps[created.ID].ClientSecretHash
	repo.mu.RUnlock()

	if newHash == oldHash {
		t.Error("stored hash should be updated after regeneration")
	}
	if !helpers.BcryptCheck(newSecret, newHash) {
		t.Error("new hash does not match new plain secret")
	}
}

func TestApplicationUsecase_Delete(t *testing.T) {
	uc, _ := newTestApplicationUsecase()
	ctx := context.Background()

	app := &entity.Application{Name: "delete-test", TenantID: "tenant-1"}
	created, _, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := uc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err = uc.Get(ctx, created.ID)
	if err == nil {
		t.Error("Get() after Delete() should return error")
	}
}
