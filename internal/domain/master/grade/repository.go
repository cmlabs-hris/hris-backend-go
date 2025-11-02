package grade

import "context"

type GradeRepository interface {
	Create(ctx context.Context, grade Grade) (Grade, error)
	GetByID(ctx context.Context, id string) (Grade, error)
	GetByCompanyID(ctx context.Context, companyID string) ([]Grade, error)
	Update(ctx context.Context, req UpdateGradeRequest) error
	Delete(ctx context.Context, id string) error
}
