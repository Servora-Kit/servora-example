package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"

	authpb "github.com/Servora-Kit/servora/api/gen/go/auth/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/pkg/actor"
)

// AuthService is a auth service.
type AuthService struct {
	authpb.UnimplementedAuthServiceServer

	uc *biz.AuthUsecase
}

// NewAuthService new a auth service.
func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) SignupByEmail(ctx context.Context, req *authpb.SignupByEmailRequest) (*authpb.SignupByEmailResponse, error) {
	// 参数校验
	if req.Password != req.PasswordConfirm {
		return nil, errors.BadRequest("INVALID_REQUEST", "password and confirm password do not match")
	}
	// 调用 biz 层
	user, err := s.uc.SignupByEmail(ctx, &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	// 拼装返回结果
	return &authpb.SignupByEmailResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

// LoginByEmailPassword user login by email and password.
func (s *AuthService) LoginByEmailPassword(ctx context.Context, req *authpb.LoginByEmailPasswordRequest) (*authpb.LoginByEmailPasswordResponse, error) {
	user := &entity.User{
		Email:    req.Email,
		Password: req.Password,
	}
	tokenPair, err := s.uc.LoginByEmailPassword(ctx, user)
	if err != nil {
		return nil, err
	}
	return &authpb.LoginByEmailPasswordResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken refreshes the access token using a valid refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	tokenPair, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &authpb.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Logout invalidates the refresh token
func (s *AuthService) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if err := s.uc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, err
	}
	return &authpb.LogoutResponse{
		Success: true,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
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
	return &authpb.ChangePasswordResponse{Success: true}, nil
}

func (s *AuthService) LogoutAllDevices(ctx context.Context, _ *authpb.LogoutAllDevicesRequest) (*authpb.LogoutAllDevicesResponse, error) {
	a, ok := actor.FromContext(ctx)
	if !ok || a.Type() != actor.TypeUser {
		return nil, errors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}

	if err := s.uc.LogoutAllDevices(ctx, a.ID()); err != nil {
		return nil, err
	}
	return &authpb.LogoutAllDevicesResponse{Success: true}, nil
}
