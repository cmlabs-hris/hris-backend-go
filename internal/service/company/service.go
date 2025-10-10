package company

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type CompanyServiceImpl struct {
	db *database.DB
	company.CompanyRepository
}

// Create implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).Create of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) Create(ctx context.Context, req company.CreateCompanyRequest) (company.Company, error) {
	panic("unimplemented")
}

// Delete implements company.CompanyService.
func (c *CompanyServiceImpl) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

// GetByID implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).GetByID of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) GetByID(ctx context.Context, id string) (company.CompanyResponse, error) {
	panic("unimplemented")
}

// List implements company.CompanyService.
func (c *CompanyServiceImpl) List(ctx context.Context) ([]company.Company, error) {
	panic("unimplemented")
}

// Update implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).Update of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) Update(ctx context.Context, id string, req company.UpdateCompanyRequest) error {
	err := c.CompanyRepository.Update(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update company with id %s: %w", id, err)
	}
	return nil
}

func NewCompanyService(db *database.DB, companyRepository company.CompanyRepository) company.CompanyService {
	return &CompanyServiceImpl{
		db:                db,
		CompanyRepository: companyRepository,
	}
}
