package biz

import (
	"testing"
)

func TestRole_Level(t *testing.T) {
	tests := []struct {
		role  Role
		level int
	}{
		{RoleOwner, 3},
		{RoleAdmin, 2},
		{RoleMember, 1},
		{RoleViewer, 0},
		{Role("unknown"), -1},
	}
	for _, tt := range tests {
		if got := tt.role.Level(); got != tt.level {
			t.Errorf("Role(%q).Level() = %d, want %d", tt.role, got, tt.level)
		}
	}
}

func TestRole_IsOwner(t *testing.T) {
	if !RoleOwner.IsOwner() {
		t.Error("RoleOwner.IsOwner() = false, want true")
	}
	if RoleAdmin.IsOwner() {
		t.Error("RoleAdmin.IsOwner() = true, want false")
	}
	if RoleMember.IsOwner() {
		t.Error("RoleMember.IsOwner() = true, want false")
	}
}

func TestRole_CanManageMembers(t *testing.T) {
	if !RoleOwner.CanManageMembers() {
		t.Error("RoleOwner.CanManageMembers() = false, want true")
	}
	if !RoleAdmin.CanManageMembers() {
		t.Error("RoleAdmin.CanManageMembers() = false, want true")
	}
	if RoleMember.CanManageMembers() {
		t.Error("RoleMember.CanManageMembers() = true, want false")
	}
	if RoleViewer.CanManageMembers() {
		t.Error("RoleViewer.CanManageMembers() = true, want false")
	}
}

func TestValidateTenantRole(t *testing.T) {
	if err := ValidateTenantRole("admin"); err != nil {
		t.Errorf("ValidateTenantRole(admin) = %v, want nil", err)
	}
	if err := ValidateTenantRole("member"); err != nil {
		t.Errorf("ValidateTenantRole(member) = %v, want nil", err)
	}
	if err := ValidateTenantRole("owner"); err == nil {
		t.Error("ValidateTenantRole(owner) = nil, want error (owner not directly settable)")
	}
	if err := ValidateTenantRole("viewer"); err == nil {
		t.Error("ValidateTenantRole(viewer) = nil, want error (not a tenant role)")
	}
	if err := ValidateTenantRole("invalid"); err == nil {
		t.Error("ValidateTenantRole(invalid) = nil, want error")
	}
}

func TestValidateOrganizationRole(t *testing.T) {
	if err := ValidateOrganizationRole("admin"); err != nil {
		t.Errorf("ValidateOrganizationRole(admin) = %v, want nil", err)
	}
	if err := ValidateOrganizationRole("member"); err != nil {
		t.Errorf("ValidateOrganizationRole(member) = %v, want nil", err)
	}
	if err := ValidateOrganizationRole("viewer"); err != nil {
		t.Errorf("ValidateOrganizationRole(viewer) = %v, want nil", err)
	}
	if err := ValidateOrganizationRole("owner"); err == nil {
		t.Error("ValidateOrganizationRole(owner) = nil, want error (owner not directly settable)")
	}
}
