package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	userpb "github.com/Servora-Kit/servora/api/gen/go/servora/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent/user"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/mapper"
)

type userRepo struct {
	data   *Data
	log    *logger.Helper
	mapper *mapper.CopierMapper[userpb.User, ent.User]
}

func NewUserRepo(data *Data, l logger.Logger) biz.UserRepo {
	return &userRepo{
		data:   data,
		log:    logger.For(l, "user/data/iam"),
		mapper: newUserMapper(),
	}
}

func (r *userRepo) SaveUser(ctx context.Context, u *userpb.User, hashedPassword string) (*userpb.User, error) {
	profileJSON := profileToJSON(u.Profile)
	b := r.data.Ent(ctx).User.Create().
		SetUsername(u.Username).
		SetEmail(u.Email).
		SetPassword(hashedPassword).
		SetPhone(u.Phone).
		SetPhoneVerified(u.PhoneVerified).
		SetRole(u.Role).
		SetEmailVerified(u.EmailVerified).
		SetProfile(profileJSON)

	if u.Status != "" {
		b.SetStatus(u.Status)
	}
	if u.EmailVerifiedAt != nil {
		b.SetEmailVerifiedAt(u.EmailVerifiedAt.AsTime())
	}
	if u.Id != "" {
		uid, err := uuid.Parse(u.Id)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
		b.SetID(uid)
	}

	created, err := b.Save(ctx)
	if err != nil {
		r.log.Errorf("SaveUser failed: %v", err)
		return nil, err
	}
	return r.mapper.MustToProto(created), nil
}

func (r *userRepo) GetUserById(ctx context.Context, id string) (*userpb.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	entUser, err := r.data.Ent(ctx).User.Query().Where(user.IDEQ(uid), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return r.mapper.MustToProto(entUser), nil
}

func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	return r.data.Ent(ctx).User.UpdateOneID(uid).SetDeletedAt(time.Now()).Exec(ctx)
}

func (r *userRepo) PurgeUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	return r.data.Ent(ctx).User.DeleteOneID(uid).Exec(ctx)
}

func (r *userRepo) PurgeCascade(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	return r.data.Ent(ctx).User.DeleteOneID(uid).Exec(ctx)
}

func (r *userRepo) RestoreUser(ctx context.Context, id string) (*userpb.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	u, err := r.data.Ent(ctx).User.UpdateOneID(uid).ClearDeletedAt().Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.MustToProto(u), nil
}

func (r *userRepo) GetUserByIdIncludingDeleted(ctx context.Context, id string) (*userpb.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	entUser, err := r.data.Ent(ctx).User.Query().Where(user.IDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return r.mapper.MustToProto(entUser), nil
}

func (r *userRepo) UpdateUser(ctx context.Context, u *userpb.User) (*userpb.User, error) {
	uid, err := uuid.Parse(u.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profileJSON := profileToJSON(u.Profile)
	upd := r.data.Ent(ctx).User.UpdateOneID(uid).
		SetProfile(profileJSON)

	if u.Username != "" {
		upd.SetUsername(u.Username)
	}
	if u.Email != "" {
		upd.SetEmail(u.Email)
	}
	if u.Phone != "" {
		upd.SetPhone(u.Phone)
	}
	if u.Role != "" {
		upd.SetRole(u.Role)
	}
	if u.Status != "" {
		upd.SetStatus(u.Status)
	}

	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.MustToProto(updated), nil
}

func (r *userRepo) ListUsers(ctx context.Context, page int32, pageSize int32) ([]*userpb.User, int64, error) {
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	query := r.data.Ent(ctx).User.Query().Where(user.DeletedAtIsNil()).Order(user.ByID(sql.OrderDesc()))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	entUsers, err := query.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	users, err := r.mapper.ToProtoList(entUsers)
	if err != nil {
		return nil, 0, err
	}
	return users, int64(total), nil
}

func profileToJSON(p *userpb.UserProfile) map[string]interface{} {
	if p == nil {
		return nil
	}
	b, _ := json.Marshal(p)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}
