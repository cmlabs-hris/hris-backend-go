package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountLocked       = errors.New("account is locked")
	ErrEmailNotVerified    = errors.New("email not verified")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrTokenExpired        = errors.New("token has expired")
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
	ErrCompanyNotFound     = errors.New("company not found")
	ErrUserNotFound        = errors.New("user not found")
)
