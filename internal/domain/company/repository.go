package company

import "context"

type CompanyRepository interface {
	GetByID(ctx context.Context, id string) (Company, error)
	GetByUsername(ctx context.Context, username string) (Company, error)
	Create(ctx context.Context, newCompany Company) (Company, error)
	ExistsByIDOrUsername(ctx context.Context, id, username *string) (bool, error)
	Update(ctx context.Context, id string, req UpdateCompanyRequest) error
	Delete(ctx context.Context, id string) error
}
