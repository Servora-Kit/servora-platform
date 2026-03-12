package service

import (
	"context"

	projectpb "github.com/Servora-Kit/servora/api/gen/go/project/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/pkg/pagination"
)

type ProjectService struct {
	projectpb.UnimplementedProjectServiceServer

	uc *biz.ProjectUsecase
}

func NewProjectService(uc *biz.ProjectUsecase) *ProjectService {
	return &ProjectService{uc: uc}
}

func (s *ProjectService) CreateProject(ctx context.Context, req *projectpb.CreateProjectRequest) (*projectpb.CreateProjectResponse, error) {
	p, err := s.uc.Create(ctx, &entity.Project{
		OrganizationID: req.OrganizationId,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &projectpb.CreateProjectResponse{Project: projectInfoMapper.Map(p)}, nil
}

func (s *ProjectService) GetProject(ctx context.Context, req *projectpb.GetProjectRequest) (*projectpb.GetProjectResponse, error) {
	p, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &projectpb.GetProjectResponse{Project: projectInfoMapper.Map(p)}, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, req *projectpb.ListProjectsRequest) (*projectpb.ListProjectsResponse, error) {
	page, pageSize := pagination.ExtractPage(req.Pagination)
	projects, total, err := s.uc.List(ctx, req.OrganizationId, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := projectInfoMapper.MapSlice(projects)
	return &projectpb.ListProjectsResponse{
		Projects:   items,
		Pagination: pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, req *projectpb.UpdateProjectRequest) (*projectpb.UpdateProjectResponse, error) {
	p, err := s.uc.Update(ctx, &entity.Project{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &projectpb.UpdateProjectResponse{Project: projectInfoMapper.Map(p)}, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, req *projectpb.DeleteProjectRequest) (*projectpb.DeleteProjectResponse, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &projectpb.DeleteProjectResponse{Success: true}, nil
}

func (s *ProjectService) PurgeProject(ctx context.Context, req *projectpb.PurgeProjectRequest) (*projectpb.PurgeProjectResponse, error) {
	if err := s.uc.Purge(ctx, req.Id); err != nil {
		return nil, err
	}
	return &projectpb.PurgeProjectResponse{Success: true}, nil
}

func (s *ProjectService) RestoreProject(ctx context.Context, req *projectpb.RestoreProjectRequest) (*projectpb.RestoreProjectResponse, error) {
	p, err := s.uc.Restore(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &projectpb.RestoreProjectResponse{Project: projectInfoMapper.Map(p)}, nil
}

func (s *ProjectService) AddMember(ctx context.Context, req *projectpb.AddMemberRequest) (*projectpb.AddMemberResponse, error) {
	m, err := s.uc.AddMember(ctx, &entity.ProjectMember{
		ProjectID: req.ProjectId,
		UserID:    req.UserId,
		Role:      req.Role,
	})
	if err != nil {
		return nil, err
	}
	return &projectpb.AddMemberResponse{Member: projectMemberInfoMapper.Map(m)}, nil
}

func (s *ProjectService) RemoveMember(ctx context.Context, req *projectpb.RemoveMemberRequest) (*projectpb.RemoveMemberResponse, error) {
	if err := s.uc.RemoveMember(ctx, req.ProjectId, req.UserId); err != nil {
		return nil, err
	}
	return &projectpb.RemoveMemberResponse{Success: true}, nil
}

func (s *ProjectService) ListMembers(ctx context.Context, req *projectpb.ListMembersRequest) (*projectpb.ListMembersResponse, error) {
	page, pageSize := pagination.ExtractPage(req.Pagination)
	members, total, err := s.uc.ListMembers(ctx, req.ProjectId, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := projectMemberInfoMapper.MapSlice(members)
	return &projectpb.ListMembersResponse{
		Members:    items,
		Pagination: pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *ProjectService) UpdateMemberRole(ctx context.Context, req *projectpb.UpdateMemberRoleRequest) (*projectpb.UpdateMemberRoleResponse, error) {
	m, err := s.uc.UpdateMemberRole(ctx, req.ProjectId, req.UserId, req.Role)
	if err != nil {
		return nil, err
	}
	return &projectpb.UpdateMemberRoleResponse{Member: projectMemberInfoMapper.Map(m)}, nil
}
