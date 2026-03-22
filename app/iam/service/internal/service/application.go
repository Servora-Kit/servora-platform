package service

import (
	"context"

	apppb "github.com/Servora-Kit/servora/api/gen/go/servora/application/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/pkg/pagination"
)

type ApplicationService struct {
	apppb.UnimplementedApplicationServiceServer

	uc *biz.ApplicationUsecase
}

func NewApplicationService(uc *biz.ApplicationUsecase) *ApplicationService {
	return &ApplicationService{uc: uc}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, req *apppb.CreateApplicationRequest) (*apppb.CreateApplicationResponse, error) {
	data := req.GetData()
	if data.ApplicationType == "" {
		data.ApplicationType = "web"
	}
	if data.Type == "" {
		data.Type = "web"
	}
	if data.AccessTokenType == "" {
		data.AccessTokenType = "jwt"
	}
	if data.IdTokenLifetime <= 0 {
		data.IdTokenLifetime = 3600
	}

	app, secret, err := s.uc.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	return &apppb.CreateApplicationResponse{
		Application:  app,
		ClientSecret: secret,
	}, nil
}

func (s *ApplicationService) GetApplication(ctx context.Context, req *apppb.GetApplicationRequest) (*apppb.GetApplicationResponse, error) {
	app, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &apppb.GetApplicationResponse{Application: app}, nil
}

func (s *ApplicationService) ListApplications(ctx context.Context, req *apppb.ListApplicationsRequest) (*apppb.ListApplicationsResponse, error) {
	page, pageSize := pagination.ExtractPage(req.Pagination)
	apps, total, err := s.uc.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &apppb.ListApplicationsResponse{
		Applications: apps,
		Pagination:   pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, req *apppb.UpdateApplicationRequest) (*apppb.UpdateApplicationResponse, error) {
	data := req.GetData()
	if data != nil {
		data.Id = req.Id
	}
	app, err := s.uc.Update(ctx, data)
	if err != nil {
		return nil, err
	}
	return &apppb.UpdateApplicationResponse{Application: app}, nil
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, req *apppb.DeleteApplicationRequest) (*apppb.DeleteApplicationResponse, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &apppb.DeleteApplicationResponse{Success: true}, nil
}

func (s *ApplicationService) RegenerateClientSecret(ctx context.Context, req *apppb.RegenerateClientSecretRequest) (*apppb.RegenerateClientSecretResponse, error) {
	secret, err := s.uc.RegenerateClientSecret(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &apppb.RegenerateClientSecretResponse{ClientSecret: secret}, nil
}
