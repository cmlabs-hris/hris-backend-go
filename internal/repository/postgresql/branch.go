package postgresql

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type branchRepositoryImpl struct {
	db *database.DB
}

func NewBranchRepository(db *database.DB) branch.BranchRepository {
	return &branchRepositoryImpl{db: db}
}

// GetTimezoneByEmployeeID implements branch.BranchRepository.
func (r *branchRepositoryImpl) GetTimezoneByEmployeeID(ctx context.Context, employeeID string, companyID string) (string, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT b.timezone
		FROM branches b
		JOIN employees e ON b.id = e.branch_id
		WHERE e.id = $1 AND e.company_id = $2
	`

	var timezone string
	err := q.QueryRow(ctx, query, employeeID, companyID).Scan(&timezone)

	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("employee or branch not found: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get timezone by employee: %w", err)
	}

	return timezone, nil
}

// GetTimezone implements branch.BranchRepository.
func (r *branchRepositoryImpl) GetTimezone(ctx context.Context, id string, companyID string) (string, error) {
	q := GetQuerier(ctx, r.db)

	query := `
			SELECT timezone
			FROM branches
			WHERE id = $1 AND company_id = $2
		`

	var timezone string
	err := q.QueryRow(ctx, query, id, companyID).Scan(&timezone)

	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("branch not found: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get branch timezone: %w", err)
	}

	return timezone, nil
}

// Create implements branch.BranchRepository.
func (r *branchRepositoryImpl) Create(ctx context.Context, b branch.Branch) (branch.Branch, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO branches (id, company_id, name, address, timezone, created_at, updated_at)
		VALUES (uuidv7(), $1, $2, $3, COALESCE($4, 'Asia/Jakarta'), NOW(), NOW())
		RETURNING id, company_id, name, address, timezone
	`

	b.Timezone = "Asia/Jakarta"
	var result branch.Branch
	err := q.QueryRow(ctx, query, b.CompanyID, b.Name, b.Address, b.Timezone).Scan(
		&result.ID,
		&result.CompanyID,
		&result.Name,
		&result.Address,
		&result.Timezone,
	)

	if err != nil {
		return branch.Branch{}, fmt.Errorf("failed to create branch: %w", err)
	}

	return result, nil
}

// GetByID implements branch.BranchRepository.
func (r *branchRepositoryImpl) GetByID(ctx context.Context, id string, companyID string) (branch.Branch, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name, address, timezone
		FROM branches
		WHERE id = $1 AND company_id = $2
	`

	var result branch.Branch
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&result.ID,
		&result.CompanyID,
		&result.Name,
		&result.Address,
		&result.Timezone,
	)

	if err == pgx.ErrNoRows {
		return branch.Branch{}, fmt.Errorf("branch not found: %w", err)
	}

	if err != nil {
		return branch.Branch{}, fmt.Errorf("failed to get branch: %w", err)
	}

	return result, nil
}

// GetByCompanyID implements branch.BranchRepository.
func (r *branchRepositoryImpl) GetByCompanyID(ctx context.Context, companyID string) ([]branch.Branch, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name, address, timezone
		FROM branches
		WHERE company_id = $1
		ORDER BY name ASC
	`

	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}
	defer rows.Close()

	var branches []branch.Branch
	for rows.Next() {
		var b branch.Branch
		err := rows.Scan(
			&b.ID,
			&b.CompanyID,
			&b.Name,
			&b.Address,
			&b.Timezone,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		branches = append(branches, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return branches, nil
}

// Update implements branch.BranchRepository.
func (r *branchRepositoryImpl) Update(ctx context.Context, req branch.UpdateBranchRequest) error {
	q := GetQuerier(ctx, r.db)

	// Build dynamic update query
	query := `UPDATE branches SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *req.Name)
		argIdx++
	}

	if req.Address != nil {
		query += fmt.Sprintf(", address = $%d", argIdx)
		args = append(args, *req.Address)
		argIdx++
	}

	if req.Timezone != nil {
		query += fmt.Sprintf(", timezone = $%d", argIdx)
		args = append(args, *req.Timezone)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND company_id = $%d", argIdx, argIdx+1)
	args = append(args, req.ID, req.CompanyID)

	commandTag, err := q.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("branch not found: %w", pgx.ErrNoRows)
	}

	return nil
}

// Delete implements branch.BranchRepository.
func (r *branchRepositoryImpl) Delete(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, r.db)

	query := `DELETE FROM branches WHERE id = $1 AND company_id = $2`

	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("branch not found: %w", pgx.ErrNoRows)
	}

	return nil
}
