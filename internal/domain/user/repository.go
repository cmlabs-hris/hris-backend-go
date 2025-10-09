package user

import (
	"context"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	Create(ctx context.Context, newUser User) (User, error)
	ExistsByIDOrEmail(ctx context.Context, id, email *string) (bool, error)
}
