package auth

import (
	"context"
)

type AuthService interface {
	Register(ctx context.Context, registerReq RegisterRequest, sessionReq SessionTrackingRequest) (TokenResponse, error)
	Login(ctx context.Context, loginReq LoginRequest, sessionReq SessionTrackingRequest) (TokenResponse, error)
	LoginWithEmployeeCode(ctx context.Context, req LoginEmployeeCodeRequest, sessionReq SessionTrackingRequest) (TokenResponse, error)
	LoginWithGoogle(ctx context.Context, googleEmail string, googleID string, sessionReq SessionTrackingRequest) (TokenResponse, error)
	OAuthCallbackGoogle(ctx context.Context) (TokenResponse, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (AccessTokenResponse, error)
	ForgotPassword(ctx context.Context, req RefreshTokenRequest) error
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
}
