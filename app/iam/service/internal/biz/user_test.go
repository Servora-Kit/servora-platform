package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	userpb "github.com/Servora-Kit/servora/api/gen/go/servora/user/service/v1"
)

// --- fakes ---

type fakeUserRepo struct {
	purgeCascadeErr   error
	purgeCascadeCalls []string
}

func (r *fakeUserRepo) PurgeCascade(_ context.Context, id string) error {
	r.purgeCascadeCalls = append(r.purgeCascadeCalls, id)
	return r.purgeCascadeErr
}

func (r *fakeUserRepo) SaveUser(context.Context, *userpb.User, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) GetUserById(context.Context, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) DeleteUser(context.Context, string) error  { return nil }
func (r *fakeUserRepo) PurgeUser(context.Context, string) error   { return nil }
func (r *fakeUserRepo) RestoreUser(context.Context, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) GetUserByIdIncludingDeleted(context.Context, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) UpdateUser(context.Context, *userpb.User) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) ListUsers(context.Context, int32, int32) ([]*userpb.User, int64, error) {
	return nil, 0, nil
}

type fakeAuthnRepo struct {
	deleteTokensErr   error
	deleteTokensCalls []string
}

func (r *fakeAuthnRepo) DeleteUserRefreshTokens(_ context.Context, userID string) error {
	r.deleteTokensCalls = append(r.deleteTokensCalls, userID)
	return r.deleteTokensErr
}

func (r *fakeAuthnRepo) SaveUser(context.Context, *userpb.User, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeAuthnRepo) GetUserByEmail(context.Context, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeAuthnRepo) GetUserByUserName(context.Context, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeAuthnRepo) GetUserByID(context.Context, string) (*userpb.User, error) {
	return nil, nil
}
func (r *fakeAuthnRepo) GetPasswordHash(context.Context, string) (string, error) {
	return "", nil
}
func (r *fakeAuthnRepo) UpdatePassword(context.Context, string, string) error   { return nil }
func (r *fakeAuthnRepo) UpdateEmailVerified(context.Context, string, bool) error { return nil }
func (r *fakeAuthnRepo) SaveRefreshToken(context.Context, string, string, time.Duration) error {
	return nil
}
func (r *fakeAuthnRepo) GetRefreshToken(context.Context, string) (string, error) { return "", nil }
func (r *fakeAuthnRepo) DeleteRefreshToken(context.Context, string) error        { return nil }

type fakeAuthZRepo struct {
	deleteTuplesCalls [][]Tuple
	deleteTuplesErr   error
	listObjectsMap    map[string][]string
	listObjectsErr    error
}

func (r *fakeAuthZRepo) DeleteTuples(_ context.Context, tuples ...Tuple) error {
	r.deleteTuplesCalls = append(r.deleteTuplesCalls, tuples)
	return r.deleteTuplesErr
}

func (r *fakeAuthZRepo) WriteTuples(context.Context, ...Tuple) error { return nil }
func (r *fakeAuthZRepo) Check(context.Context, string, string, string, string) (bool, error) {
	return false, nil
}
func (r *fakeAuthZRepo) ListObjects(_ context.Context, _ string, relation, objectType string) ([]string, error) {
	if r.listObjectsErr != nil {
		return nil, r.listObjectsErr
	}
	if r.listObjectsMap != nil {
		return r.listObjectsMap[relation+":"+objectType], nil
	}
	return nil, nil
}
func (r *fakeAuthZRepo) CachedListObjects(context.Context, time.Duration, string, string, string) ([]string, error) {
	return nil, nil
}
func (r *fakeAuthZRepo) InvalidateCheck(context.Context, string, string, string, string) {}
func (r *fakeAuthZRepo) InvalidateListObjects(context.Context, string, string, string)   {}

// --- helpers ---

func newTestPurgeUserUC(userRepo *fakeUserRepo, authnRepo *fakeAuthnRepo) *UserUsecase {
	return NewUserUsecase(userRepo, log.DefaultLogger, nil, authnRepo, nil, &fakeAuthZRepo{})
}

// --- tests ---

func TestPurgeUser_HappyPath(t *testing.T) {
	ur := &fakeUserRepo{}
	ar := &fakeAuthnRepo{}

	uc := newTestPurgeUserUC(ur, ar)
	ok, err := uc.PurgeUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("PurgeUser() unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("PurgeUser() returned false, want true")
	}

	if len(ur.purgeCascadeCalls) != 1 || ur.purgeCascadeCalls[0] != "user-1" {
		t.Errorf("PurgeCascade calls = %v, want [user-1]", ur.purgeCascadeCalls)
	}
	if len(ar.deleteTokensCalls) != 1 || ar.deleteTokensCalls[0] != "user-1" {
		t.Errorf("DeleteUserRefreshTokens calls = %v, want [user-1]", ar.deleteTokensCalls)
	}
}

func TestPurgeUser_CascadeFails_StopsEarly(t *testing.T) {
	ur := &fakeUserRepo{purgeCascadeErr: errors.New("db error")}
	ar := &fakeAuthnRepo{}

	uc := newTestPurgeUserUC(ur, ar)
	ok, err := uc.PurgeUser(context.Background(), "user-1")
	if err == nil {
		t.Fatal("PurgeUser() expected error when PurgeCascade fails")
	}
	if ok {
		t.Fatal("PurgeUser() returned true, want false")
	}

	if len(ar.deleteTokensCalls) != 0 {
		t.Error("Redis should not be called when PurgeCascade fails")
	}
}

func TestPurgeUser_RedisFails_StillSucceeds(t *testing.T) {
	ur := &fakeUserRepo{}
	ar := &fakeAuthnRepo{deleteTokensErr: errors.New("redis error")}

	uc := newTestPurgeUserUC(ur, ar)
	ok, err := uc.PurgeUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("PurgeUser() should succeed even when Redis fails: %v", err)
	}
	if !ok {
		t.Fatal("PurgeUser() returned false, want true")
	}
}

func TestPurgeUser_ExecutionOrder_DBBeforeRedis(t *testing.T) {
	var order []string

	ur := &orderTrackingUserRepo{order: &order}
	ar := &orderTrackingAuthnRepo{order: &order}

	uc := NewUserUsecase(ur, log.DefaultLogger, nil, ar, nil, &fakeAuthZRepo{})
	_, _ = uc.PurgeUser(context.Background(), "user-1")

	if len(order) < 2 {
		t.Fatalf("expected at least 2 operations, got %v", order)
	}
	if order[0] != "db" {
		t.Errorf("first operation = %q, want 'db'", order[0])
	}
	if order[1] != "redis" {
		t.Errorf("second operation = %q, want 'redis'", order[1])
	}
}

// Order-tracking fakes

type orderTrackingUserRepo struct {
	fakeUserRepo
	order *[]string
}

func (r *orderTrackingUserRepo) PurgeCascade(_ context.Context, _ string) error {
	*r.order = append(*r.order, "db")
	return nil
}

type orderTrackingAuthnRepo struct {
	fakeAuthnRepo
	order *[]string
}

func (r *orderTrackingAuthnRepo) DeleteUserRefreshTokens(_ context.Context, _ string) error {
	*r.order = append(*r.order, "redis")
	return nil
}
