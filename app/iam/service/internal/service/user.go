package service

import (
	"context"

	userpb "github.com/Servora-Kit/servora/api/gen/go/servora/user/service/v1"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
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
	return &userpb.CurrentUserInfoResponse{User: user}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	u, err := s.uc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserResponse{User: u}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	page, pageSize := pagination.ExtractPage(req.GetPagination())
	users, total, err := s.uc.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &userpb.ListUsersResponse{
		Users:      users,
		Pagination: pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	callerID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data := req.GetData()
	if data != nil {
		data.Id = req.Id
	}
	updated, err := s.uc.UpdateUser(ctx, callerID, data)
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateUserResponse{User: updated}, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user, err := s.uc.CreateUser(ctx, req.GetData(), req.Password)
	if err != nil {
		return nil, err
	}
	return &userpb.CreateUserResponse{User: user}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	success, err := s.uc.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &userpb.DeleteUserResponse{Success: success}, nil
}

func (s *UserService) PurgeUser(ctx context.Context, req *userpb.PurgeUserRequest) (*userpb.PurgeUserResponse, error) {
	success, err := s.uc.PurgeUser(ctx, req.Id)
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
	return &userpb.RestoreUserResponse{User: u}, nil
}
