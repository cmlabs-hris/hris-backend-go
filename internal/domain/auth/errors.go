package auth

import "errors"

var (
	ErrInvalidCredentials             = errors.New("invalid email or password")
	ErrInvalidEmployeeCodeCredentials = errors.New("invalid company, employee code, or password")
	ErrAccountLocked                  = errors.New("account is locked")
	ErrEmailNotVerified               = errors.New("email not verified")
	ErrInvalidToken                   = errors.New("invalid or expired token")
	ErrTokenExpired                   = errors.New("token has expired")
	ErrRefreshTokenRevoked            = errors.New("refresh token has been revoked")
	ErrCompanyNotFound                = errors.New("company not found")
	ErrUserNotFound                   = errors.New("user not found")
	ErrStateCookieNotFound            = errors.New("state cookie not found")
	ErrRefreshTokenCookieNotFound     = errors.New("refresh_token cookie not found")
	ErrStateMismatch                  = errors.New("state mismatch: value from cookie does not match value from URL parameter")
	ErrStateParamEmpty                = errors.New("state param is empty")
	ErrStateCookieEmpty               = errors.New("state cookie is empty")
	ErrCodeValueEmpty                 = errors.New("code value is empty")
	ErrEmailAlreadyExists             = errors.New("account with this email already exists")
	ErrGoogleAccessDeniedByUser       = errors.New("continue with google access denied by user")
	ErrRefreshTokenCookieEmpty        = errors.New("refresh token cookie is empty")

	// Password reset errors
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired  = errors.New("password reset token has expired")
	ErrPasswordResetTokenUsed     = errors.New("password reset token has already been used")

	// Email verification errors
	ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
	ErrEmailVerificationTokenExpired  = errors.New("email verification token has expired")
	ErrEmailAlreadyVerified           = errors.New("email is already verified")
)
