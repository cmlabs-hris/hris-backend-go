package company

import (
	"context"
)

type CompanyService interface {
	List(ctx context.Context) ([]Company, error)
	Create(ctx context.Context, req CreateCompanyRequest) (Company, error)
	GetByID(ctx context.Context, id string) (CompanyResponse, error)
	Update(ctx context.Context, id string, req UpdateCompanyRequest) error
	Delete(ctx context.Context, id string) error
}
