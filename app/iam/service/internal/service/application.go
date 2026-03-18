package service

import (
	"context"
	"time"

	apppb "github.com/Servora-Kit/servora/api/gen/go/application/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
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
	_, tenantID, err := requireTenantScope(ctx)
	if err != nil {
		return nil, err
	}

	appType := "web"
	if req.ApplicationType != nil {
		appType = *req.ApplicationType
	}
	tokenType := "jwt"
	if req.AccessTokenType != nil {
		tokenType = *req.AccessTokenType
	}
	lifetime := time.Duration(3600) * time.Second
	if req.IdTokenLifetime != nil && *req.IdTokenLifetime > 0 {
		lifetime = time.Duration(*req.IdTokenLifetime) * time.Second
	}

	app, secret, err := s.uc.Create(ctx, &entity.Application{
		Name:            req.Name,
		RedirectURIs:    req.RedirectUris,
		Scopes:          req.Scopes,
		GrantTypes:      req.GrantTypes,
		ApplicationType: appType,
		AccessTokenType: tokenType,
		TenantID:        tenantID,
		IDTokenLifetime: lifetime,
	})
	if err != nil {
		return nil, err
	}
	return &apppb.CreateApplicationResponse{
		Application:  applicationInfoMapper.Map(app),
		ClientSecret: secret,
	}, nil
}

func (s *ApplicationService) GetApplication(ctx context.Context, req *apppb.GetApplicationRequest) (*apppb.GetApplicationResponse, error) {
	app, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &apppb.GetApplicationResponse{Application: applicationInfoMapper.Map(app)}, nil
}

func (s *ApplicationService) ListApplications(ctx context.Context, req *apppb.ListApplicationsRequest) (*apppb.ListApplicationsResponse, error) {
	_, tenantID, err := requireTenantScope(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := pagination.ExtractPage(req.Pagination)
	apps, total, err := s.uc.List(ctx, tenantID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &apppb.ListApplicationsResponse{
		Applications: applicationInfoMapper.MapSlice(apps),
		Pagination:   pagination.BuildPageResponse(total, page, pageSize),
	}, nil
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, req *apppb.UpdateApplicationRequest) (*apppb.UpdateApplicationResponse, error) {
	app, err := s.uc.Update(ctx, &entity.Application{
		ID:              req.Id,
		Name:            req.Name,
		RedirectURIs:    req.RedirectUris,
		Scopes:          req.Scopes,
		GrantTypes:      req.GrantTypes,
		IDTokenLifetime: time.Duration(req.IdTokenLifetime) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &apppb.UpdateApplicationResponse{Application: applicationInfoMapper.Map(app)}, nil
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
