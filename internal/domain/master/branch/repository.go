package branch

import "context"

type BranchRepository interface {
	Create(ctx context.Context, branch Branch) (Branch, error)
	GetByID(ctx context.Context, id string) (Branch, error)
	GetByCompanyID(ctx context.Context, companyID string) ([]Branch, error)
	Update(ctx context.Context, req UpdateBranchRequest) error
	Delete(ctx context.Context, id string) error
}
