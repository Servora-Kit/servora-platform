package biz

import (
	"context"
	"fmt"
	"sync"
	"testing"

	apppb "github.com/Servora-Kit/servora/api/gen/go/servora/application/service/v1"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
)

type fakeApplicationRepo struct {
	mu     sync.RWMutex
	apps   map[string]*apppb.Application
	hashes map[string]string // id -> client_secret_hash
}

func newFakeApplicationRepo() *fakeApplicationRepo {
	return &fakeApplicationRepo{
		apps:   make(map[string]*apppb.Application),
		hashes: make(map[string]string),
	}
}

func (r *fakeApplicationRepo) Create(_ context.Context, app *apppb.Application, clientSecretHash string) (*apppb.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if app.Id == "" {
		app.Id = fmt.Sprintf("app-%d", len(r.apps)+1)
	}
	out := cloneApp(app)
	r.apps[out.Id] = out
	r.hashes[out.Id] = clientSecretHash
	result := cloneApp(out)
	return result, nil
}

func (r *fakeApplicationRepo) GetByID(_ context.Context, id string) (*apppb.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.apps[id]
	if !ok {
		return nil, fmt.Errorf("application not found: %s", id)
	}
	return cloneApp(a), nil
}

func (r *fakeApplicationRepo) GetByClientID(_ context.Context, clientID string) (*apppb.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.apps {
		if a.ClientId == clientID {
			return cloneApp(a), nil
		}
	}
	return nil, fmt.Errorf("application not found by client_id: %s", clientID)
}

func (r *fakeApplicationRepo) List(_ context.Context, page, pageSize int32) ([]*apppb.Application, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*apppb.Application
	for _, a := range r.apps {
		result = append(result, cloneApp(a))
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

func (r *fakeApplicationRepo) Update(_ context.Context, app *apppb.Application) (*apppb.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.apps[app.Id]; !ok {
		return nil, fmt.Errorf("application not found: %s", app.Id)
	}
	out := cloneApp(app)
	r.apps[out.Id] = out
	return cloneApp(out), nil
}

func (r *fakeApplicationRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.apps[id]; !ok {
		return fmt.Errorf("application not found: %s", id)
	}
	delete(r.apps, id)
	delete(r.hashes, id)
	return nil
}

func (r *fakeApplicationRepo) UpdateClientSecretHash(_ context.Context, id string, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.apps[id]; !ok {
		return fmt.Errorf("application not found: %s", id)
	}
	r.hashes[id] = hash
	return nil
}

func cloneApp(a *apppb.Application) *apppb.Application {
	return proto.Clone(a).(*apppb.Application)
}

func newTestApplicationUsecase() (*ApplicationUsecase, *fakeApplicationRepo) {
	repo := newFakeApplicationRepo()
	uc := NewApplicationUsecase(repo, log.DefaultLogger)
	return uc, repo
}

func TestApplicationUsecase_Create(t *testing.T) {
	uc, repo := newTestApplicationUsecase()
	ctx := context.Background()

	app := &apppb.Application{Name: "test-app", Type: "web"}
	created, plainSecret, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if len(created.ClientId) != 32 {
		t.Errorf("ClientId length = %d, want 32 hex chars", len(created.ClientId))
	}
	if len(plainSecret) != 64 {
		t.Errorf("plainSecret length = %d, want 64 hex chars", len(plainSecret))
	}

	repo.mu.RLock()
	hash := repo.hashes[created.Id]
	repo.mu.RUnlock()
	if hash == "" {
		t.Fatal("client secret hash not stored in repo")
	}
	if !helpers.BcryptCheck(plainSecret, hash) {
		t.Error("stored hash does not match plain secret")
	}
}

func TestApplicationUsecase_Get(t *testing.T) {
	uc, _ := newTestApplicationUsecase()
	ctx := context.Background()

	app := &apppb.Application{Name: "get-test", Type: "m2m"}
	created, _, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := uc.Get(ctx, created.Id)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Id != created.Id {
		t.Errorf("Get().Id = %q, want %q", got.Id, created.Id)
	}
	if got.Name != "get-test" {
		t.Errorf("Get().Name = %q, want %q", got.Name, "get-test")
	}
}

func TestApplicationUsecase_RegenerateClientSecret(t *testing.T) {
	uc, repo := newTestApplicationUsecase()
	ctx := context.Background()

	app := &apppb.Application{Name: "regen-test", Type: "web"}
	created, oldSecret, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	repo.mu.RLock()
	oldHash := repo.hashes[created.Id]
	repo.mu.RUnlock()

	newSecret, err := uc.RegenerateClientSecret(ctx, created.Id)
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
	newHash := repo.hashes[created.Id]
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

	app := &apppb.Application{Name: "delete-test", Type: "web"}
	created, _, err := uc.Create(ctx, app)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := uc.Delete(ctx, created.Id); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err = uc.Get(ctx, created.Id)
	if err == nil {
		t.Error("Get() after Delete() should return error")
	}
}
