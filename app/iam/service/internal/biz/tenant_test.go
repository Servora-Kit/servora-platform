package biz

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
)

// richTenantRepo is a fakeTenantRepo with member state for testing.
type richTenantRepo struct {
	fakeTenantRepo
	members map[string]*entity.TenantMember // key: tenantID+":"+userID
}

func newRichTenantRepo() *richTenantRepo {
	return &richTenantRepo{
		members: make(map[string]*entity.TenantMember),
	}
}

func (r *richTenantRepo) key(tenantID, userID string) string {
	return tenantID + ":" + userID
}

func (r *richTenantRepo) AddMember(_ context.Context, m *entity.TenantMember) (*entity.TenantMember, error) {
	r.members[r.key(m.TenantID, m.UserID)] = m
	return m, nil
}

func (r *richTenantRepo) GetMember(_ context.Context, tenantID, userID string) (*entity.TenantMember, error) {
	m, ok := r.members[r.key(tenantID, userID)]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *m
	return &cp, nil
}

func (r *richTenantRepo) UpdateMemberRole(_ context.Context, tenantID, userID, role string) (*entity.TenantMember, error) {
	k := r.key(tenantID, userID)
	m, ok := r.members[k]
	if !ok {
		return nil, ErrNotFound
	}
	m.Role = role
	return m, nil
}

func (r *richTenantRepo) RemoveMember(_ context.Context, tenantID, userID string) error {
	delete(r.members, r.key(tenantID, userID))
	return nil
}

func (r *richTenantRepo) GetOwnerMember(_ context.Context, tenantID string) (*entity.TenantMember, error) {
	for k, m := range r.members {
		if m.TenantID == tenantID && m.Role == "owner" {
			_ = k
			cp := *m
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

var ErrNotFound = &notFoundError{}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "not found" }

func newTenantUCWithRepo(repo TenantRepo) *TenantUsecase {
	orgUC := &OrganizationUsecase{}
	return NewTenantUsecase(repo, orgUC, nil, log.DefaultLogger)
}

func TestTenantTransferOwnership_HappyPath(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	repo.members["t1:owner-user"] = &entity.TenantMember{TenantID: "t1", UserID: "owner-user", Role: "owner"}
	repo.members["t1:admin-user"] = &entity.TenantMember{TenantID: "t1", UserID: "admin-user", Role: "admin"}

	uc := newTenantUCWithRepo(repo)
	err := uc.TransferOwnership(ctx, "t1", "owner-user", "admin-user")
	if err != nil {
		t.Fatalf("TransferOwnership() unexpected error: %v", err)
	}

	ownerMember := repo.members["t1:owner-user"]
	if ownerMember.Role != "admin" {
		t.Errorf("old owner role = %q, want admin", ownerMember.Role)
	}

	newOwnerMember := repo.members["t1:admin-user"]
	if newOwnerMember.Role != "owner" {
		t.Errorf("new owner role = %q, want owner", newOwnerMember.Role)
	}
}

// TestTenantTransferOwnership_ForceTransfer verifies that a platform admin (who is NOT
// a DB member of the tenant) can force-transfer ownership via the biz layer.
// The authz middleware is responsible for verifying can_transfer_ownership; the biz
// layer only enforces business rules (target must be admin, owner must exist).
func TestTenantTransferOwnership_ForceTransfer(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	// Current owner and an admin exist; platform-admin is NOT a member.
	repo.members["t1:owner-user"] = &entity.TenantMember{TenantID: "t1", UserID: "owner-user", Role: "owner"}
	repo.members["t1:admin-user"] = &entity.TenantMember{TenantID: "t1", UserID: "admin-user", Role: "admin"}

	uc := newTenantUCWithRepo(repo)
	// platform-admin is the caller but not in the member map.
	err := uc.TransferOwnership(ctx, "t1", "platform-admin", "admin-user")
	if err != nil {
		t.Fatalf("ForceTransfer() unexpected error: %v", err)
	}

	// original owner should now be admin
	if repo.members["t1:owner-user"].Role != "admin" {
		t.Errorf("old owner role = %q, want admin", repo.members["t1:owner-user"].Role)
	}
	// target admin should now be owner
	if repo.members["t1:admin-user"].Role != "owner" {
		t.Errorf("new owner role = %q, want owner", repo.members["t1:admin-user"].Role)
	}
}

func TestTenantTransferOwnership_SameUser(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	repo.members["t1:owner-user"] = &entity.TenantMember{TenantID: "t1", UserID: "owner-user", Role: "owner"}

	uc := newTenantUCWithRepo(repo)
	err := uc.TransferOwnership(ctx, "t1", "owner-user", "owner-user")
	if err == nil {
		t.Fatal("TransferOwnership() should fail when caller and new owner are the same")
	}
}

func TestTenantTransferOwnership_TargetNotAdmin(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	repo.members["t1:owner-user"] = &entity.TenantMember{TenantID: "t1", UserID: "owner-user", Role: "owner"}
	repo.members["t1:member-user"] = &entity.TenantMember{TenantID: "t1", UserID: "member-user", Role: "member"}

	uc := newTenantUCWithRepo(repo)
	err := uc.TransferOwnership(ctx, "t1", "owner-user", "member-user")
	if err == nil {
		t.Fatal("TransferOwnership() should fail when target is not admin")
	}
}

func TestTenantRemoveMember_OwnerProtected(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	repo.members["t1:owner-user"] = &entity.TenantMember{TenantID: "t1", UserID: "owner-user", Role: "owner"}

	uc := newTenantUCWithRepo(repo)
	err := uc.RemoveMember(ctx, "t1", "owner-user")
	if err == nil {
		t.Fatal("RemoveMember() should fail when target is the owner")
	}
}

func TestTenantUpdateMemberRole_OwnerProtected(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	repo.members["t1:owner-user"] = &entity.TenantMember{TenantID: "t1", UserID: "owner-user", Role: "owner"}

	uc := newTenantUCWithRepo(repo)
	_, err := uc.UpdateMemberRole(ctx, "t1", "owner-user", "admin")
	if err == nil {
		t.Fatal("UpdateMemberRole() should fail when target is the owner")
	}
}

func TestTenantUpdateMemberRole_CannotSetOwner(t *testing.T) {
	repo := newRichTenantRepo()
	ctx := context.Background()

	repo.members["t1:admin-user"] = &entity.TenantMember{TenantID: "t1", UserID: "admin-user", Role: "admin"}

	uc := newTenantUCWithRepo(repo)
	_, err := uc.UpdateMemberRole(ctx, "t1", "admin-user", "owner")
	if err == nil {
		t.Fatal("UpdateMemberRole() should fail when trying to set owner role")
	}
}
