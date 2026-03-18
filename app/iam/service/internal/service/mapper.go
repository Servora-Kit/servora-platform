package service

import (
	apppb "github.com/Servora-Kit/servora/api/gen/go/application/service/v1"
	orgpb "github.com/Servora-Kit/servora/api/gen/go/organization/service/v1"
	tenantpb "github.com/Servora-Kit/servora/api/gen/go/tenant/service/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/pkg/mapper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var userInfoMapper = mapper.NewForwardMapper(func(u *entity.User) *userpb.UserInfo {
	return &userpb.UserInfo{
		Id:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
	}
})

var orgInfoMapper = mapper.NewForwardMapper(func(o *entity.Organization) *orgpb.OrganizationInfo {
	return &orgpb.OrganizationInfo{
		Id:          o.ID,
		Name:        o.Name,
		Slug:        o.Slug,
		DisplayName: o.DisplayName,
		CreatedAt:   timestamppb.New(o.CreatedAt),
		UpdatedAt:   timestamppb.New(o.UpdatedAt),
	}
})

var orgMemberInfoMapper = mapper.NewForwardMapper(func(m *entity.OrganizationMember) *orgpb.OrganizationMemberInfo {
	return &orgpb.OrganizationMemberInfo{
		Id:             m.ID,
		OrganizationId: m.OrganizationID,
		UserId:         m.UserID,
		UserName:       m.UserName,
		UserEmail:      m.UserEmail,
		Role:           m.Role,
		CreatedAt:      timestamppb.New(m.CreatedAt),
	}
})

var tenantInfoMapper = mapper.NewForwardMapper(func(t *entity.Tenant) *tenantpb.TenantInfo {
	return &tenantpb.TenantInfo{
		Id:          t.ID,
		Slug:        t.Slug,
		Name:        t.Name,
		DisplayName: t.DisplayName,
		Kind:        t.Kind,
		Domain:      t.Domain,
		Status:      t.Status,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
})

var tenantMemberInfoMapper = mapper.NewForwardMapper(func(m *entity.TenantMember) *tenantpb.TenantMemberInfo {
	info := &tenantpb.TenantMemberInfo{
		Id:        m.ID,
		TenantId:  m.TenantID,
		UserId:    m.UserID,
		UserName:  m.UserName,
		UserEmail: m.UserEmail,
		Role:      m.Role,
		Status:    m.Status,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
	if m.JoinedAt != nil {
		info.JoinedAt = timestamppb.New(*m.JoinedAt)
	}
	return info
})

var applicationInfoMapper = mapper.NewForwardMapper(func(a *entity.Application) *apppb.ApplicationInfo {
	return &apppb.ApplicationInfo{
		Id:              a.ID,
		ClientId:        a.ClientID,
		Name:            a.Name,
		RedirectUris:    a.RedirectURIs,
		Scopes:          a.Scopes,
		GrantTypes:      a.GrantTypes,
		ApplicationType: a.ApplicationType,
		AccessTokenType: a.AccessTokenType,
		TenantId:        a.TenantID,
		IdTokenLifetime: int32(a.IDTokenLifetime.Seconds()),
		CreatedAt:       timestamppb.New(a.CreatedAt),
		UpdatedAt:       timestamppb.New(a.UpdatedAt),
	}
})
