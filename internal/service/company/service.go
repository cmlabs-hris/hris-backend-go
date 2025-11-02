package company

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/cmlabs-hris/hris-backend-go/internal/service/file"
	"github.com/jackc/pgx/v5"
)

type CompanyServiceImpl struct {
	db *database.DB
	company.CompanyRepository
	fileService file.FileService
}

// Create implements company.CompanyService.
// Subtle: this method shadows the method (CompanyRepository).Create of CompanyServiceImpl.CompanyRepository.
func (c *CompanyServiceImpl) Create(ctx context.Context, req company.CreateCompanyRequest) (company.Company, error) {
	var newCompany company.Company
	err := postgresql.WithTransaction(ctx, c.db, func(tx pgx.Tx) error {
		_, err := c.CompanyRepository.GetByUsername(ctx, req.Username)
		if err != nil {
			if err != pgx.ErrNoRows {
				return fmt.Errorf("failed to get company by username: %w", err)
			}
			return company.ErrCompanyUsernameExists
		}
		if req.File != nil && req.FileHeader != nil {
			if req.FileHeader.Size > 5<<20 {
				return company.ErrFileSizeExceeds
			}

			allowedExts := []string{".pdf", ".jpg", ".jpeg", ".png"}
			ext := strings.ToLower(filepath.Ext(req.FileHeader.Filename))

			isValidExt := false
			for _, allowed := range allowedExts {
				if ext == allowed {
					isValidExt = true
					break
				}
			}

			if !isValidExt {
				return leave.ErrFileTypeNotAllowed
			}

			attachmentURL, err := c.fileService.UploadCompanyLogo(ctx, newCompany.Username, req.File, req.FileHeader.Filename)
			if err != nil {
				return fmt.Errorf("failed to upload leave attachment: %w", err)
			}
			req.AttachmentURL = &attachmentURL
		}
		newCompany, err = c.CompanyRepository.Create(ctx, company.Company{
			Name:     req.Name,
			Username: req.Username,
			Address:  req.Address,
			LogoURL:  req.AttachmentURL,
		})
		if err != nil {
			return fmt.Errorf("failed to create company: %w", err)
		}

		return nil
	})
	// TODO ADD PREMADE MASTER SEEDING DATA
	if err != nil {
		return company.Company{}, err
	}

	return newCompany, nil
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

func NewCompanyService(db *database.DB, companyRepository company.CompanyRepository, fileService file.FileService) company.CompanyService {
	return &CompanyServiceImpl{
		db:                db,
		CompanyRepository: companyRepository,
		fileService:       fileService,
	}
}
