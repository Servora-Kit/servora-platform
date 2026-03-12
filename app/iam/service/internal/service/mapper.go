package service

import (
	orgpb "github.com/Servora-Kit/servora/api/gen/go/organization/service/v1"
	projectpb "github.com/Servora-Kit/servora/api/gen/go/project/service/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/pkg/mapper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var userInfoMapper = mapper.NewForwardMapper(func(u *entity.User) *userpb.UserInfo {
	return &userpb.UserInfo{
		Id:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
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

var projectInfoMapper = mapper.NewForwardMapper(func(p *entity.Project) *projectpb.ProjectInfo {
	return &projectpb.ProjectInfo{
		Id:             p.ID,
		OrganizationId: p.OrganizationID,
		Name:           p.Name,
		Slug:           p.Slug,
		Description:    p.Description,
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
	}
})

var projectMemberInfoMapper = mapper.NewForwardMapper(func(m *entity.ProjectMember) *projectpb.ProjectMemberInfo {
	return &projectpb.ProjectMemberInfo{
		Id:        m.ID,
		ProjectId: m.ProjectID,
		UserId:    m.UserID,
		UserName:  m.UserName,
		UserEmail: m.UserEmail,
		Role:      m.Role,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
})
