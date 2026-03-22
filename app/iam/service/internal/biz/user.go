package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"

	authnpb "github.com/Servora-Kit/servora/api/gen/go/servora/authn/service/v1"
	"github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/servora/user/service/v1"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/Servora-Kit/servora/pkg/logger"
)

type UserRepo interface {
	SaveUser(ctx context.Context, user *userpb.User, hashedPassword string) (*userpb.User, error)
	GetUserById(context.Context, string) (*userpb.User, error)
	DeleteUser(context.Context, string) error
	PurgeUser(context.Context, string) error
	PurgeCascade(ctx context.Context, id string) error
	RestoreUser(context.Context, string) (*userpb.User, error)
	GetUserByIdIncludingDeleted(context.Context, string) (*userpb.User, error)
	UpdateUser(context.Context, *userpb.User) (*userpb.User, error)
	ListUsers(context.Context, int32, int32) ([]*userpb.User, int64, error)
}

type UserUsecase struct {
	repo      UserRepo
	log       *logger.Helper
	cfg       *conf.App
	authnRepo AuthnRepo
	authnUC   *AuthnUsecase
	authz     AuthZRepo
}

func NewUserUsecase(
	repo UserRepo,
	l logger.Logger,
	cfg *conf.App,
	authnRepo AuthnRepo,
	authnUC *AuthnUsecase,
	authz AuthZRepo,
) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		log:       logger.For(l, "user/biz/iam"),
		cfg:       cfg,
		authnRepo: authnRepo,
		authnUC:   authnUC,
		authz:     authz,
	}
}

func (uc *UserUsecase) CurrentUserInfo(ctx context.Context, callerID string) (*userpb.User, error) {
	if callerID == "" {
		return nil, authnpb.ErrorUnauthorized("user not authenticated")
	}
	u, err := uc.repo.GetUserById(ctx, callerID)
	if err != nil {
		return nil, userpb.ErrorUserNotFound("user not found")
	}
	return u, nil
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (*userpb.User, error) {
	u, err := uc.repo.GetUserById(ctx, id)
	if err != nil {
		uc.log.Errorf("get user by id failed: %v", err)
		return nil, userpb.ErrorUserNotFound("user not found")
	}
	return u, nil
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, callerID string, user *userpb.User) (*userpb.User, error) {
	if callerID == "" {
		return nil, authnpb.ErrorUnauthorized("user not authenticated")
	}

	origUser, err := uc.repo.GetUserById(ctx, user.Id)
	if err != nil {
		uc.log.Errorf("get user failed: %v", err)
		return nil, userpb.ErrorUserNotFound("user not found")
	}

	if callerID != user.Id {
		return nil, authnpb.ErrorUnauthorized("you can only update your own information")
	}

	if user.Username != "" && user.Username != origUser.Username {
		userWithSameName, err := uc.authnRepo.GetUserByUserName(ctx, user.Username)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				uc.log.Errorf("check username failed: %v", err)
				return nil, errors.InternalServer("INTERNAL", "internal error")
			}
		}
		if userWithSameName != nil {
			return nil, authnpb.ErrorUserAlreadyExists("username already exists")
		}
	}

	if user.Email != "" && user.Email != origUser.Email {
		userWithSameEmail, err := uc.authnRepo.GetUserByEmail(ctx, user.Email)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				uc.log.Errorf("check email failed: %v", err)
				return nil, errors.InternalServer("INTERNAL", "internal error")
			}
		}
		if userWithSameEmail != nil {
			return nil, authnpb.ErrorUserAlreadyExists("email already exists")
		}
	}

	updatedUser, err := uc.repo.UpdateUser(ctx, user)
	if err != nil {
		uc.log.Errorf("update user failed: %v", err)
		return nil, userpb.ErrorUpdateUserFailed("failed to update user")
	}
	return updatedUser, nil
}

// CreateUser creates a new user in IAM.
// The created user starts with email_verified=false; a verification email is sent.
func (uc *UserUsecase) CreateUser(ctx context.Context, user *userpb.User, password string) (*userpb.User, error) {
	if err := uc.checkUserExists(ctx, user.Username, user.Email); err != nil {
		return nil, err
	}

	if user.Role == "" {
		user.Role = "user"
	}
	user.EmailVerified = false

	hashedPassword, err := helpers.BcryptHash(password)
	if err != nil {
		uc.log.Errorf("hash password failed: %v", err)
		return nil, userpb.ErrorCreateUserFailed("failed to create user")
	}

	savedUser, err := uc.repo.SaveUser(ctx, user, hashedPassword)
	if err != nil {
		uc.log.Errorf("create user failed: %v", err)
		return nil, userpb.ErrorCreateUserFailed("failed to create user")
	}

	if uc.authnUC != nil {
		if err := uc.authnUC.SendVerificationEmail(ctx, savedUser); err != nil {
			uc.log.Warnf("send verification email failed for user %s: %v", savedUser.Id, err)
		}
	}

	return savedUser, nil
}

func (uc *UserUsecase) ListUsers(ctx context.Context, page, pageSize int32) ([]*userpb.User, int64, error) {
	users, total, err := uc.repo.ListUsers(ctx, page, pageSize)
	if err != nil {
		uc.log.Errorf("list users failed: %v", err)
		return nil, 0, errors.InternalServer("INTERNAL", "internal error")
	}
	return users, total, nil
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) (bool, error) {
	if _, err := uc.repo.GetUserById(ctx, id); err != nil {
		return false, userpb.ErrorUserNotFound("user not found")
	}
	if err := uc.repo.DeleteUser(ctx, id); err != nil {
		uc.log.Errorf("soft delete user failed: %v", err)
		return false, userpb.ErrorDeleteUserFailed("failed to delete user")
	}
	return true, nil
}

func (uc *UserUsecase) PurgeUser(ctx context.Context, id string) (bool, error) {
	uc.log.Infof("PurgeUser start: user_id=%s", id)

	if err := uc.repo.PurgeCascade(ctx, id); err != nil {
		uc.log.Errorf("PurgeUser PurgeCascade failed: user_id=%s err=%v", id, err)
		return false, userpb.ErrorDeleteUserFailed("failed to delete user")
	}

	if err := uc.authnRepo.DeleteUserRefreshTokens(ctx, id); err != nil {
		uc.log.Warnf("PurgeUser Redis cleanup partial failure: user_id=%s err=%v", id, err)
	}

	uc.log.Infof("PurgeUser complete: user_id=%s", id)
	return true, nil
}

func (uc *UserUsecase) RestoreUser(ctx context.Context, id string) (*userpb.User, error) {
	if _, err := uc.repo.GetUserByIdIncludingDeleted(ctx, id); err != nil {
		return nil, userpb.ErrorUserNotFound("user not found")
	}
	return uc.repo.RestoreUser(ctx, id)
}

func (uc *UserUsecase) checkUserExists(ctx context.Context, username, email string) error {
	existingUser, err := uc.authnRepo.GetUserByUserName(ctx, username)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			uc.log.Errorf("check username failed: %v", err)
			return errors.InternalServer("INTERNAL", "internal error")
		}
	}
	if existingUser != nil {
		return authnpb.ErrorUserAlreadyExists("username already exists")
	}

	existingEmail, err := uc.authnRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			uc.log.Errorf("check email failed: %v", err)
			return errors.InternalServer("INTERNAL", "internal error")
		}
	}
	if existingEmail != nil {
		return authnpb.ErrorUserAlreadyExists("email already exists")
	}
	return nil
}
