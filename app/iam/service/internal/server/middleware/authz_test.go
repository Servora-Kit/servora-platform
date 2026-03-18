package middleware

import (
	"testing"

	authzpb "github.com/Servora-Kit/servora/api/gen/go/authz/service/v1"
	iamv1 "github.com/Servora-Kit/servora/api/gen/go/iam/service/v1"
	orgpb "github.com/Servora-Kit/servora/api/gen/go/organization/service/v1"
	"github.com/Servora-Kit/servora/pkg/actor"
)

func userActorWithScope(orgID string) *actor.UserActor {
	ua := actor.NewUserActor("user-1", "Test", "test@example.com", nil)
	if orgID != "" {
		ua.SetOrganizationID(orgID)
	}
	return ua
}

func TestResolveObject_OrgScope_FromActor(t *testing.T) {
	ua := userActorWithScope("org-uuid-1")
	rule := iamv1.AuthzRuleEntry{
		Mode:     authzpb.AuthzMode_AUTHZ_MODE_ORGANIZATION,
		Relation: authzpb.Relation_RELATION_CAN_VIEW,
		IDField:  "",
	}

	objType, objID, err := resolveObject(rule, nil, ua)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if objType != "organization" {
		t.Errorf("objectType = %q, want %q", objType, "organization")
	}
	if objID != "org-uuid-1" {
		t.Errorf("objectID = %q, want %q", objID, "org-uuid-1")
	}
}

func TestResolveObject_OrgScope_MissingHeader(t *testing.T) {
	ua := userActorWithScope("")
	rule := iamv1.AuthzRuleEntry{
		Mode:     authzpb.AuthzMode_AUTHZ_MODE_ORGANIZATION,
		Relation: authzpb.Relation_RELATION_CAN_VIEW,
		IDField:  "",
	}

	_, _, err := resolveObject(rule, nil, ua)
	if err == nil {
		t.Fatal("expected error for missing org scope")
	}
}

func TestResolveObject_OrgResource_FromRequest(t *testing.T) {
	ua := userActorWithScope("other-org")
	rule := iamv1.AuthzRuleEntry{
		Mode:     authzpb.AuthzMode_AUTHZ_MODE_ORGANIZATION,
		Relation: authzpb.Relation_RELATION_CAN_VIEW,
		IDField:  "id",
	}
	req := &orgpb.GetOrganizationRequest{Id: "org-from-request"}

	objType, objID, err := resolveObject(rule, req, ua)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if objType != "organization" {
		t.Errorf("objectType = %q, want %q", objType, "organization")
	}
	if objID != "org-from-request" {
		t.Errorf("objectID = %q, want %q", objID, "org-from-request")
	}
}

func TestResolveObject_TenantRoot(t *testing.T) {
	ua := userActorWithScope("")
	ua.SetTenantID("tenant-root-id")
	rule := iamv1.AuthzRuleEntry{
		Mode:       authzpb.AuthzMode_AUTHZ_MODE_OBJECT,
		ObjectType: authzpb.ObjectType_OBJECT_TYPE_TENANT,
		Relation:   authzpb.Relation_RELATION_ADMIN,
		IDField:    "",
	}

	objType, objID, err := resolveObject(rule, nil, ua)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if objType != "tenant" {
		t.Errorf("objectType = %q, want %q", objType, "tenant")
	}
	if objID != "tenant-root-id" {
		t.Errorf("objectID = %q, want %q", objID, "tenant-root-id")
	}
}

func TestScopeFromActor_OrganizationID(t *testing.T) {
	ua := userActorWithScope("org-123")
	id, err := scopeFromActor(ua, "OrganizationID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "org-123" {
		t.Errorf("id = %q, want %q", id, "org-123")
	}
}

func TestScopeFromActor_MissingOrg(t *testing.T) {
	ua := userActorWithScope("")
	_, err := scopeFromActor(ua, "OrganizationID")
	if err == nil {
		t.Fatal("expected error for missing org")
	}
}

func TestScopeFromActor_UnknownField(t *testing.T) {
	ua := userActorWithScope("org")
	_, err := scopeFromActor(ua, "UnknownField")
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}
