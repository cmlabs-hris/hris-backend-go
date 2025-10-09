package auth

import (
	"context"
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (TokenResponse, error)
	Login(ctx context.Context, req LoginRequest) (TokenResponse, error)
	LoginWithEmployeeCode(ctx context.Context, req LoginEmployeeCodeRequest) (TokenResponse, error)
	LoginWithGoogle(ctx context.Context, email string, id string) error
	OAuthCallbackGoogle(ctx context.Context) (TokenResponse, error)
	Logout(ctx context.Context) error
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (AccessTokenResponse, error)
	ForgotPassword(ctx context.Context, req RefreshTokenRequest) error
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
}
