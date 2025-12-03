package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type companyRepositoryImpl struct {
	db *database.DB
}

// Delete implements company.CompanyRepository.
func (c *companyRepositoryImpl) Delete(ctx context.Context, id string) error {
	q := GetQuerier(ctx, c.db)

	query := `
		UPDATE companies
		SET deleted_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := q.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete company with id %s: %w", id, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no company found with id %s, %w", id, pgx.ErrNoRows)
	}

	return nil
}

// Update implements company.CompanyRepository.
func (c *companyRepositoryImpl) Update(ctx context.Context, id string, req company.UpdateCompanyRequest) error {
	q := GetQuerier(ctx, c.db)

	updates := make(map[string]interface{})

	if req.Name != nil {
		if *req.Name == "" {
			updates["name"] = nil
		} else {
			updates["name"] = *req.Name
		}
	}
	if req.Address != nil {
		if *req.Address == "" {
			updates["address"] = nil
		} else {
			updates["address"] = *req.Address
		}
	}
	if req.LogoURL != nil {
		if *req.LogoURL == "" {
			updates["logo_url"] = nil
		} else {
			updates["logo_url"] = *req.LogoURL
		}
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for company update")
	}
	updates["updated_at"] = time.Now()

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	sql := "UPDATE companies SET " + strings.Join(setClauses, ", ") + fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, id)

	fmt.Println(sql)
	var updatedID string
	if err := q.QueryRow(ctx, sql+" RETURNING id", args...).Scan(&updatedID); err != nil {
		return fmt.Errorf("failed to update company with id %s: %w", id, err)
	}
	return nil
}

func NewCompanyRepository(db *database.DB) company.CompanyRepository {
	return &companyRepositoryImpl{db: db}
}

// ExistsByIDOrUsername implements company.CompanyRepositoryImpl.
func (c *companyRepositoryImpl) ExistsByIDOrUsername(ctx context.Context, id *string, username *string) (bool, error) {
	q := GetQuerier(ctx, c.db)

	var query string
	var arg interface{}

	switch {
	case id != nil:
		query = `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`
		arg = *id
	case username != nil:
		query = `SELECT EXISTS(SELECT 1 FROM companies WHERE username = $1)`
		arg = *username
	default:
		return false, nil
	}

	var exists bool
	err := q.QueryRow(ctx, query, arg).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Create implements company.CompanyRepositoryImpl.
func (c *companyRepositoryImpl) Create(ctx context.Context, newCompany company.Company) (company.Company, error) {
	q := GetQuerier(ctx, c.db)

	query := `
		INSERT INTO companies (name, username, address, logo_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, username, address, logo_url, created_at, updated_at
	`

	var created company.Company

	var addr interface{}
	if newCompany.Address == nil {
		addr = nil
	} else {
		addr = *newCompany.Address
	}

	var logo interface{}
	if newCompany.LogoURL == nil {
		logo = nil
	} else {
		logo = *newCompany.LogoURL
	}

	err := q.QueryRow(ctx, query, newCompany.Name, newCompany.Username, addr, logo).
		Scan(&created.ID, &created.Name, &created.Username, &created.Address, &created.LogoURL, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return company.Company{}, err
	}
	return created, nil
}

// GetByID implements company.CompanyRepositoryImpl.
func (c *companyRepositoryImpl) GetByID(ctx context.Context, id string) (company.Company, error) {
	q := GetQuerier(ctx, c.db)

	query := `
		SELECT id, name, username, address, logo_url, created_at, updated_at, deleted_at
		FROM companies
		WHERE id = $1
	`

	var found company.Company
	err := q.QueryRow(ctx, query, id).
		Scan(&found.ID, &found.Name, &found.Username, &found.Address, &found.LogoURL, &found.CreatedAt, &found.UpdatedAt, &found.DeletedAt)
	if err != nil {
		return company.Company{}, err
	}

	return found, nil
}

// GetByUsername implements company.CompanyRepositoryImpl.
func (c *companyRepositoryImpl) GetByUsername(ctx context.Context, username string) (company.Company, error) {
	q := GetQuerier(ctx, c.db)

	query := `
		SELECT id, name, username, address, created_at, updated_at
		FROM companies
		WHERE username = $1
	`

	var found company.Company
	err := q.QueryRow(ctx, query, username).
		Scan(&found.ID, &found.Name, &found.Username, &found.Address, &found.CreatedAt, &found.UpdatedAt)
	if err != nil {
		return company.Company{}, err
	}

	return found, nil
}
