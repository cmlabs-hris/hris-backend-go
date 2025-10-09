package user

import "errors"

var (
	ErrUserNotFound                = errors.New("user not found")
	ErrUserEmailExists             = errors.New("email already registered")
	ErrInvalidEmailFormat          = errors.New("invalid email format")
	ErrInvalidPasswordLength       = errors.New("password must be at least 8 characters")
	ErrInvalidOAuthProvider        = errors.New("invalid oauth provider")
	ErrOAuthProviderIDExists       = errors.New("oauth provider id already registered")
	ErrEmailNotVerified            = errors.New("email not verified")
	ErrEmailVerificationTokenEmpty = errors.New("email verification token is empty")
	ErrAdminPrivilegeRequired      = errors.New("admin privilege required")
	ErrUpdatedAtBeforeCreatedAt    = errors.New("updated_at cannot be before created_at")
)
