package service

import (
	"context"

	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/pkg/pagination"
)

type UserService struct {
	userpb.UnimplementedUserServiceServer

	uc *biz.UserUsecase
}

func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) CurrentUserInfo(ctx context.Context, req *userpb.CurrentUserInfoRequest) (*userpb.CurrentUserInfoResponse, error) {
	callerID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.uc.CurrentUserInfo(ctx, callerID)
	if err != nil {
		return nil, err
	}
	return &userpb.CurrentUserInfoResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	u, err := s.uc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserResponse{User: userInfoMapper.Map(u)}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	_, tenantID, err := requireTenantScope(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := pagination.ExtractPage(req.GetPagination())
	users, total, err := s.uc.ListUsers(ctx, tenantID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &userpb.ListUsersResponse{
		Users:      userInfoMapper.MapSlice(users),
		Pagination: pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	callerID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := s.uc.UpdateUser(ctx, callerID, &entity.User{
		ID:       req.Id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateUserResponse{
		User: userInfoMapper.Map(updated),
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	_, tenantID, err := requireTenantScope(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.uc.CreateUser(ctx, tenantID, req.OrganizationId, &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.CreateUserResponse{Id: user.ID}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	success, err := s.uc.DeleteUser(ctx, &entity.User{
		ID: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.DeleteUserResponse{Success: success}, nil
}

func (s *UserService) PurgeUser(ctx context.Context, req *userpb.PurgeUserRequest) (*userpb.PurgeUserResponse, error) {
	success, err := s.uc.PurgeUser(ctx, &entity.User{ID: req.Id})
	if err != nil {
		return nil, err
	}
	return &userpb.PurgeUserResponse{Success: success}, nil
}

func (s *UserService) RestoreUser(ctx context.Context, req *userpb.RestoreUserRequest) (*userpb.RestoreUserResponse, error) {
	u, err := s.uc.RestoreUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &userpb.RestoreUserResponse{User: userInfoMapper.Map(u)}, nil
}
