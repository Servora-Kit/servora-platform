package biz

import "fmt"

// Role represents an IAM role within a tenant or organization.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer" // org only
)

// IsOwner returns true if the role is owner.
func (r Role) IsOwner() bool { return r == RoleOwner }

// CanManageMembers returns true if the role is owner or admin.
func (r Role) CanManageMembers() bool { return r == RoleOwner || r == RoleAdmin }

// Level returns the numeric level of the role for hierarchy comparisons.
// Higher value means higher privilege.
func (r Role) Level() int {
	switch r {
	case RoleOwner:
		return 3
	case RoleAdmin:
		return 2
	case RoleMember:
		return 1
	case RoleViewer:
		return 0
	default:
		return -1
	}
}

// String returns the string representation of the role.
func (r Role) String() string { return string(r) }

// ValidateTenantRole validates that a role string is valid for tenant membership.
// Owner is not directly settable via UpdateMemberRole (only via TransferOwnership).
func ValidateTenantRole(role string) error {
	r := Role(role)
	if r != RoleAdmin && r != RoleMember {
		return fmt.Errorf("invalid tenant role %q; allowed: admin, member", role)
	}
	return nil
}

// ValidateOrganizationRole validates that a role string is valid for organization membership.
// Owner is not directly settable via UpdateMemberRole (only via TransferOwnership).
func ValidateOrganizationRole(role string) error {
	r := Role(role)
	if r != RoleAdmin && r != RoleMember && r != RoleViewer {
		return fmt.Errorf("invalid organization role %q; allowed: admin, member, viewer", role)
	}
	return nil
}
