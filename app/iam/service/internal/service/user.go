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
	user, err := s.uc.CurrentUserInfo(ctx)
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

func (s *UserService) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	page, pageSize := pagination.ExtractPage(req.GetPagination())
	users, total, err := s.uc.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	respUsers := userInfoMapper.MapSlice(users)

	return &userpb.ListUsersResponse{
		Users:      respUsers,
		Pagination: pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	updated, err := s.uc.UpdateUser(ctx, &entity.User{
		ID:       req.Id,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateUserResponse{
		User: userInfoMapper.Map(updated),
	}, nil
}

func (s *UserService) SaveUser(ctx context.Context, req *userpb.SaveUserRequest) (*userpb.SaveUserResponse, error) {
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
	user, err := s.uc.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return &userpb.SaveUserResponse{Id: user.ID}, nil
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
