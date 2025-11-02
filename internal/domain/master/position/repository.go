package position

import "context"

type PositionRepository interface {
	Create(ctx context.Context, position Position) (Position, error)
	GetByID(ctx context.Context, id string) (Position, error)
	GetByCompanyID(ctx context.Context, companyID string) ([]Position, error)
	Update(ctx context.Context, req UpdatePositionRequest) error
	Delete(ctx context.Context, id string) error
}
