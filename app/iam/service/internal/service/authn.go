package service

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"

	authnpb "github.com/Servora-Kit/servora/api/gen/go/servora/authn/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/pkg/actor"
	"github.com/Servora-Kit/servora/pkg/cap"
)

type AuthnService struct {
	authnpb.UnimplementedAuthnServiceServer

	uc  *biz.AuthnUsecase
	cap *cap.Cap
}

func NewAuthnService(uc *biz.AuthnUsecase, capSvc *cap.Cap) *AuthnService {
	return &AuthnService{uc: uc, cap: capSvc}
}

func (s *AuthnService) SignupByEmail(ctx context.Context, req *authnpb.SignupByEmailRequest) (*authnpb.SignupByEmailResponse, error) {
	if req.Password != req.PasswordConfirm {
		return nil, errors.BadRequest("INVALID_REQUEST", "password and confirm password do not match")
	}

	if req.CapToken == "" {
		return nil, authnpb.ErrorInvalidCaptcha("captcha token is required")
	}
	valid, err := s.cap.ValidateToken(ctx, req.CapToken)
	if err != nil || !valid {
		return nil, authnpb.ErrorInvalidCaptcha("invalid or expired captcha token")
	}

	user, err := s.uc.SignupByEmail(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &authnpb.SignupByEmailResponse{
		Id:    user.Id,
		Name:  user.Username,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *AuthnService) LoginByEmailPassword(ctx context.Context, req *authnpb.LoginByEmailPasswordRequest) (*authnpb.LoginByEmailPasswordResponse, error) {
	tokenPair, err := s.uc.LoginByEmailPassword(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &authnpb.LoginByEmailPasswordResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *AuthnService) RefreshToken(ctx context.Context, req *authnpb.RefreshTokenRequest) (*authnpb.RefreshTokenResponse, error) {
	tokenPair, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &authnpb.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *AuthnService) Logout(ctx context.Context, req *authnpb.LogoutRequest) (*authnpb.LogoutResponse, error) {
	if err := s.uc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, err
	}
	return &authnpb.LogoutResponse{
		Success: true,
	}, nil
}

func (s *AuthnService) ChangePassword(ctx context.Context, req *authnpb.ChangePasswordRequest) (*authnpb.ChangePasswordResponse, error) {
	if req.NewPassword != req.NewPasswordConfirm {
		return nil, errors.BadRequest("INVALID_REQUEST", "new password and confirm password do not match")
	}

	a, ok := actor.FromContext(ctx)
	if !ok || a.Type() != actor.TypeUser {
		return nil, errors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}

	if err := s.uc.ChangePassword(ctx, a.ID(), req.CurrentPassword, req.NewPassword); err != nil {
		return nil, err
	}
	return &authnpb.ChangePasswordResponse{Success: true}, nil
}

func (s *AuthnService) LogoutAllDevices(ctx context.Context, _ *authnpb.LogoutAllDevicesRequest) (*authnpb.LogoutAllDevicesResponse, error) {
	a, ok := actor.FromContext(ctx)
	if !ok || a.Type() != actor.TypeUser {
		return nil, errors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}

	if err := s.uc.LogoutAllDevices(ctx, a.ID()); err != nil {
		return nil, err
	}
	return &authnpb.LogoutAllDevicesResponse{Success: true}, nil
}

func (s *AuthnService) RequestEmailVerification(ctx context.Context, req *authnpb.RequestEmailVerificationRequest) (*authnpb.RequestEmailVerificationResponse, error) {
	if err := s.uc.RequestEmailVerification(ctx, req.Email); err != nil {
		return nil, err
	}
	return &authnpb.RequestEmailVerificationResponse{Success: true}, nil
}

func (s *AuthnService) VerifyEmail(ctx context.Context, req *authnpb.VerifyEmailRequest) (*authnpb.VerifyEmailResponse, error) {
	if err := s.uc.VerifyEmail(ctx, req.Token); err != nil {
		return nil, err
	}
	return &authnpb.VerifyEmailResponse{Success: true}, nil
}

func (s *AuthnService) RequestPasswordReset(ctx context.Context, req *authnpb.RequestPasswordResetRequest) (*authnpb.RequestPasswordResetResponse, error) {
	if err := s.uc.RequestPasswordReset(ctx, req.Email); err != nil {
		return nil, err
	}
	return &authnpb.RequestPasswordResetResponse{Success: true}, nil
}

func (s *AuthnService) ResetPassword(ctx context.Context, req *authnpb.ResetPasswordRequest) (*authnpb.ResetPasswordResponse, error) {
	if req.NewPassword != req.NewPasswordConfirm {
		return nil, errors.BadRequest("INVALID_REQUEST", "new password and confirm password do not match")
	}
	if err := s.uc.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		return nil, err
	}
	return &authnpb.ResetPasswordResponse{Success: true}, nil
}

// VerifyAuthorizationHeader 校验 Bearer access token 并返回 user ID，供网关 ForwardAuth 使用。
func (s *AuthnService) VerifyAuthorizationHeader(authHeader string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", authnpb.ErrorMissingToken("invalid authorization header format")
	}
	userID, err := s.uc.VerifyAccessToken(parts[1])
	if err != nil {
		return "", err
	}
	if userID == "" {
		return "", authnpb.ErrorInvalidCredentials("missing user id in token")
	}
	return userID, nil
}
