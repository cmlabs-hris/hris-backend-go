package user

import (
	"context"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	Create(ctx context.Context, newUser User) (User, error)
	ExistsByIDOrEmail(ctx context.Context, id, email *string) (bool, error)
	LinkGoogleAccount(ctx context.Context, GoogleID string, email string) (User, error)
	LinkPasswordAccount(ctx context.Context, id string, password string) (User, error)
	UpdateRole(ctx context.Context, req UpdateUserRoleRequest) error
	Update(ctx context.Context, req UpdateUserRequest) error
	UpdateCompanyAndRole(ctx context.Context, userID, companyID, role string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	VerifyEmail(ctx context.Context, userID string) error
	GetByEmailVerificationToken(ctx context.Context, token string) (User, error)
}
