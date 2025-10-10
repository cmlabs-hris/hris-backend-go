package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type companyRepositoryImpl struct {
	db *database.DB
}

// Update implements company.CompanyRepository.
func (c *companyRepositoryImpl) Update(ctx context.Context, id string, req company.UpdateCompanyRequest) error {
	q := GetQuerier(ctx, c.db)

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Address != nil {
		updates["address"] = *req.Address
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
		INSERT INTO companies (name, username)
		VALUES ($1, $2)
		RETURNING id, name, username, created_at, updated_at
	`

	var created company.Company
	err := q.QueryRow(ctx, query, newCompany.Name, newCompany.Username).
		Scan(&created.ID, &created.Name, &created.Username, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return company.Company{}, err
	}
	return created, nil
}

// GetByID implements company.CompanyRepositoryImpl.
func (c *companyRepositoryImpl) GetByID(ctx context.Context, id string) (company.Company, error) {
	q := GetQuerier(ctx, c.db)

	query := `
		SELECT id, name, username, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	var found company.Company
	err := q.QueryRow(ctx, query, id).
		Scan(&found.ID, &found.Name, &found.Username, &found.CreatedAt, &found.UpdatedAt)
	if err != nil {
		return company.Company{}, err
	}

	return found, nil
}

// GetByUsername implements company.CompanyRepositoryImpl.
func (c *companyRepositoryImpl) GetByUsername(ctx context.Context, username string) (company.Company, error) {
	q := GetQuerier(ctx, c.db)

	query := `
		SELECT id, name, username, created_at, updated_at
		FROM companies
		WHERE username = $1
	`

	var found company.Company
	err := q.QueryRow(ctx, query, username).
		Scan(&found.ID, &found.Name, &found.Username, &found.CreatedAt, &found.UpdatedAt)
	if err != nil {
		return company.Company{}, err
	}

	return found, nil
}
