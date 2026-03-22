package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	apppb "github.com/Servora-Kit/servora/api/gen/go/servora/application/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/application"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/mapper"
)

type applicationRepo struct {
	data   *Data
	log    *logger.Helper
	mapper *mapper.CopierMapper[apppb.Application, ent.Application]
}

func NewApplicationRepo(data *Data, l logger.Logger) biz.ApplicationRepo {
	return &applicationRepo{
		data:   data,
		log:    logger.For(l, "application/data/iam"),
		mapper: newApplicationMapper(),
	}
}

func (r *applicationRepo) Create(ctx context.Context, app *apppb.Application, clientSecretHash string) (*apppb.Application, error) {
	created, err := r.data.Ent(ctx).Application.Create().
		SetClientID(app.ClientId).
		SetClientSecretHash(clientSecretHash).
		SetName(app.Name).
		SetRedirectUris(app.RedirectUris).
		SetScopes(app.Scopes).
		SetGrantTypes(app.GrantTypes).
		SetApplicationType(app.ApplicationType).
		SetAccessTokenType(app.AccessTokenType).
		SetType(app.Type).
		SetIDTokenLifetime(int(app.IdTokenLifetime)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.MustToProto(created), nil
}

func (r *applicationRepo) GetByID(ctx context.Context, id string) (*apppb.Application, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid application id: %w", err)
	}
	a, err := r.data.Ent(ctx).Application.Query().
		Where(application.IDEQ(uid), application.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return r.mapper.MustToProto(a), nil
}

func (r *applicationRepo) GetByClientID(ctx context.Context, clientID string) (*apppb.Application, error) {
	a, err := r.data.Ent(ctx).Application.Query().
		Where(application.ClientIDEQ(clientID), application.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return r.mapper.MustToProto(a), nil
}

func (r *applicationRepo) List(ctx context.Context, page, pageSize int32) ([]*apppb.Application, int64, error) {
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	query := r.data.Ent(ctx).Application.Query().
		Where(application.DeletedAtIsNil()).
		Order(application.ByCreatedAt(sql.OrderDesc()))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	apps, err := query.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	result, err := r.mapper.ToProtoList(apps)
	if err != nil {
		return nil, 0, err
	}
	return result, int64(total), nil
}

func (r *applicationRepo) Update(ctx context.Context, app *apppb.Application) (*apppb.Application, error) {
	uid, err := uuid.Parse(app.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid application id: %w", err)
	}
	n, err := r.data.Ent(ctx).Application.Update().
		Where(application.IDEQ(uid), application.DeletedAtIsNil()).
		SetName(app.Name).
		SetRedirectUris(app.RedirectUris).
		SetScopes(app.Scopes).
		SetGrantTypes(app.GrantTypes).
		SetIDTokenLifetime(int(app.IdTokenLifetime)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("application not found")
	}
	return r.GetByID(ctx, app.Id)
}

func (r *applicationRepo) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid application id: %w", err)
	}
	n, err := r.data.Ent(ctx).Application.Update().
		Where(application.IDEQ(uid), application.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("application not found")
	}
	return nil
}

func (r *applicationRepo) UpdateClientSecretHash(ctx context.Context, id string, hash string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid application id: %w", err)
	}
	n, err := r.data.Ent(ctx).Application.Update().
		Where(application.IDEQ(uid), application.DeletedAtIsNil()).
		SetClientSecretHash(hash).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("application not found")
	}
	return nil
}
