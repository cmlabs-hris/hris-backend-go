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
	ErrPendingRoleRequired         = errors.New("pending role required")
	ErrOwnerAccessRequired         = errors.New("owner access required")
	ErrManagerAccessRequired       = errors.New("manager access required")
	ErrPendingRoleAccessRequired   = errors.New("pending role access required")
	ErrInsufficientPermissions     = errors.New("insufficient permissions")
	ErrCompanyIDRequired           = errors.New("company ID is required")
	ErrUpdatedAtBeforeCreatedAt    = errors.New("updated_at cannot be before created_at")
)
